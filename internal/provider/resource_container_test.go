package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseContainerUpdatePayload(t *testing.T) {
	t.Run("valid object", func(t *testing.T) {
		payload, err := parseContainerUpdatePayload(`{"RestartPolicy":{"Name":"no"}}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payload == nil {
			t.Fatalf("expected payload map")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if _, err := parseContainerUpdatePayload(`{bad`); err == nil {
			t.Fatalf("expected error for invalid json")
		}
	})
}

func TestFlattenContainerPorts(t *testing.T) {
	t.Run("converts host_port string to numeric payload", func(t *testing.T) {
		ports, err := flattenContainerPorts([]containerPortModel{
			{
				ContainerPort: types.Int64Value(80),
				HostPort:      types.StringValue("80"),
				Protocol:      types.StringNull(),
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ports) != 1 {
			t.Fatalf("expected 1 port payload, got %d", len(ports))
		}
		if ports[0].ContainerPort != 80 || ports[0].HostPort != 80 || ports[0].Protocol != "tcp" {
			t.Fatalf("unexpected payload: %+v", ports[0])
		}
	})

	t.Run("rejects invalid host_port", func(t *testing.T) {
		_, err := flattenContainerPorts([]containerPortModel{
			{
				ContainerPort: types.Int64Value(80),
				HostPort:      types.StringValue("abc"),
				Protocol:      types.StringValue("tcp"),
			},
		})
		if err == nil {
			t.Fatalf("expected error for invalid host_port")
		}
	})
}
