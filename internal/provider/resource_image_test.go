package provider

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestImageNameFromTags(t *testing.T) {
	name := imageNameFromTags(&imageResponse{
		Tags: []string{"library/busybox:1.36.1", "busybox:1.36.1"},
	})
	if name != "busybox:1.36.1" {
		t.Fatalf("got %q want busybox:1.36.1", name)
	}

	name = imageNameFromTags(&imageResponse{Tags: []string{"nginx:latest"}})
	if name != "nginx:latest" {
		t.Fatalf("got %q want nginx:latest", name)
	}
}

func TestImageNameForStatePrefersConfigured(t *testing.T) {
	got := imageNameForState("custom:tag", &imageResponse{Tags: []string{"other:tag"}})
	if got != "custom:tag" {
		t.Fatalf("got %q want custom:tag", got)
	}
}

func TestDeleteImageRetriesRunningContainerConflict(t *testing.T) {
	attempts := 0
	client, err := NewClient("http://example.com", "", "", "1", true)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client.SetRequestRetryPolicy(requestRetryConfig{
		attempts: 3,
		minDelay: time.Nanosecond,
		maxDelay: time.Nanosecond,
	})
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return &http.Response{
				StatusCode: http.StatusConflict,
				Body: io.NopCloser(strings.NewReader(
					`{"error":"Cannot delete image: it is being used by a running container. Stop the container first."}`,
				)),
				Header:  make(http.Header),
				Request: r,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	status, err := client.DeleteImage(t.Context(), "1", "sha256:test")
	if err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}
	if status != http.StatusNoContent {
		t.Fatalf("DeleteImage() status = %d, want %d", status, http.StatusNoContent)
	}
	if attempts != 3 {
		t.Fatalf("DeleteImage() attempts = %d, want 3", attempts)
	}
}
