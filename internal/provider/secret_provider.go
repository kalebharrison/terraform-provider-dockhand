package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	dockhandSecretSelectorVar = "DOCKHAND_SECRET_SELECTOR"
	legacySecretSelectorVar   = "OP_ENVIRONMENT_ID"
)

// secretProviderID is omitted from JSON when unset. Removing a previously set
// value sends JSON null so Dockhand clears the binding. The omitzero tag uses IsZero.
type secretProviderID struct {
	Include bool
	Value   *int64
}

func (s secretProviderID) IsZero() bool { return !s.Include }

func (s secretProviderID) MarshalJSON() ([]byte, error) {
	if s.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*s.Value)
}

func secretProviderIDFromAttrs(plan, previous types.Int64, updating bool) secretProviderID {
	if !plan.IsNull() && !plan.IsUnknown() {
		value := plan.ValueInt64()
		return secretProviderID{Include: true, Value: &value}
	}
	if updating && !previous.IsNull() && !previous.IsUnknown() {
		return secretProviderID{Include: true, Value: nil}
	}
	return secretProviderID{}
}

type selectorVar struct {
	Key      string
	Value    string
	IsSecret bool
}

// mergeSelectorVars writes the Dockhand bulk-selector env var. The UI path
// field is that variable, not a separate stack column. A legacy
// OP_ENVIRONMENT_ID row is replaced by DOCKHAND_SECRET_SELECTOR.
func mergeSelectorVars(in []selectorVar, path string, clear bool) []selectorVar {
	out := make([]selectorVar, 0, len(in)+1)
	for _, item := range in {
		if item.Key == dockhandSecretSelectorVar || item.Key == legacySecretSelectorVar {
			continue
		}
		out = append(out, item)
	}
	if clear {
		return out
	}
	return append(out, selectorVar{Key: dockhandSecretSelectorVar, Value: path, IsSecret: false})
}

func selectorValue(vars []selectorVar) (string, bool) {
	for _, key := range []string{legacySecretSelectorVar, dockhandSecretSelectorVar} {
		for _, item := range vars {
			if item.Key == key {
				return item.Value, true
			}
		}
	}
	return "", false
}

func applyGitStackSecretProvider(payload *gitStackPayload, plan, previous gitStackModel, updating bool) {
	if payload == nil {
		return
	}
	payload.SecretProviderID = secretProviderIDFromAttrs(plan.SecretProviderID, previous.SecretProviderID, updating)

	set := !plan.SecretProviderPath.IsNull() && !plan.SecretProviderPath.IsUnknown()
	clear := updating && !set && !previous.SecretProviderPath.IsNull() && !previous.SecretProviderPath.IsUnknown()
	if !set && !clear {
		return
	}

	current := make([]selectorVar, 0, len(payload.EnvVars))
	for _, item := range payload.EnvVars {
		current = append(current, selectorVar(item))
	}
	path := ""
	if set {
		path = plan.SecretProviderPath.ValueString()
	}
	merged := mergeSelectorVars(current, path, clear)
	if clear && len(merged) == 0 {
		payload.EnvVars = nil
		payload.ForceEmptyEnvVars = true
		return
	}
	payload.EnvVars = make([]gitStackEnvVarPayload, 0, len(merged))
	for _, item := range merged {
		payload.EnvVars = append(payload.EnvVars, gitStackEnvVarPayload(item))
	}
}

func marshalGitStackPayload(payload gitStackPayload) (json.RawMessage, error) {
	raw, err := json.Marshal(payload)
	if err != nil || !payload.ForceEmptyEnvVars {
		return raw, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	fields["envVars"] = json.RawMessage("[]")
	return json.Marshal(fields)
}
