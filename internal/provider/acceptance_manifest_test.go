package provider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type acceptanceManifest struct {
	Version     int                       `json:"version"`
	Resources   []acceptanceManifestEntry `json:"resources"`
	DataSources []acceptanceManifestEntry `json:"data_sources"`
}

type acceptanceManifestEntry struct {
	Name                string   `json:"name"`
	Mode                string   `json:"mode"`
	Operations          []string `json:"operations"`
	AcceptanceTestRegex string   `json:"acceptance_test_regex"`
}

var temporaryGenericAcceptanceRegexAllowlist = map[string]struct{}{}

// manifestOperationExemptions skips acceptance-source checks for known coverage gaps.
// Keys are manifest entry names; values list operations (import, update, delete).
var manifestOperationExemptions = map[string][]string{
	"dockhand_stack":                          {"delete"},
	"dockhand_git_repository":                 {"delete"},
	"dockhand_git_stack_env_file":             {"delete"},
	"dockhand_container_file":                 {"delete"},
	"dockhand_schedule":                       {"delete"},
	"dockhand_schedule_settings":              {"delete"},
	"dockhand_registry_image_delete_action":   {"delete"},
	"dockhand_git_repository_test_action":     {"delete"},
	"dockhand_git_stack_webhook_action":       {"delete"},
	"dockhand_git_stack_deploy_action":        {"delete"},
	"dockhand_notification_test_action":       {"delete"},
	"dockhand_environment_test_action":        {"delete"},
	"dockhand_environment_scanner_action":     {"delete"},
	"dockhand_network_connection_action":      {"delete"},
	"dockhand_volume_clone_action":            {"delete"},
	"dockhand_image_push_action":              {"delete"},
	"dockhand_image_scan_action":              {"delete"},
	"dockhand_container_action":               {"delete"},
	"dockhand_container_rename_action":        {"delete"},
	"dockhand_container_update_action":        {"delete"},
	"dockhand_container_check_updates_action": {"delete"},
	"dockhand_schedule_run_action":            {"delete"},
	"dockhand_prune_action":                   {"delete"},
	"dockhand_batch_action":                   {"delete"},
	"dockhand_stack_action":                   {"delete"},
	"dockhand_stack_scan_action":              {"delete"},
	"dockhand_stack_adopt_action":             {"delete"},
}

func TestAcceptanceManifestCoverage(t *testing.T) {
	t.Helper()

	manifest := loadAcceptanceManifest(t)
	if manifest.Version <= 0 {
		t.Fatalf("manifest version must be positive, got %d", manifest.Version)
	}

	actualResources, actualDataSources := providerSurfaceNames(t)
	testNames := acceptanceTestFunctionNames(t)

	validateManifestSection(t, "resource", actualResources, manifest.Resources, testNames)
	validateManifestSection(t, "data source", actualDataSources, manifest.DataSources, testNames)
}

func loadAcceptanceManifest(t *testing.T) acceptanceManifest {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "acceptance_manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var manifest acceptanceManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse manifest json: %v", err)
	}

	return manifest
}

func providerSurfaceNames(t *testing.T) ([]string, []string) {
	t.Helper()

	prov := New("dev")()
	ctx := context.Background()

	resourceNames := make([]string, 0)
	for _, f := range prov.Resources(ctx) {
		r := f()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "dockhand"}, &resp)
		resourceNames = append(resourceNames, resp.TypeName)
	}

	dataSourceNames := make([]string, 0)
	for _, f := range prov.DataSources(ctx) {
		d := f()
		var resp datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "dockhand"}, &resp)
		dataSourceNames = append(dataSourceNames, resp.TypeName)
	}

	slices.Sort(resourceNames)
	slices.Sort(dataSourceNames)
	return resourceNames, dataSourceNames
}

func acceptanceTestFunctionNames(t *testing.T) []string {
	t.Helper()

	pattern := regexp.MustCompile(`func\s+(TestAcc[[:alnum:]_]+)\s*\(`)

	matches := make([]string, 0)
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatalf("glob test files: %v", err)
	}
	for _, f := range files {
		// #nosec G304 -- f comes from filepath.Glob("*_test.go") in this package.
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, m := range pattern.FindAllStringSubmatch(string(raw), -1) {
			if len(m) >= 2 {
				matches = append(matches, m[1])
			}
		}
	}

	if len(matches) == 0 {
		t.Fatalf("no TestAcc functions discovered")
	}

	slices.Sort(matches)
	return slices.Compact(matches)
}

func validateManifestSection(t *testing.T, surfaceType string, actual []string, entries []acceptanceManifestEntry, testNames []string) {
	t.Helper()

	seen := map[string]struct{}{}
	for _, e := range entries {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			t.Fatalf("%s manifest contains empty name", surfaceType)
		}
		if _, ok := seen[name]; ok {
			t.Fatalf("%s %q listed multiple times in manifest", surfaceType, name)
		}
		seen[name] = struct{}{}
		validateManifestEntry(t, surfaceType, e, testNames)
	}

	for _, name := range actual {
		if _, ok := seen[name]; !ok {
			t.Fatalf("manifest missing %s %q", surfaceType, name)
		}
	}

	for name := range seen {
		if !slices.Contains(actual, name) {
			t.Fatalf("manifest contains unknown %s %q", surfaceType, name)
		}
	}
}

