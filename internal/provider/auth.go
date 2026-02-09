package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type loginResponse struct {
	Success     bool   `json:"success"`
	RequiresMFA bool   `json:"requiresMfa"`
	Error       string `json:"error"`
}

// Login authenticates with Dockhand and returns a Cookie header value like "dockhand_session=...".
func Login(ctx context.Context, endpoint string, username string, password string, mfaToken string, provider string, insecure bool) (string, error) {
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
		var lr loginResponse
		if err := json.Unmarshal(b, &lr); err == nil && lr.Error != "" {
			return "", fmt.Errorf("dockhand login failed: %s", lr.Error)
		}
		return "", fmt.Errorf("dockhand login failed (status %d): %s", res.StatusCode, strings.TrimSpace(string(b)))
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
