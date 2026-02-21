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
