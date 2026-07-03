package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListRegistries(ctx context.Context) ([]registryResponse, int, error) {
	var out []registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registries", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetRegistry(ctx context.Context, id string) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registries/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateRegistry(ctx context.Context, payload map[string]any) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/registries", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateRegistry(ctx context.Context, id string, payload map[string]any) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/registries/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteRegistry(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/registries/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) SearchRegistry(ctx context.Context, term string, registry string) ([]map[string]any, int, error) {
	query := map[string]string{
		"term": strings.TrimSpace(term),
	}
	if v := strings.TrimSpace(registry); v != "" {
		query["registry"] = v
	}

	var out []map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registry/search", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out, status, nil
}

func (c *Client) ListRegistryTags(ctx context.Context, image string, registry string, page int64, pageSize int64) (*registryTagsResponse, int, error) {
	query := map[string]string{
		"image": strings.TrimSpace(image),
	}
	if v := strings.TrimSpace(registry); v != "" {
		query["registry"] = v
	}
	if page > 0 {
		query["page"] = strconv.FormatInt(page, 10)
	}
	if pageSize > 0 {
		query["pageSize"] = strconv.FormatInt(pageSize, 10)
	}

	var out registryTagsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registry/tags", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out.Tags == nil {
		out.Tags = []registryTagResponse{}
	}
	return &out, status, nil
}

func (c *Client) GetRegistryCatalogRaw(ctx context.Context, registry string, page int64, pageSize int64) (json.RawMessage, int, error) {
	query := map[string]string{}
	if v := strings.TrimSpace(registry); v != "" {
		query["registry"] = v
	}
	if page > 0 {
		query["page"] = strconv.FormatInt(page, 10)
	}
	if pageSize > 0 {
		query["pageSize"] = strconv.FormatInt(pageSize, 10)
	}

	var raw json.RawMessage
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registry/catalog", query, nil, &raw)
	if err != nil {
		return nil, status, err
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	return raw, status, nil
}

func (c *Client) DeleteRegistryImage(ctx context.Context, registry string, image string, tag string) (map[string]any, int, error) {
	query := map[string]string{
		"registry": strings.TrimSpace(registry),
		"image":    strings.TrimSpace(image),
		"tag":      strings.TrimSpace(tag),
	}

	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodDelete, "/api/registry/image", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, status, nil
}

func (c *Client) Prune(ctx context.Context, env string, mode string) (map[string]any, int, error) {
	pruneMode := strings.ToLower(strings.TrimSpace(mode))
	path := ""
	switch pruneMode {
	case "all":
		path = "/api/prune/all"
	case "containers":
		path = "/api/prune/containers"
	case "images":
		path = "/api/prune/images"
	case "networks":
		path = "/api/prune/networks"
	case "volumes":
		path = "/api/prune/volumes"
	default:
		return nil, 0, fmt.Errorf("unsupported prune mode %q", mode)
	}

	query := map[string]string{}
	if resolved := strings.TrimSpace(c.resolveEnv(env)); resolved != "" {
		query["env"] = resolved
	}

	var out map[string]any
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, path, query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, status, nil
}
