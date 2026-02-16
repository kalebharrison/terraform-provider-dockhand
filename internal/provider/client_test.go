package provider

import "testing"

func TestNewClientAllowsEmptySessionCookie(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://example.com", "", "1", true)
	if err != nil {
		t.Fatalf("expected no error creating client without session cookie, got: %v", err)
	}
	if client == nil {
		t.Fatalf("expected client instance, got nil")
	}
}
