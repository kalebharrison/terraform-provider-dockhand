package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListEnvironments(ctx context.Context) ([]environmentResponse, int, error) {
	var out []environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetEnvironment(ctx context.Context, id string) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateEnvironment(ctx context.Context, payload environmentPayload) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateEnvironment(ctx context.Context, id string, payload environmentPayload) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/environments/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/environments/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) TestEnvironmentConnection(ctx context.Context, payload environmentPayload) (*environmentTestResponse, int, error) {
	var out environmentTestResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments/test", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	if out.Info == nil {
		out.Info = map[string]any{}
	}
	if out.Hawser == nil {
		out.Hawser = map[string]any{}
	}
	return &out, status, nil
}

func (c *Client) TestEnvironmentConnectionByID(ctx context.Context, id string) (*environmentTestResponse, int, error) {
	var out environmentTestResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments/"+url.PathEscape(id)+"/test", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out.Info == nil {
		out.Info = map[string]any{}
	}
	if out.Hawser == nil {
		out.Hawser = map[string]any{}
	}
	return &out, status, nil
}

func (c *Client) DetectEnvironmentSockets(ctx context.Context) (*environmentDetectSocketResponse, int, error) {
	var out environmentDetectSocketResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/detect-socket", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	if out.Sockets == nil {
		out.Sockets = []any{}
	}
	return &out, status, nil
}

func (c *Client) CreateHawserToken(ctx context.Context, name string, environmentID int64, rawToken string) (*hawserTokenResponse, int, error) {
	payload := hawserTokenPayload{
		Name:          strings.TrimSpace(name),
		EnvironmentID: environmentID,
	}
	if t := strings.TrimSpace(rawToken); t != "" {
		payload.RawToken = &t
	}

	var out hawserTokenResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/hawser/tokens", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetEnvironmentTimezone(ctx context.Context, id string) (*environmentTimezoneResponse, int, error) {
	var out environmentTimezoneResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/"+url.PathEscape(id)+"/timezone", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) SetEnvironmentTimezone(ctx context.Context, id string, timezone string) (int, error) {
	payload := environmentTimezonePayload{Timezone: timezone}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments/"+url.PathEscape(id)+"/timezone", nil, payload, nil)
}

func (c *Client) GetEnvironmentUpdateCheck(ctx context.Context, id string) (*environmentUpdateCheckResponse, int, error) {
	var out environmentUpdateCheckResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/"+url.PathEscape(id)+"/update-check", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) SetEnvironmentUpdateCheck(ctx context.Context, id string, payload environmentUpdateCheckPayload) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments/"+url.PathEscape(id)+"/update-check", nil, payload, nil)
}

func (c *Client) GetEnvironmentImagePrune(ctx context.Context, id string) (*environmentImagePruneResponse, int, error) {
	var out environmentImagePruneResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/"+url.PathEscape(id)+"/image-prune", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) SetEnvironmentImagePrune(ctx context.Context, id string, payload environmentImagePrunePayload) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments/"+url.PathEscape(id)+"/image-prune", nil, payload, nil)
}

func (c *Client) GetScannerSettings(ctx context.Context, envID string, settingsOnly bool) (*scannerSettingsResponse, int, error) {
	query := map[string]string{}
	if settingsOnly {
		query["settingsOnly"] = "true"
	}
	if resolvedEnv := c.resolveEnv(envID); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out scannerSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/settings/scanner", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) SetScannerSettings(ctx context.Context, envID string, scanner string) (int, error) {
	parsedEnvID, err := strconv.ParseInt(strings.TrimSpace(envID), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid environment id %q for scanner settings: %w", envID, err)
	}

	payload := scannerSettingsPayload{
		Scanner: scanner,
		EnvID:   parsedEnvID,
	}

	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/settings/scanner", nil, payload, nil)
}

func (c *Client) RemoveScannerImage(ctx context.Context, envID string, scanner string) (bool, int, error) {
	scanner = strings.ToLower(strings.TrimSpace(scanner))
	if scanner != "grype" && scanner != "trivy" {
		return false, 0, fmt.Errorf("invalid scanner %q: expected grype or trivy", scanner)
	}

	query := map[string]string{
		"removeImages": "true",
		"scanner":      scanner,
	}
	if resolvedEnv := c.resolveEnv(envID); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out simpleSuccessResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodDelete, "/api/settings/scanner", query, nil, &out)
	if err != nil {
		return false, status, err
	}
	return out.Success, status, nil
}

func (c *Client) CheckScannerUpdates(ctx context.Context, envID string) (*scannerCheckUpdatesResponse, int, error) {
	query := map[string]string{
		"checkUpdates": "true",
	}
	if resolvedEnv := c.resolveEnv(envID); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var out scannerCheckUpdatesResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/settings/scanner", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}
