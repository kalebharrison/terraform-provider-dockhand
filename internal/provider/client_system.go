package provider

import (
	"context"
	"net/http"
	"strings"
)

func (c *Client) GetSystemInfo(ctx context.Context) (map[string]any, int, error) {
	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/system", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, status, nil
}

func (c *Client) GetSystemDisk(ctx context.Context, env string) (map[string]any, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/system/disk", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, status, nil
}

func (c *Client) ListSystemFiles(ctx context.Context, path string) (*systemFilesResponse, int, error) {
	query := map[string]string{
		"path": strings.TrimSpace(path),
	}

	var out systemFilesResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/system/files", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetSystemFileContent(ctx context.Context, path string) (*systemFileContentResponse, int, error) {
	query := map[string]string{
		"path": strings.TrimSpace(path),
	}

	var out systemFileContentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/system/files/content", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}
