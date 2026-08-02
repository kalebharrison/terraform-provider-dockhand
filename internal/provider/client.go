package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewClient(endpoint string, sessionCookie string, authHeader string, defaultEnv string, insecure bool) (*Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure,
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		sessionCookie:        sessionCookie,
		authHeader:           authHeader,
		defaultEnv:           defaultEnv,
		requestRetryAttempts: defaultRequestRetryAttempts,
		requestRetryMinDelay: defaultRequestRetryMinDelay,
		requestRetryMaxDelay: defaultRequestRetryMaxDelay,
	}, nil
}

func (c *Client) SetRequestRetryPolicy(retry requestRetryConfig) {
	if c == nil {
		return
	}
	if retry.attempts > 0 {
		c.requestRetryAttempts = retry.attempts
	}
	if retry.minDelay > 0 {
		c.requestRetryMinDelay = retry.minDelay
	}
	if retry.maxDelay > 0 {
		c.requestRetryMaxDelay = retry.maxDelay
	}
}

func (c *Client) requestRetrySleep(ctx context.Context, attempt int) error {
	retry := requestRetryConfig{
		attempts: c.requestRetryAttempts,
		minDelay: c.requestRetryMinDelay,
		maxDelay: c.requestRetryMaxDelay,
	}
	return retry.sleep(ctx, attempt)
}

func (c *Client) ListActivity(ctx context.Context) ([]activityEventResponse, int, error) {
	var out activityResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/activity", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out.Events, status, nil
}

func (c *Client) GetHawserStatus(ctx context.Context) (*hawserConnectStatus, int, error) {
	var out hawserConnectStatus
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/hawser/connect", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) Health(ctx context.Context, env string) (*healthResponse, error) {
	// Dockhand docs do not expose a dedicated health endpoint.
	// We treat a successful dashboard stats request as API health.
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	if _, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/dashboard/stats", query, nil, nil); err != nil {
		return nil, err
	}

	return &healthResponse{Status: "ok"}, nil
}

func (c *Client) doJSONWithStatus(ctx context.Context, method string, path string, query map[string]string, in any, out any) (int, error) {
	return c.doJSONWithStatusUsingClient(ctx, c.httpClient, method, path, query, nil, in, out)
}

func (c *Client) doJSONWithStatusHeaders(ctx context.Context, method string, path string, query map[string]string, headers map[string]string, in any, out any) (int, error) {
	return c.doJSONWithStatusUsingClient(ctx, c.httpClient, method, path, query, headers, in, out)
}

func (c *Client) doJSONWithStatusUsingClient(ctx context.Context, httpClient *http.Client, method string, path string, query map[string]string, headers map[string]string, in any, out any) (int, error) {
	var payloadBytes []byte
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return 0, err
		}
		payloadBytes = data
	}

	// Build the URL once; the request itself may be retried.
	ref := &url.URL{Path: path}
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			if v != "" {
				values.Set(k, v)
			}
		}
		ref.RawQuery = values.Encode()
	}
	fullURL := c.baseURL.ResolveReference(ref).String()

	var lastStatus int
	var responseBody []byte
	maxAttempts := c.requestRetryAttempts
	if maxAttempts < 1 {
		maxAttempts = defaultRequestRetryAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var body io.Reader
		if payloadBytes != nil {
			body = bytes.NewReader(payloadBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
		if err != nil {
			return 0, err
		}

		req.Header.Set("Accept", "application/json")
		if payloadBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for key, value := range headers {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			req.Header.Set(key, value)
		}
		c.applyAuthHeaders(req)

		//nolint:gosec // Provider endpoint is an explicit user-configured Dockhand API target.
		res, err := httpClient.Do(req)
		if err != nil {
			if shouldRetryRequest(method, 0, err) && attempt < maxAttempts-1 {
				if sleepErr := c.requestRetrySleep(ctx, attempt); sleepErr != nil {
					return 0, err
				}
				continue
			}
			return 0, err
		}

		lastStatus = res.StatusCode

		// On errors, keep the body very small to avoid huge allocations in diagnostics.
		limit := int64(10 << 20) // 10 MiB
		if res.StatusCode < 200 || res.StatusCode > 299 {
			limit = 64 << 10 // 64 KiB
		}

		responseBody, err = io.ReadAll(io.LimitReader(res.Body, limit))
		res.Body.Close()
		if err != nil {
			if shouldRetryRequest(method, lastStatus, err) && attempt < maxAttempts-1 {
				if sleepErr := c.requestRetrySleep(ctx, attempt); sleepErr != nil {
					return lastStatus, err
				}
				continue
			}
			return lastStatus, err
		}

		if shouldRetryRequest(method, lastStatus, nil) && attempt < maxAttempts-1 {
			if sleepErr := c.requestRetrySleep(ctx, attempt); sleepErr != nil {
				break
			}
			continue
		}

		break
	}

	if lastStatus < 200 || lastStatus > 299 {
		if len(responseBody) == 0 {
			return lastStatus, fmt.Errorf("dockhand api returned status %d", lastStatus)
		}
		return lastStatus, fmt.Errorf("dockhand api returned status %d: %s", lastStatus, strings.TrimSpace(string(responseBody)))
	}

	if out != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, out); err != nil {
			return lastStatus, err
		}
	}

	return lastStatus, nil
}

