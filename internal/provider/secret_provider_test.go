package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSecretProviderIDJSON(t *testing.T) {
	payload := stackPayload{Name: "app", Compose: "services: {}\n"}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"name":"app","compose":"services: {}\n"}` {
		t.Fatalf("unset provider leaked into JSON: %s", raw)
	}

	id := int64(7)
	payload.SecretProviderID = secretProviderID{Include: true, Value: &id}
	payload.EnvVars = []stackEnvVariable{{Key: dockhandSecretSelectorVar, Value: "vault/prod", IsSecret: false}}
	raw, err = json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["secretProviderId"] != float64(7) {
		t.Fatalf("secretProviderId = %#v", decoded["secretProviderId"])
	}

	payload.SecretProviderID = secretProviderID{Include: true, Value: nil}
	payload.EnvVars = nil
	raw, err = json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["secretProviderId"]; !ok || decoded["secretProviderId"] != nil {
		t.Fatalf("expected JSON null secretProviderId, got %#v", decoded["secretProviderId"])
	}
}

func TestMergeSelectorVarsReplacesLegacyKey(t *testing.T) {
	merged := mergeSelectorVars([]selectorVar{
		{Key: "KEEP", Value: "1"},
		{Key: legacySecretSelectorVar, Value: "old"},
	}, "op://homelab/item", false)

	if len(merged) != 2 {
		t.Fatalf("len = %d, vars %#v", len(merged), merged)
	}
	if merged[0].Key != "KEEP" || merged[1].Key != dockhandSecretSelectorVar || merged[1].Value != "op://homelab/item" {
		t.Fatalf("merged = %#v", merged)
	}

	cleared := mergeSelectorVars(merged, "", true)
	if len(cleared) != 1 || cleared[0].Key != "KEEP" {
		t.Fatalf("cleared = %#v", cleared)
	}
}

func TestApplyGitStackSecretProvider(t *testing.T) {
	payload := gitStackPayload{
		StackName:   "app",
		ComposePath: "compose.yml",
		EnvVars: []gitStackEnvVarPayload{
			{Key: "KEEP", Value: "1"},
		},
	}
	plan := gitStackModel{
		SecretProviderID:   types.Int64Value(4),
		SecretProviderPath: types.StringValue("prod/app"),
	}
	applyGitStackSecretProvider(&payload, plan, gitStackModel{}, false)

	raw, err := marshalGitStackPayload(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["secretProviderId"] != float64(4) {
		t.Fatalf("secretProviderId = %#v", decoded["secretProviderId"])
	}
	vars, ok := decoded["envVars"].([]any)
	if !ok || len(vars) != 2 {
		t.Fatalf("envVars = %#v", decoded["envVars"])
	}

	cleared := gitStackPayload{StackName: "app", ComposePath: "compose.yml"}
	applyGitStackSecretProvider(&cleared, gitStackModel{}, plan, true)
	raw, err = marshalGitStackPayload(cleared)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["secretProviderId"] != nil {
		t.Fatalf("expected cleared provider, got %#v", decoded["secretProviderId"])
	}
	vars, ok = decoded["envVars"].([]any)
	if !ok || len(vars) != 0 {
		t.Fatalf("expected empty envVars, got %#v", decoded["envVars"])
	}
}
