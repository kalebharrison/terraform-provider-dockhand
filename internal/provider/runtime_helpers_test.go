package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGitStackBoolTriggerRequested(t *testing.T) {
	t.Run("create with true", func(t *testing.T) {
		if !gitStackBoolTriggerRequested(types.BoolValue(true), types.BoolNull(), true) {
			t.Fatal("expected deploy trigger on create")
		}
	})
	t.Run("update unchanged true", func(t *testing.T) {
		if gitStackBoolTriggerRequested(types.BoolValue(true), types.BoolValue(true), false) {
			t.Fatal("expected no deploy trigger when plan matches state")
		}
	})
	t.Run("update false to true", func(t *testing.T) {
		if !gitStackBoolTriggerRequested(types.BoolValue(true), types.BoolValue(false), false) {
			t.Fatal("expected deploy trigger on transition to true")
		}
	})
}

func TestSplitImportEnvID(t *testing.T) {
	env, id := splitImportEnvID("11:abc123")
	if env != "11" || id != "abc123" {
		t.Fatalf("got env=%q id=%q", env, id)
	}
	env, id = splitImportEnvID("sha256:deadbeef")
	if env != "" || id != "sha256:deadbeef" {
		t.Fatalf("got env=%q id=%q want whole sha256 id", env, id)
	}
}

func TestRuntimeEnabledFromStatus(t *testing.T) {
	enabled, ok := runtimeEnabledFromStatus("running")
	if !ok || !enabled {
		t.Fatalf("expected running -> enabled true, got %v %v", enabled, ok)
	}
	enabled, ok = runtimeEnabledFromStatus("exited")
	if !ok || enabled {
		t.Fatalf("expected exited -> enabled false, got %v %v", enabled, ok)
	}
	enabled, ok = runtimeEnabledFromStatus("syncing")
	if !ok || !enabled {
		t.Fatalf("expected syncing -> enabled true, got %v %v", enabled, ok)
	}
	_, ok = runtimeEnabledFromStatus("created")
	if ok {
		t.Fatal("expected created -> preserve configured enabled (ok=false)")
	}
}

func TestJobPayloadIndicatesFailure(t *testing.T) {
	if jobPayloadIndicatesFailure(map[string]any{"success": false}) != true {
		t.Fatal("expected success=false to indicate failure")
	}
	if jobPayloadIndicatesFailure(map[string]any{"success": true}) {
		t.Fatal("expected success=true to pass")
	}
	if jobPayloadIndicatesFailure(map[string]any{"failed": float64(2)}) != true {
		t.Fatal("expected failed count to indicate failure")
	}
}

func TestNormalizeJobStatus(t *testing.T) {
	if got := normalizeJobStatus("completed", nil); got != "done" {
		t.Fatalf("got %q want done", got)
	}
}
