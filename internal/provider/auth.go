package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultLoginRetryAttempts = 6

type loginResponse struct {
	Success     bool   `json:"success"`
	RequiresMFA bool   `json:"requiresMfa"`
	Error       string `json:"error"`
}

// Login authenticates with Dockhand and returns a Cookie header value like "dockhand_session=...".
func Login(ctx context.Context, endpoint string, username string, password string, mfaToken string, provider string, insecure bool) (string, error) {
	return loginWithRetry(ctx, endpoint, username, password, mfaToken, provider, insecure, defaultLoginRetryAttempts)
}

func loginWithRetry(ctx context.Context, endpoint string, username string, password string, mfaToken string, provider string, insecure bool, attempts int) (string, error) {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		cookie, err := loginOnce(ctx, endpoint, username, password, mfaToken, provider, insecure)
		if err == nil {
			return cookie, nil
		}
		lastErr = err
		if !isLoginRetryable(err) || attempt == attempts-1 {
			return "", err
		}
		if sleepErr := sleepLoginBackoff(ctx, attempt); sleepErr != nil {
			return "", sleepErr
		}
	}
	return "", lastErr
}

func loginOnce(ctx context.Context, endpoint string, username string, password string, mfaToken string, provider string, insecure bool) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required for login-based auth")
	}
	if provider == "" {
		provider = "local"
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure,
		},
	}
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	body := map[string]any{
		"username": username,
		"password": password,
		"provider": provider,
	}
	if mfaToken != "" {
		body["mfaToken"] = mfaToken
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	fullURL := baseURL.ResolveReference(&url.URL{Path: "/api/auth/login"}).String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", loginHTTPError(res.StatusCode, b)
	}

	for _, c := range res.Cookies() {
		if c.Name == "dockhand_session" && c.Value != "" {
			return fmt.Sprintf("%s=%s", c.Name, c.Value), nil
		}
	}

	// Fallback: parse Set-Cookie header manually (in case Go doesn't surface it as a Cookie).
	for _, h := range res.Header.Values("Set-Cookie") {
		if strings.HasPrefix(h, "dockhand_session=") {
			parts := strings.SplitN(h, ";", 2)
			return strings.TrimSpace(parts[0]), nil
		}
	}

	return "", fmt.Errorf("dockhand login succeeded but no dockhand_session cookie was returned")
}

func loginHTTPError(status int, body []byte) error {
	var lr loginResponse
	if err := json.Unmarshal(body, &lr); err == nil && lr.Error != "" {
		return fmt.Errorf("dockhand login failed: %s", lr.Error)
	}
	return fmt.Errorf("dockhand login failed (status %d): %s", status, strings.TrimSpace(string(body)))
}

func isLoginHTTPRetryable(status int) bool {
	switch status {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isLoginRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "no route to host"),
		strings.Contains(msg, "eof"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "temporarily unavailable"):
		return true
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}

	return isLoginHTTPRetryable(extractLoginHTTPStatus(err))
}

func extractLoginHTTPStatus(err error) int {
	msg := err.Error()
	if !strings.Contains(msg, "status ") {
		return 0
	}
	var status int
	if _, scanErr := fmt.Sscanf(msg, "dockhand login failed (status %d)", &status); scanErr == nil {
		return status
	}
	return 0
}

func sleepLoginBackoff(ctx context.Context, attempt int) error {
	delays := []time.Duration{
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}
	delay := delays[len(delays)-1]
	if attempt >= 0 && attempt < len(delays) {
		delay = delays[attempt]
	}

	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
