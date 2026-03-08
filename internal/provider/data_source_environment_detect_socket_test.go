package provider

import (
	"reflect"
	"testing"
)

func TestExtractSocketPaths(t *testing.T) {
	t.Parallel()

	in := []any{
		"/var/run/docker.sock",
		map[string]any{"path": "/run/user/1000/docker.sock"},
		map[string]any{"socketPath": "/run/user/1000/docker.sock"},
		map[string]any{"socket": "  /tmp/docker.sock  "},
		"",
		map[string]any{"unknown": "value"},
	}

	got := extractSocketPaths(in)
	want := []string{
		"/run/user/1000/docker.sock",
		"/tmp/docker.sock",
		"/var/run/docker.sock",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractSocketPaths() = %#v, want %#v", got, want)
	}
}
