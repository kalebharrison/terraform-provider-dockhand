package provider

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveRequestRetryConfigDefaults(t *testing.T) {
	retry := resolveRequestRetryConfig(dockhandProviderModel{})
	if retry.attempts != defaultRequestRetryAttempts {
		t.Fatalf("attempts = %d, want %d", retry.attempts, defaultRequestRetryAttempts)
	}
	if retry.minDelay != time.Duration(defaultRequestRetryMinSeconds)*time.Second {
		t.Fatalf("minDelay = %s, want %s", retry.minDelay, time.Duration(defaultRequestRetryMinSeconds)*time.Second)
	}
	if retry.maxDelay != time.Duration(defaultRequestRetryMaxSeconds)*time.Second {
		t.Fatalf("maxDelay = %s, want %s", retry.maxDelay, time.Duration(defaultRequestRetryMaxSeconds)*time.Second)
	}
}

func TestResolveRequestRetryConfigFromProvider(t *testing.T) {
	cfg := dockhandProviderModel{
		RequestRetryAttempts:   types.Int64Value(8),
		RequestRetryMinSeconds: types.Int64Value(2),
		RequestRetryMaxSeconds: types.Int64Value(10),
	}
	retry := resolveRequestRetryConfig(cfg)
	if retry.attempts != 8 {
		t.Fatalf("attempts = %d, want 8", retry.attempts)
	}
	if retry.minDelay != 2*time.Second {
		t.Fatalf("minDelay = %s, want 2s", retry.minDelay)
	}
	if retry.maxDelay != 10*time.Second {
		t.Fatalf("maxDelay = %s, want 10s", retry.maxDelay)
	}
}

func TestShouldRetryRequestOnConnectionRefused(t *testing.T) {
	if !shouldRetryRequest(http.MethodPost, 0, fmt.Errorf("dial tcp: connect: connection refused")) {
		t.Fatal("expected POST connection refused to retry")
	}
}

func TestShouldNotRetryRequestOnClientError(t *testing.T) {
	if shouldRetryRequest(http.MethodPost, http.StatusBadRequest, nil) {
		t.Fatal("expected 400 POST not to retry")
	}
}
