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

func TestImagePullStreamError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "no error",
			body: `{"status":"Pulling from library/alpine","id":"latest"}` + "\n" +
				`{"status":"Download complete","id":"sha256:123"}`,
			want: "",
		},
		{
			name: "status error with error field",
			body: `{"status":"error","error":"manifest unknown"}`,
			want: "manifest unknown",
		},
		{
			name: "errorDetail message",
			body: `{"errorDetail":{"message":"dial tcp timeout"}}`,
			want: "dial tcp timeout",
		},
		{
			name: "non json lines ignored",
			body: `not-json` + "\n" + `{"status":"error","error":"pull failed"}`,
			want: "pull failed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := imagePullStreamError([]byte(tc.body))
			if got != tc.want {
				t.Fatalf("imagePullStreamError() = %q, want %q", got, tc.want)
			}
		})
	}
}