func (c *Client) httpClientWithTimeout(timeout time.Duration) *http.Client {
	if c == nil || c.httpClient == nil {
		return &http.Client{Timeout: timeout}
	}
	clone := *c.httpClient
	clone.Timeout = timeout
	return &clone
}

func (c *Client) resolveEnv(value string) string {
	if value != "" {
		return value
	}
	return c.defaultEnv
}

func (c *Client) applyAuthHeaders(req *http.Request) {
	if c.sessionCookie != "" {
		req.Header.Set("Cookie", c.sessionCookie)
	}
	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}
}

func sleepBackoffWithInterval(ctx context.Context, delay time.Duration) error {
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func imagePullStreamError(body []byte) string {
	return apiStreamError(body)
}

func apiStreamError(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	lines := bytes.Split(body, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal(line, &obj); err != nil {
			continue
		}

		if status, ok := obj["status"].(string); ok && strings.EqualFold(strings.TrimSpace(status), "error") {
			if msg := imagePullErrorMessage(obj); msg != "" {
				return msg
			}
			return "unknown pull error"
		}

		if msg := imagePullErrorMessage(obj); msg != "" {
			return msg
		}
	}

	return ""
}

func imagePullErrorMessage(obj map[string]any) string {
	if obj == nil {
		return ""
	}

	if msg, ok := obj["error"].(string); ok && strings.TrimSpace(msg) != "" {
		return strings.TrimSpace(msg)
	}

	if detail, ok := obj["errorDetail"].(map[string]any); ok {
		if msg, ok := detail["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return strings.TrimSpace(msg)
		}
	}

	return ""
}

func firstString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
	}

	return ""
}

func firstMap(item map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		if m, ok := value.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func extractBatchJobID(payload map[string]any) string {
	if payload == nil {
		return ""
	}

	if id := strings.TrimSpace(firstString(payload, "jobId", "jobID", "job_id")); id != "" {
		return id
	}

	for _, key := range []string{"job", "data", "result"} {
		if m := firstMap(payload, key); m != nil {
			if id := strings.TrimSpace(firstString(m, "jobId", "jobID", "job_id", "id")); id != "" {
				return id
			}
		}
	}

	return strings.TrimSpace(firstString(payload, "id"))
}

func parseJobResponse(payload map[string]any) jobResponse {
	source := payload
	for _, key := range []string{"job", "data"} {
		if m := firstMap(payload, key); m != nil {
			source = m
			break
		}
	}

	out := jobResponse{
		ID:     strings.TrimSpace(firstString(source, "id", "jobId", "jobID", "job_id")),
		Status: strings.TrimSpace(firstString(source, "status", "state")),
		Result: mapFromAny(source["result"]),
		Lines:  parseJobLines(source["lines"]),
	}

	if out.Result == nil {
		out.Result = map[string]any{}
	}
	if out.Lines == nil {
		out.Lines = []jobLineResponse{}
	}

	return out
}

func mapFromAny(value any) map[string]any {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case map[string]any:
		return v
	case map[string]string:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = val
		}
		return out
	default:
		return nil
	}
}

func toStringSlice(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func firstInt64(item map[string]any, keys ...string) int64 {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case json.Number:
			parsed, err := v.Int64()
			if err == nil {
				return parsed
			}
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}
