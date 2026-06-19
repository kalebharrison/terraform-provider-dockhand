package provider

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClientAllowsEmptySessionCookie(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://example.com", "", "", "1", true)
	if err != nil {
		t.Fatalf("expected no error creating client without session cookie, got: %v", err)
	}
	if client == nil {
		t.Fatalf("expected client instance, got nil")
	}
}

func TestClientUsesBearerTokenAuthHeader(t *testing.T) {
	t.Parallel()

	var gotCookie string
	var gotAuth string

	client, err := NewClient("http://example.com", "", "Bearer test-token", "1", false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotCookie = r.Header.Get("Cookie")
		gotAuth = r.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	var out map[string]any
	if _, err := client.doJSONWithStatus(t.Context(), http.MethodGet, "/api/settings/general", nil, nil, &out); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if gotCookie != "" {
		t.Fatalf("expected no cookie auth header, got %q", gotCookie)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected bearer auth header, got %q", gotAuth)
	}
}

func TestClientRetriesPOSTOnConnectionRefused(t *testing.T) {
	t.Parallel()

	attempts := 0
	client, err := NewClient("http://example.com", "", "", "1", true)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetRequestRetryPolicy(requestRetryConfig{
		attempts: 3,
		minDelay: time.Millisecond,
		maxDelay: time.Millisecond,
	})

	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, fmt.Errorf("dial tcp 127.0.0.1:80: connect: connection refused")
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(`{"id":1,"username":"admin"}`)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	var out userResponse
	if _, err := client.doJSONWithStatus(t.Context(), http.MethodPost, "/api/users", nil, map[string]any{"username": "admin"}, &out); err != nil {
		t.Fatalf("expected POST retry to succeed, got: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
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

func TestExtractBatchJobID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload map[string]any
		want    string
	}{
		{
			name:    "top-level jobId",
			payload: map[string]any{"jobId": "abc-123"},
			want:    "abc-123",
		},
		{
			name:    "nested data id",
			payload: map[string]any{"data": map[string]any{"id": "job-2"}},
			want:    "job-2",
		},
		{
			name:    "nested job job_id",
			payload: map[string]any{"job": map[string]any{"job_id": "job-3"}},
			want:    "job-3",
		},
		{
			name:    "fallback top-level id",
			payload: map[string]any{"id": "job-4"},
			want:    "job-4",
		},
		{
			name:    "missing id",
			payload: map[string]any{"ok": true},
			want:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := extractBatchJobID(tc.payload)
			if got != tc.want {
				t.Fatalf("extractBatchJobID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseJobResponse(t *testing.T) {
	t.Parallel()

	job := parseJobResponse(map[string]any{
		"job": map[string]any{
			"id":     "j-1",
			"status": "done",
			"result": map[string]any{"ok": true},
			"lines": []any{
				map[string]any{"message": "line-1"},
				"line-2",
			},
		},
	})

	if job.ID != "j-1" {
		t.Fatalf("job.ID = %q, want %q", job.ID, "j-1")
	}
	if job.Status != "done" {
		t.Fatalf("job.Status = %q, want %q", job.Status, "done")
	}
	if _, ok := job.Result["ok"]; !ok {
		t.Fatalf("job.Result missing key %q", "ok")
	}
	if len(job.Lines) != 2 {
		t.Fatalf("len(job.Lines) = %d, want 2", len(job.Lines))
	}
	if got := job.Lines[1].Data["message"]; got != "line-2" {
		t.Fatalf("job.Lines[1].Data[message] = %v, want %q", got, "line-2")
	}
}

func TestExtractBatchStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload map[string]any
		want    string
	}{
		{
			name:    "inline complete",
			payload: map[string]any{"type": "complete"},
			want:    "done",
		},
		{
			name:    "nested job running",
			payload: map[string]any{"job": map[string]any{"status": "running"}},
			want:    "running",
		},
		{
			name:    "status failed",
			payload: map[string]any{"status": "failed"},
			want:    "failed",
		},
		{
			name:    "missing",
			payload: map[string]any{"ok": true},
			want:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := extractBatchStatus(tc.payload)
			if got != tc.want {
				t.Fatalf("extractBatchStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}
