package provider

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) ListNetworks(ctx context.Context, env string) ([]networkResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []networkResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/networks", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetNetworkInspect(ctx context.Context, env string, id string) (*networkInspectResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out networkInspectResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/networks/"+url.PathEscape(id)+"/inspect", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateNetwork(ctx context.Context, env string, payload networkPayload) (*networkResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out networkResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/networks", query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteNetwork(ctx context.Context, env string, id string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/networks/"+url.PathEscape(id), query, nil, nil)
}

func (c *Client) ConnectNetwork(ctx context.Context, env string, id string, containerID string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := networkContainerPayload{
		ContainerID: containerID,
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/networks/"+url.PathEscape(id)+"/connect", query, payload, nil)
}

func (c *Client) DisconnectNetwork(ctx context.Context, env string, id string, containerID string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := networkContainerPayload{
		ContainerID: containerID,
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/networks/"+url.PathEscape(id)+"/disconnect", query, payload, nil)
}

func (c *Client) ListVolumes(ctx context.Context, env string) ([]volumeResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out []volumeResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/volumes", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetVolumeInspect(ctx context.Context, env string, name string) (*volumeResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out volumeResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/volumes/"+url.PathEscape(name)+"/inspect", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateVolume(ctx context.Context, env string, payload volumePayload) (*volumeResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out volumeResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/volumes", query, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteVolume(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{
		"force": "true",
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/volumes/"+url.PathEscape(name), query, nil, nil)
}

func (c *Client) CloneVolume(ctx context.Context, env string, sourceName string, newName string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := volumeClonePayload{
		Name: newName,
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/volumes/"+url.PathEscape(sourceName)+"/clone", query, payload, nil)
}
