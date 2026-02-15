package provider

import "testing"

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
