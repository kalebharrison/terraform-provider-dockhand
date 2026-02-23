package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildGitStackPayloadWebhookDisabledAutoGenerateSendsEmptySecret(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WebhookSecret == nil || *payload.WebhookSecret != "" {
		t.Fatalf("expected empty webhook secret payload when auto-generate is disabled")
	}
}

func TestBuildGitStackPayloadWebhookAllowsAutoGenerate(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(true),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WebhookSecret != nil {
		t.Fatalf("expected webhook secret to remain nil when auto-generate is enabled")
	}
}

func TestBuildGitStackPayloadWebhookExplicitSecret(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(true),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringValue("custom-secret"),
		AutoUpdateEnabled:         types.BoolValue(false),
		AutoUpdateCron:            types.StringValue("0 3 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.WebhookSecret == nil || *payload.WebhookSecret != "custom-secret" {
		t.Fatalf("expected explicit webhook secret to be sent")
	}
}

func TestBuildGitStackPayloadSetsBothAutoUpdateKeys(t *testing.T) {
	plan := gitStackModel{
		StackName:                 types.StringValue("test-stack"),
		ComposePath:               types.StringValue("docker-compose.yml"),
		WebhookEnabled:            types.BoolValue(false),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		WebhookSecret:             types.StringNull(),
		AutoUpdateEnabled:         types.BoolValue(true),
		AutoUpdateCron:            types.StringValue("15 6 * * *"),
		DeployNow:                 types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
		URL:                       types.StringValue("https://example.com/repo.git"),
		Branch:                    types.StringValue("main"),
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !payload.AutoUpdateEnabled {
		t.Fatalf("expected AutoUpdateEnabled true in payload")
	}
	if payload.AutoUpdate == nil || !*payload.AutoUpdate {
		t.Fatalf("expected AutoUpdate pointer to be true in payload")
	}
}

func TestMergeGitStackStateWebhookSecretExplicitEmptyWins(t *testing.T) {
	preferred := gitStackModel{WebhookSecret: types.StringValue("")}
	remote := gitStackModel{WebhookSecret: types.StringValue("server-generated")}

	merged := mergeGitStackState(preferred, remote)
	if merged.WebhookSecret.IsNull() || merged.WebhookSecret.ValueString() != "" {
		t.Fatalf("expected explicit empty webhook_secret to win over server value")
	}
}

func TestMergeGitStackStateWebhookSecretUnsetInConfig(t *testing.T) {
	preferred := gitStackModel{
		WebhookSecret: types.StringNull(),
	}
	remote := gitStackModel{
		WebhookSecret: types.StringValue("server-generated-secret"),
	}

	merged := mergeGitStackState(preferred, remote)
	if !merged.WebhookSecret.IsNull() {
		t.Fatalf("expected webhook_secret to remain null when not configured, got %q", merged.WebhookSecret.ValueString())
	}
}

func TestMergeGitStackStateWebhookSecretConfiguredInConfig(t *testing.T) {
	preferred := gitStackModel{
		WebhookSecret: types.StringValue("from-config"),
	}
	remote := gitStackModel{
		WebhookSecret: types.StringNull(),
	}

	merged := mergeGitStackState(preferred, remote)
	if merged.WebhookSecret.IsNull() || merged.WebhookSecret.ValueString() != "from-config" {
		t.Fatalf("expected webhook_secret to preserve configured value, got null=%v value=%q", merged.WebhookSecret.IsNull(), merged.WebhookSecret.ValueString())
	}
}

func TestModelFromGitStackResponsePrefersAutoUpdateEnabled(t *testing.T) {
	falseVal := false
	resp := &gitStackResponse{
		ID:                1,
		StackName:         "test-stack",
		AutoUpdate:        true,
		AutoUpdateEnabled: &falseVal,
		WebhookEnabled:    false,
	}

	model := modelFromGitStackResponse(resp)
	if model.AutoUpdateEnabled.IsNull() || model.AutoUpdateEnabled.IsUnknown() {
		t.Fatalf("expected AutoUpdateEnabled to be set")
	}
	if model.AutoUpdateEnabled.ValueBool() {
		t.Fatalf("expected AutoUpdateEnabled=false when explicit autoUpdateEnabled is false")
	}
}

func TestModelFromGitStackResponseFallsBackToAutoUpdate(t *testing.T) {
	resp := &gitStackResponse{
		ID:                1,
		StackName:         "test-stack",
		AutoUpdate:        true,
		AutoUpdateEnabled: nil,
		WebhookEnabled:    false,
	}

	model := modelFromGitStackResponse(resp)
	if model.AutoUpdateEnabled.IsNull() || model.AutoUpdateEnabled.IsUnknown() {
		t.Fatalf("expected AutoUpdateEnabled to be set")
	}
	if !model.AutoUpdateEnabled.ValueBool() {
		t.Fatalf("expected AutoUpdateEnabled=true when falling back to autoUpdate")
	}
}
