package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) ListImages(ctx context.Context, env string) ([]imageResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []imageResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/images", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) PullImage(ctx context.Context, env string, image string, scanAfterPull bool) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := imagePullPayload{
		Image:         image,
		ScanAfterPull: scanAfterPull,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	ref := &url.URL{Path: "/api/images/pull"}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	c.applyAuthHeaders(req)

	res, err := c.httpClientWithTimeout(5 * time.Minute).Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 10<<20)) // 10 MiB max stream capture
	if err != nil {
		return res.StatusCode, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		if len(body) == 0 {
			return res.StatusCode, fmt.Errorf("dockhand api returned status %d", res.StatusCode)
		}
		return res.StatusCode, fmt.Errorf("dockhand api returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	if msg := imagePullStreamError(body); msg != "" {
		return res.StatusCode, fmt.Errorf("dockhand image pull reported error: %s", msg)
	}

	return res.StatusCode, nil
}

func (c *Client) DeleteImage(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{
		"force": "true",
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var (
		status int
		err    error
	)
	maxAttempts := c.requestRetryAttempts
	if maxAttempts < 1 {
		maxAttempts = defaultRequestRetryAttempts
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		status, err = c.doJSONWithStatus(ctx, http.MethodDelete, "/api/images/"+url.PathEscape(id), query, nil, nil)
		if !isImageInUseConflict(status, err) {
			return status, err
		}
		if attempt == maxAttempts-1 {
			break
		}
		if sleepErr := c.requestRetrySleep(ctx, attempt); sleepErr != nil {
			return status, sleepErr
		}
	}
	return status, err
}

func isImageInUseConflict(status int, err error) bool {
	if status != http.StatusConflict || err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cannot delete image") && strings.Contains(msg, "used by a running container")
}

func (c *Client) PushImage(ctx context.Context, env string, imageID string, registryID int64) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := imagePushPayload{
		ImageID:    imageID,
		RegistryID: registryID,
	}
	pushClient := c.httpClientWithTimeout(5 * time.Minute)
	return c.doJSONWithStatusUsingClient(ctx, pushClient, http.MethodPost, "/api/images/push", query, nil, payload, nil)
}

func (c *Client) ScanImage(ctx context.Context, env string, imageName string) (string, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	scanClient := *c.httpClient
	scanClient.Timeout = 3 * time.Minute

	status, err := c.doJSONWithStatusUsingClient(ctx, &scanClient, http.MethodPost, "/api/images/scan", query, nil, imageScanPayload{ImageName: imageName}, nil)
	if err != nil {
		return "", status, err
	}
	// Endpoint streams scan progress; if request succeeded we return a generic completion marker.
	return "scan_requested", status, nil
}
