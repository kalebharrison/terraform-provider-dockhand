package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) ListGitCredentials(ctx context.Context) ([]gitCredentialResponse, int, error) {
	var out []gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/credentials", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetGitCredential(ctx context.Context, id string) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/credentials/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateGitCredential(ctx context.Context, payload gitCredentialPayload) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/credentials", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGitCredential(ctx context.Context, id string, payload gitCredentialPayload) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/git/credentials/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteGitCredential(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/git/credentials/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListGitRepositories(ctx context.Context) ([]gitRepositoryResponse, int, error) {
	var out []gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/repositories", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetGitRepository(ctx context.Context, id string) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/repositories/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateGitRepository(ctx context.Context, payload gitRepositoryPayload) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/repositories", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGitRepository(ctx context.Context, id string, payload gitRepositoryPayload) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/git/repositories/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteGitRepository(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/git/repositories/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) TestGitRepository(ctx context.Context, payload gitRepositoryTestPayload) (*gitRepositoryTestResponse, int, error) {
	var out gitRepositoryTestResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/repositories/test", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) PreviewGitEnv(ctx context.Context, payload gitPreviewEnvPayload) (*gitPreviewEnvResponse, int, error) {
	var out gitPreviewEnvResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/preview-env", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	if out.Vars == nil {
		out.Vars = map[string]any{}
	}
	if out.Sources == nil {
		out.Sources = map[string]any{}
	}
	return &out, status, nil
}

func (c *Client) ListGitStacks(ctx context.Context, env string) ([]gitStackResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []gitStackResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/stacks", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetGitStackByID(ctx context.Context, env string, id string) (*gitStackResponse, int, error) {
	items, status, err := c.ListGitStacks(ctx, env)
	if err != nil {
		return nil, status, err
	}
	for i := range items {
		if fmt.Sprintf("%d", items[i].ID) == strings.TrimSpace(id) {
			return &items[i], status, nil
		}
	}
	return nil, status, nil
}

func (c *Client) CreateGitStack(ctx context.Context, env string, payload gitStackPayload) (*gitStackResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out gitStackResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/stacks", query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGitStack(ctx context.Context, env string, id string, payload gitStackPayload) (*gitStackResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out gitStackResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/git/stacks/"+url.PathEscape(id), query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteGitStack(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/git/stacks/"+url.PathEscape(id), query, nil, nil)
}

func (c *Client) TriggerGitStackWebhook(ctx context.Context, id string, secret string) (int, error) {
	body := map[string]any{}
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	query := map[string]string{}
	headers := map[string]string{}
	secret = strings.TrimSpace(secret)
	if secret != "" {
		query["secret"] = secret
		headers["X-Gitlab-Token"] = secret
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(payload)
		headers["X-Hub-Signature-256"] = "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}

	// json.RawMessage keeps HMAC bytes identical to the request body Dockhand verifies.
	return c.doJSONWithStatusHeaders(ctx, http.MethodPost, "/api/git/stacks/"+url.PathEscape(id)+"/webhook", query, headers, json.RawMessage(payload), nil)
}

func (c *Client) ListGitStackEnvFiles(ctx context.Context, id string) ([]string, int, error) {
	var out struct {
		Files []string `json:"files"`
	}
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/stacks/"+url.PathEscape(id)+"/env-files", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out.Files, status, nil
}

func (c *Client) GetGitStackEnvFileVars(ctx context.Context, id string, path string) (map[string]string, int, error) {
	payload := map[string]string{
		"path": path,
	}
	var out struct {
		Vars map[string]string `json:"vars"`
	}
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/stacks/"+url.PathEscape(id)+"/env-files", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return out.Vars, status, nil
}

func (c *Client) DeployGitStack(ctx context.Context, id string) (int, string, error) {
	endpoint, err := c.baseURL.Parse("/api/git/stacks/" + url.PathEscape(id) + "/deploy-stream")
	if err != nil {
		return 0, "", fmt.Errorf("compose deploy URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return 0, "", err
	}
	c.applyAuthHeaders(req)

	res, err := c.httpClientWithTimeout(5 * time.Minute).Do(req)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 1024*1024))
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		if msg := apiStreamError(body); msg != "" {
			return res.StatusCode, strings.TrimSpace(string(body)), fmt.Errorf("dockhand git deploy reported error: %s", msg)
		}
	}
	return res.StatusCode, strings.TrimSpace(string(body)), nil
}
