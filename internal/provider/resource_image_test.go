package provider

import "testing"

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
