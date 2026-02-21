package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
