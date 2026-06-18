package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderEndpointValueTrimsWhitespace(t *testing.T) {
	cfg := dockhandProviderModel{
		Endpoint: types.StringValue("  https://dockhand.example  "),
	}
	if got := providerEndpointValue(cfg); got != "https://dockhand.example" {
		t.Fatalf("expected trimmed endpoint, got %q", got)
	}
}
