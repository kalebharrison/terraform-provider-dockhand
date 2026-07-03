package provider

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ListConfigSets(ctx context.Context) ([]configSetResponse, int, error) {
	var out []configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/config-sets", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetConfigSet(ctx context.Context, id string) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/config-sets/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateConfigSet(ctx context.Context, payload configSetPayload) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/config-sets", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateConfigSet(ctx context.Context, id string, payload configSetPayload) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/config-sets/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteConfigSet(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/config-sets/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListNotifications(ctx context.Context) ([]notificationResponse, int, error) {
	var out []notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/notifications", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetNotification(ctx context.Context, id string) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/notifications/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateNotification(ctx context.Context, payload notificationPayload) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/notifications", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateNotification(ctx context.Context, id string, payload notificationPayload) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/notifications/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteNotification(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/notifications/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) TestNotification(ctx context.Context, typ string, config map[string]any) (*notificationTestResponse, int, error) {
	payload := notificationTestPayload{
		Type:   strings.TrimSpace(typ),
		Config: config,
	}
	if payload.Config == nil {
		payload.Config = map[string]any{}
	}

	var out notificationTestResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/notifications/test", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}
