package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderEndpointValueUsesConfig(t *testing.T) {
	t.Setenv("DOCKHAND_ENDPOINT", "https://env.example")

	cfg := dockhandProviderModel{
		Endpoint: types.StringValue("https://config.example"),
	}
	if got := providerEndpointValue(cfg); got != "https://config.example" {
		t.Fatalf("expected config endpoint, got %q", got)
	}
}

func TestProviderEndpointValueUsesEnvWhenUnset(t *testing.T) {
	t.Setenv("DOCKHAND_ENDPOINT", "https://env.example")

	cfg := dockhandProviderModel{
		Endpoint: types.StringNull(),
	}
	if got := providerEndpointValue(cfg); got != "https://env.example" {
		t.Fatalf("expected env endpoint, got %q", got)
	}
}

func TestProviderEndpointDeferredWhenUnknown(t *testing.T) {
	cfg := dockhandProviderModel{
		Endpoint: types.StringUnknown(),
	}
	if !providerEndpointDeferred(cfg) {
		t.Fatal("expected unknown endpoint to defer provider configuration")
	}
}

func TestProviderEndpointNotDeferredWhenKnown(t *testing.T) {
	cfg := dockhandProviderModel{
		Endpoint: types.StringValue("https://dockhand.example"),
	}
	if providerEndpointDeferred(cfg) {
		t.Fatal("expected known endpoint not to defer provider configuration")
	}
}

func TestLoginRetriesTransientHTTPFailures(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "dockhand_session",
			Value:    "session-token",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
		_ = json.NewEncoder(w).Encode(loginResponse{Success: true})
	}))
	defer server.Close()

	cookie, err := loginWithRetry(context.Background(), server.URL, "admin", "password", "", "local", false, requestRetryConfig{attempts: 6, minDelay: time.Millisecond, maxDelay: time.Millisecond})
	if err != nil {
		t.Fatalf("loginWithRetry returned error: %v", err)
	}
	if cookie != "dockhand_session=session-token" {
		t.Fatalf("unexpected cookie: %q", cookie)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 login attempts, got %d", attempts)
	}
}

func TestLoginDoesNotRetryAuthFailures(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(loginResponse{Success: false, Error: "invalid credentials"})
	}))
	defer server.Close()

	_, err := loginWithRetry(context.Background(), server.URL, "admin", "wrong", "", "local", false, requestRetryConfig{attempts: 6, minDelay: time.Millisecond, maxDelay: time.Millisecond})
	if err == nil {
		t.Fatal("expected login error")
	}
	if attempts != 1 {
		t.Fatalf("expected single login attempt for auth failure, got %d", attempts)
	}
}

func TestIsLoginRetryableConnectionErrors(t *testing.T) {
	cases := []string{
		"dial tcp 127.0.0.1:443: connect: connection refused",
		"read tcp: connection reset by peer",
		"dockhand login failed (status 503): unavailable",
	}
	for _, msg := range cases {
		if !isLoginRetryable(&retryableError{msg: msg}) {
			t.Fatalf("expected retryable error for %q", msg)
		}
	}
	if isLoginRetryable(context.Canceled) {
		t.Fatal("expected context.Canceled not to retry")
	}
	if isLoginRetryable(fmt.Errorf("dockhand login failed: invalid credentials")) {
		t.Fatal("expected auth failure not to retry")
	}
}

type retryableError struct {
	msg string
}

func (e *retryableError) Error() string { return e.msg }
