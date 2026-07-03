package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// runtimeEnabledFromStatus maps Dockhand runtime status strings to desired enabled state.
// Returns (enabled, ok). When ok is false the status is unknown and state should not change.
func runtimeEnabledFromStatus(status string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running", "started", "up", "healthy", "active":
		return true, true
	case "starting", "syncing", "restarting":
		return true, true
	case "stopped", "exited", "down", "paused", "dead", "inactive":
		return false, true
	case "stopping":
		return false, true
	default:
		return false, false
	}
}

// splitImportEnvID parses `<env>:<id>` import IDs. IDs that contain non-numeric
// prefixes before the first colon (for example sha256 digests) are returned whole.
func splitImportEnvID(raw string) (env string, id string) {
	id = strings.TrimSpace(raw)
	if id == "" {
		return "", ""
	}
	idx := strings.Index(id, ":")
	if idx <= 0 {
		return "", id
	}
	prefix := strings.TrimSpace(id[:idx])
	if prefix == "" || !isDigitsOnly(prefix) {
		return "", id
	}
	return prefix, strings.TrimSpace(id[idx+1:])
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (c *Client) requireResolvedEnv(value string) (string, error) {
	env := strings.TrimSpace(c.resolveEnv(value))
	if env == "" {
		return "", fmt.Errorf("set resource `env` or provider `default_env`")
	}
	return env, nil
}

func (c *Client) persistEnvAttr(current types.String) types.String {
	resolved := strings.TrimSpace(c.resolveEnv(current.ValueString()))
	if resolved == "" {
		return current
	}
	return types.StringValue(resolved)
}

func gitStackBoolTriggerRequested(planVal, stateVal types.Bool, isCreate bool) bool {
	if planVal.IsNull() || planVal.IsUnknown() || !planVal.ValueBool() {
		return false
	}
	if isCreate {
		return true
	}
	return !planVal.Equal(stateVal)
}

type gitStackDeployTriggers struct {
	DeployNow     bool
	ForceRedeploy bool
}

func gitStackDeployTriggersFrom(plan, state gitStackModel, isCreate bool) gitStackDeployTriggers {
	return gitStackDeployTriggers{
		DeployNow:     gitStackBoolTriggerRequested(plan.DeployNow, state.DeployNow, isCreate),
		ForceRedeploy: gitStackBoolTriggerRequested(plan.ForceRedeploy, state.ForceRedeploy, isCreate),
	}
}

func validateEnvironmentConnectionPayload(connectionType string, payload environmentPayload) error {
	switch strings.ToLower(strings.TrimSpace(connectionType)) {
	case "direct":
		if payload.Host == nil || strings.TrimSpace(*payload.Host) == "" {
			return fmt.Errorf("host is required when connection_type is \"direct\"")
		}
		if payload.Port == nil || *payload.Port <= 0 {
			return fmt.Errorf("port must be > 0 when connection_type is \"direct\"")
		}
		if payload.Protocol == nil || strings.TrimSpace(*payload.Protocol) == "" {
			return fmt.Errorf("protocol is required when connection_type is \"direct\"")
		}
	case "socket":
		if payload.SocketPath == nil || strings.TrimSpace(*payload.SocketPath) == "" {
			return fmt.Errorf("socket_path is required when connection_type is \"socket\"")
		}
	default:
		if isAgentConnectionType(connectionType) && (payload.HawserToken == nil || strings.TrimSpace(*payload.HawserToken) == "") {
			return fmt.Errorf("agent_token is required when connection_type is %q", connectionType)
		}
	}
	return nil
}

// jobPayloadIndicatesFailure inspects inline batch/prune API payloads for failure signals.
func jobPayloadIndicatesFailure(result map[string]any) bool {
	if result == nil {
		return false
	}
	if raw, ok := result["success"]; ok {
		if b, ok := raw.(bool); ok {
			return !b
		}
	}
	for _, key := range []string{"failed", "failures", "errorCount", "errors"} {
		switch v := result[key].(type) {
		case float64:
			if v > 0 {
				return true
			}
		case int:
			if v > 0 {
				return true
			}
		case int64:
			if v > 0 {
				return true
			}
		}
	}
	if msg, ok := result["error"].(string); ok && strings.TrimSpace(msg) != "" {
		return true
	}
	return false
}

func normalizeJobStatus(status string, result map[string]any) string {
	normalized := strings.TrimSpace(status)
	if normalized == "" && result != nil {
		if extracted := extractBatchStatus(result); extracted != "" {
			normalized = extracted
		}
	}
	switch strings.ToLower(normalized) {
	case "complete", "completed", "success", "ok":
		return "done"
	default:
		return normalized
	}
}
