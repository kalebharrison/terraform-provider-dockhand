package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGitRepositoryEnvironmentIDFromList(t *testing.T) {
	envID := int64(7)
	repos := []gitRepositoryResponse{
		{ID: 1, EnvironmentID: nil},
		{ID: 42, EnvironmentID: &envID},
		{Name: "by-name", EnvironmentID: &envID},
	}

	if got := gitRepositoryEnvironmentIDFromList(repos, "42", ""); got == nil || *got != envID {
		t.Fatalf("expected environment id 7 for repo 42, got %#v", got)
	}
	if got := gitRepositoryEnvironmentIDFromList(repos, "", "by-name"); got == nil || *got != envID {
		t.Fatalf("expected environment id 7 for repo by-name, got %#v", got)
	}
	if got := gitRepositoryEnvironmentIDFromList(repos, "99", ""); got != nil {
		t.Fatalf("expected nil for missing repo, got %#v", got)
	}
}

func TestMergeGitRepositoryStatePreservesPreferredEnvironmentID(t *testing.T) {
	preferred := gitRepositoryModel{
		EnvironmentID: types.StringValue("3"),
	}
	remote := gitRepositoryModel{
		EnvironmentID: types.StringNull(),
	}

	merged := mergeGitRepositoryState(preferred, remote)
	if merged.EnvironmentID.ValueString() != "3" {
		t.Fatalf("expected environment_id 3, got %q", merged.EnvironmentID.ValueString())
	}
}
