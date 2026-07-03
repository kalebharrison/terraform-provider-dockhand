package provider

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func (c *Client) ListContainers(ctx context.Context, env string) ([]containerResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []containerResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetContainerByID(ctx context.Context, env string, id string) (*containerResponse, bool, error) {
	containers, _, err := c.ListContainers(ctx, env)
	if err != nil {
		return nil, false, err
	}
	for i := range containers {
		if containers[i].ID == id {
			return &containers[i], true, nil
		}
	}
	return nil, false, nil
}

func (c *Client) CreateContainer(ctx context.Context, env string, payload containerPayload) (*containerCreateResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out containerCreateResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers", query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) StartContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/start", query, nil, nil)
}

func (c *Client) StopContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/stop", query, nil, nil)
}

func (c *Client) RestartContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/restart", query, nil, nil)
}

func (c *Client) RenameContainer(ctx context.Context, env string, id string, name string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := map[string]string{
		"name": name,
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/rename", query, payload, nil)
}

func (c *Client) UpdateContainer(ctx context.Context, env string, id string, payload map[string]any) (map[string]any, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/update", query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) PauseContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/pause", query, nil, nil)
}

func (c *Client) UnpauseContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/unpause", query, nil, nil)
}

func (c *Client) GetContainerLogs(ctx context.Context, env string, id string, tail int64) (*containerLogsResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	if tail > 0 {
		query["tail"] = strconv.FormatInt(tail, 10)
	}

	var out containerLogsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id)+"/logs", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetContainerTop(ctx context.Context, env string, id string) (*containerTopResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out containerTopResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id)+"/top", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetContainerFileContent(ctx context.Context, env string, id string, path string) (string, int, error) {
	query := map[string]string{
		"path": path,
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out struct {
		Content string  `json:"content"`
		Error   *string `json:"error"`
	}
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id)+"/files/content", query, nil, &out)
	if err != nil {
		return "", status, err
	}
	return out.Content, status, nil
}

func (c *Client) CreateContainerFile(ctx context.Context, env string, id string, path string, fileType string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	payload := map[string]any{
		"path": path,
		"type": fileType,
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/"+url.PathEscape(id)+"/files/create", query, payload, nil)
}

func (c *Client) UpdateContainerFileContent(ctx context.Context, env string, id string, path string, content string) (int, error) {
	query := map[string]string{
		"path": path,
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	payload := map[string]any{
		"content": content,
	}
	return c.doJSONWithStatus(ctx, http.MethodPut, "/api/containers/"+url.PathEscape(id)+"/files/content", query, payload, nil)
}

func (c *Client) DeleteContainerFile(ctx context.Context, env string, id string, path string) (int, error) {
	query := map[string]string{
		"path": path,
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/containers/"+url.PathEscape(id)+"/files/delete", query, nil, nil)
}

func (c *Client) GetContainerShells(ctx context.Context, env string, id string) (*containerShellsResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		// Dockhand terminal APIs use `envId` query key instead of `env`.
		query["envId"] = resolvedEnv
	}

	var out containerShellsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id)+"/shells", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetContainerInspect(ctx context.Context, env string, id string) (map[string]any, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/"+url.PathEscape(id), query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetContainerStats(ctx context.Context, env string) ([]containerStatsResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []containerStatsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/stats", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) CheckContainerUpdates(ctx context.Context, env string) (*containerUpdateCheckResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out containerUpdateCheckResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/containers/check-updates", query, map[string]any{}, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetContainerPendingUpdates(ctx context.Context, env string) (*containerPendingUpdatesResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out containerPendingUpdatesResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/containers/pending-updates", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteContainer(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	query["force"] = "true"

	var (
		status int
		err    error
	)
	for i := range 5 {
		status, err = c.doJSONWithStatus(ctx, http.MethodDelete, "/api/containers/"+url.PathEscape(id), query, nil, nil)
		if err == nil || status == http.StatusNotFound {
			return status, err
		}
		if status < 500 {
			return status, err
		}
		if i == 4 {
			break
		}
		select {
		case <-ctx.Done():
			return status, ctx.Err()
		case <-time.After(1200 * time.Millisecond):
		}
	}

	return status, err
}
