package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type acceptancePRCI struct {
	Version int      `json:"version"`
	Suites  []string `json:"suites"`
}

func TestAcceptancePRCISuites(t *testing.T) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "acceptance_pr_ci.json"))
	if err != nil {
		t.Fatalf("read acceptance_pr_ci.json: %v", err)
	}

	var prCI acceptancePRCI
	if err := json.Unmarshal(raw, &prCI); err != nil {
		t.Fatalf("parse acceptance_pr_ci.json: %v", err)
	}
	if prCI.Version <= 0 {
		t.Fatalf("acceptance_pr_ci.json version must be positive, got %d", prCI.Version)
	}
	if len(prCI.Suites) == 0 {
		t.Fatal("acceptance_pr_ci.json must list at least one suite")
	}

	manifest := loadAcceptanceManifest(t)
	manifestPatterns := make([]*regexp.Regexp, 0)
	for _, section := range [][]acceptanceManifestEntry{manifest.Resources, manifest.DataSources} {
		for _, entry := range section {
			pattern := strings.TrimSpace(entry.AcceptanceTestRegex)
			re, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("invalid manifest regex %q: %v", pattern, err)
			}
			manifestPatterns = append(manifestPatterns, re)
		}
	}

	testNames := acceptanceTestFunctionNames(t)
	seen := map[string]struct{}{}
	for _, suite := range prCI.Suites {
		suite = strings.TrimSpace(suite)
		if suite == "" {
			t.Fatal("acceptance_pr_ci.json contains an empty suite name")
		}
		if _, ok := seen[suite]; ok {
			t.Fatalf("acceptance_pr_ci.json lists %q more than once", suite)
		}
		seen[suite] = struct{}{}

		if !containsString(testNames, suite) {
			t.Fatalf("acceptance_pr_ci.json suite %q is not a TestAcc function", suite)
		}

		matched := false
		for _, pattern := range manifestPatterns {
			if pattern.MatchString(suite) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("acceptance_pr_ci.json suite %q is not covered by acceptance_manifest.json", suite)
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