func validateManifestEntry(t *testing.T, surfaceType string, e acceptanceManifestEntry, testNames []string) {
	t.Helper()

	mode := strings.TrimSpace(e.Mode)
	switch mode {
	case "stateful", "action", "data_source":
	default:
		t.Fatalf("%s %q has unsupported mode %q", surfaceType, e.Name, e.Mode)
	}

	ops := make(map[string]struct{}, len(e.Operations))
	for _, op := range e.Operations {
		trimmed := strings.TrimSpace(op)
		if trimmed == "" {
			t.Fatalf("%s %q has empty operation", surfaceType, e.Name)
		}
		ops[trimmed] = struct{}{}
	}

	switch mode {
	case "stateful":
		requireOps(t, surfaceType, e.Name, ops, []string{"create", "read", "update", "import", "delete"})
	case "action":
		requireOps(t, surfaceType, e.Name, ops, []string{"create", "read", "delete"})
	case "data_source":
		requireOps(t, surfaceType, e.Name, ops, []string{"read"})
	}

	testRegex := strings.TrimSpace(e.AcceptanceTestRegex)
	if testRegex == "" {
		t.Fatalf("%s %q missing acceptance_test_regex", surfaceType, e.Name)
	}
	if testRegex == "TestAcc" {
		if _, ok := temporaryGenericAcceptanceRegexAllowlist[e.Name]; !ok {
			t.Fatalf("%s %q uses bare generic acceptance_test_regex %q; add explicit TestAcc... coverage or a temporary allowlist entry", surfaceType, e.Name, testRegex)
		}
	}
	re, err := regexp.Compile(testRegex)
	if err != nil {
		t.Fatalf("%s %q has invalid acceptance_test_regex %q: %v", surfaceType, e.Name, testRegex, err)
	}

	for _, testName := range testNames {
		if re.MatchString(testName) {
			validateManifestOperationsInTests(t, surfaceType, e, testNames, re)
			return
		}
	}
	t.Fatalf("%s %q regex %q does not match any TestAcc function", surfaceType, e.Name, testRegex)
}

func validateManifestOperationsInTests(t *testing.T, surfaceType string, e acceptanceManifestEntry, testNames []string, re *regexp.Regexp) {
	t.Helper()

	if len(e.Operations) == 0 {
		return
	}

	source := acceptanceTestSourceForRegex(t, testNames, re)
	exempt := map[string]struct{}{}
	for _, op := range manifestOperationExemptions[e.Name] {
		exempt[op] = struct{}{}
	}

	for _, op := range e.Operations {
		if _, ok := exempt[op]; ok {
			continue
		}
		switch e.Mode {
		case "stateful":
			switch op {
			case "import":
				if !strings.Contains(source, "ImportState") {
					t.Fatalf("%s %q requires import coverage in acceptance tests (ImportState)", surfaceType, e.Name)
				}
			case "update":
				if strings.Count(source, "resource.TestStep{") < 2 && !strings.Contains(source, "ImportStateVerify") {
					t.Fatalf("%s %q requires update coverage in acceptance tests (multiple TestStep blocks)", surfaceType, e.Name)
				}
			case "delete":
				if !strings.Contains(source, "CheckDestroy") && !strings.Contains(source, "Destroy:") {
					t.Fatalf("%s %q requires delete coverage in acceptance tests (CheckDestroy)", surfaceType, e.Name)
				}
			}
		case "action":
			switch op {
			case "delete":
				if strings.Count(source, "resource.TestStep{") < 1 {
					t.Fatalf("%s %q requires action acceptance test steps", surfaceType, e.Name)
				}
			}
		}
	}
}

func acceptanceTestSourceForRegex(t *testing.T, testNames []string, re *regexp.Regexp) string {
	t.Helper()

	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatalf("glob test files: %v", err)
	}

	var b strings.Builder
	for _, fileName := range files {
		// #nosec G304 -- fileName comes from filepath.Glob("*_test.go") in this package.
		raw, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		text := string(raw)
		for _, testName := range testNames {
			if !re.MatchString(testName) {
				continue
			}
			if strings.Contains(text, "func "+testName+"(t *testing.T)") {
				b.WriteString(text)
				break
			}
		}
	}
	return b.String()
}

func requireOps(t *testing.T, surfaceType string, name string, ops map[string]struct{}, required []string) {
	t.Helper()
	for _, op := range required {
		if _, ok := ops[op]; !ok {
			t.Fatalf("%s %q missing required operation %q", surfaceType, name, op)
		}
	}
}
