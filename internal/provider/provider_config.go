package provider

import (
	"os"
	"strings"
)

func providerEndpointValue(config dockhandProviderModel) string {
	endpoint := os.Getenv("DOCKHAND_ENDPOINT")
	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() {
		endpoint = config.Endpoint.ValueString()
	}
	return strings.TrimSpace(endpoint)
}

func providerEndpointDeferred(config dockhandProviderModel) bool {
	return config.Endpoint.IsUnknown()
}
