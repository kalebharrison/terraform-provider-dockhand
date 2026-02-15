package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildContainerUpdatePayload(t *testing.T) {
	t.Run("typed fields only", func(t *testing.T) {
		plan := containerUpdateActionModel{
			CPUShares:   types.Int64Value(1024),
			PidsLimit:   types.Int64Value(256),
			MemoryBytes: types.Int64Value(134217728),
			NanoCPUs:    types.Int64Value(500000000),
		}

		out, err := buildContainerUpdatePayload(plan, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["CpuShares"] != int64(1024) {
			t.Fatalf("expected CpuShares")
		}
		if out["PidsLimit"] != int64(256) {
			t.Fatalf("expected PidsLimit")
		}
	})

	t.Run("payload json overrides typed", func(t *testing.T) {
		plan := containerUpdateActionModel{
			CPUShares: types.Int64Value(1024),
		}

		out, err := buildContainerUpdatePayload(plan, `{"CpuShares":2048}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["CpuShares"] != float64(2048) {
			t.Fatalf("expected payload_json override, got %#v", out["CpuShares"])
		}
	})
}
