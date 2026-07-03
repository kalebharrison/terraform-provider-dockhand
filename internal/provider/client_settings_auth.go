package provider

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) GetGeneralSettings(ctx context.Context) (*generalSettings, int, error) {
	var out generalSettings
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/settings/general", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGeneralSettings(ctx context.Context, payload generalSettings) (*generalSettings, int, error) {
	var out generalSettings
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/settings/general", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetAuthSettings(ctx context.Context) (*authSettingsResponse, int, error) {
	var out authSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/auth/settings", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateAuthSettings(ctx context.Context, payload authSettingsPayload) (*authSettingsResponse, int, error) {
	var out authSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/auth/settings", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetAuthProviders(ctx context.Context) (*authProvidersResponse, int, error) {
	var out authProvidersResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/auth/providers", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetLicense(ctx context.Context) (*licenseResponse, int, error) {
	var out licenseResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/license", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) SetLicense(ctx context.Context, payload licensePayload) (*licenseResponse, int, error) {
	var out licenseResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/license", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteLicense(ctx context.Context) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/license", nil, nil, nil)
}

func (c *Client) ListUsers(ctx context.Context) ([]userResponse, int, error) {
	var out []userResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/users", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) CreateUser(ctx context.Context, payload userPayload) (*userResponse, error) {
	var out userResponse
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/users", nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*userResponse, int, error) {
	var out userResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/users/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, payload userPayload) (*userResponse, error) {
	var out userResponse
	if _, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/users/"+url.PathEscape(id), nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/users/"+url.PathEscape(id), nil, nil, nil)
}
