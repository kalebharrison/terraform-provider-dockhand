package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildGitStackPayloadWebhookRequiresSecretUnlessAutoGenerate(t *testing.T) {
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

	_, err := buildGitStackPayload(plan)
	if err == nil || !strings.Contains(err.Error(), "webhook_secret is required") {
		t.Fatalf("expected webhook_secret validation error, got: %v", err)
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
