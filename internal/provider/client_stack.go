package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetStackBasePath(ctx context.Context) (*stackBasePathResponse, int, error) {
	var out stackBasePathResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks/base-path", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetStackDefaultPath(ctx context.Context, stackName string) (*stackDefaultPathResponse, int, error) {
	query := map[string]string{
		"name": strings.TrimSpace(stackName),
	}

	var out stackDefaultPathResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks/default-path", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateStack(ctx context.Context, env string, payload stackPayload) error {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks", query, payload, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListStacks(ctx context.Context, env string) ([]stackResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var raw json.RawMessage
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks", query, nil, &raw)
	if err != nil {
		return nil, status, err
	}

	stacks, parseErr := parseStacks(raw)
	if parseErr != nil {
		return nil, status, parseErr
	}

	return stacks, status, nil
}

func (c *Client) GetStackByName(ctx context.Context, env string, name string) (*stackResponse, bool, error) {
	stacks, _, err := c.ListStacks(ctx, env)
	if err != nil {
		return nil, false, err
	}

	for i := range stacks {
		if stacks[i].Name == name {
			return &stacks[i], true, nil
		}
	}

	return nil, false, nil
}

func (c *Client) GetStackEnvVars(ctx context.Context, env string, name string) ([]stackEnvVariable, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	var out stackEnvResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks/"+url.PathEscape(name)+"/env", query, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out.Variables, status, nil
}

func (c *Client) UpdateStackEnvVars(ctx context.Context, env string, name string, variables []stackEnvVariable) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := map[string]any{
		"variables": variables,
	}
	return c.doJSONWithStatus(ctx, http.MethodPut, "/api/stacks/"+url.PathEscape(name)+"/env", query, payload, nil)
}

func (c *Client) GetStackEnvRaw(ctx context.Context, env string, name string) (string, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	var out stackEnvRawResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks/"+url.PathEscape(name)+"/env/raw", query, nil, &out)
	if err != nil {
		return "", status, err
	}
	return out.Content, status, nil
}

func (c *Client) UpdateStackEnvRaw(ctx context.Context, env string, name string, content string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	payload := map[string]string{
		"content": content,
	}
	return c.doJSONWithStatus(ctx, http.MethodPut, "/api/stacks/"+url.PathEscape(name)+"/env/raw", query, payload, nil)
}

func (c *Client) StartStack(ctx context.Context, env string, name string) error {
	_, err := c.StartStackWithStatus(ctx, env, name)
	return err
}

func (c *Client) StartStackWithStatus(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/start", query, nil, nil)
}

func (c *Client) StopStack(ctx context.Context, env string, name string) error {
	_, err := c.StopStackWithStatus(ctx, env, name)
	return err
}

func (c *Client) StopStackWithStatus(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/stop", query, nil, nil)
}

func (c *Client) RestartStackWithStatus(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/restart", query, nil, nil)
}

func (c *Client) DownStackWithStatus(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/down", query, nil, nil)
}

func (c *Client) DeleteStack(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{
		"force": "true",
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/stacks/"+url.PathEscape(name), query, nil, nil)
}

func (c *Client) ScanStacks(ctx context.Context) (*stackScanResponse, int, error) {
	var out stackScanResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/scan", nil, map[string]any{}, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) AdoptStacks(ctx context.Context, payload stackAdoptPayload) (*stackAdoptResponse, int, error) {
	var out stackAdoptResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/adopt", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetStackSources(ctx context.Context) (map[string]stackSourceResponse, int, error) {
	var out map[string]stackSourceResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks/sources", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func parseStacks(raw json.RawMessage) ([]stackResponse, error) {
	var asArray []map[string]any
	if err := json.Unmarshal(raw, &asArray); err == nil {
		return mapsToStacks(asArray), nil
	}

	var asObject map[string]json.RawMessage
	if err := json.Unmarshal(raw, &asObject); err != nil {
		return nil, err
	}

	if stacksRaw, ok := asObject["stacks"]; ok {
		if err := json.Unmarshal(stacksRaw, &asArray); err != nil {
			return nil, err
		}
		return mapsToStacks(asArray), nil
	}

	return nil, fmt.Errorf("unexpected stack list response shape")
}

func mapsToStacks(input []map[string]any) []stackResponse {
	output := make([]stackResponse, 0, len(input))

	for _, item := range input {
		name := firstString(item, "name", "stack", "stack_name")
		compose := firstString(item, "compose", "manifest")
		status := firstString(item, "status")
		containers := toStringSlice(item["containers"])

		var details []stackContainerDetailResponse
		if rawDetails, ok := item["containerDetails"]; ok {
			if parsed := toStackContainerDetails(rawDetails); len(parsed) > 0 {
				details = parsed
			}
		}
		if name == "" {
			continue
		}
		output = append(output, stackResponse{
			Name:             name,
			Compose:          compose,
			Status:           status,
			Containers:       containers,
			ContainerDetails: details,
		})
	}

	return output
}

func toStackContainerDetails(value any) []stackContainerDetailResponse {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}

	out := make([]stackContainerDetailResponse, 0, len(raw))
	for _, entry := range raw {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, stackContainerDetailResponse{
			ID:           firstString(m, "id"),
			Name:         firstString(m, "name"),
			Service:      firstString(m, "service"),
			State:        firstString(m, "state"),
			Status:       firstString(m, "status"),
			Health:       firstString(m, "health"),
			Image:        firstString(m, "image"),
			RestartCount: firstInt64(m, "restartCount"),
		})
	}
	return out
}
