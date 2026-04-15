package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccScannerAndImageScanActionsTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	testAccConfigureProviderEnv(t)

	dindHost := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_DIND_HOST"))
	if dindHost == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_DIND_HOST")
	}

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	envName := "tf-acc-scan-env-" + suffix
	imageName := "busybox:1.36.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccScannerAndImageScanConfig(envName, dindHost, imageName, "install_grype", "acc-scan-1", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "name", envName),
					resource.TestCheckResourceAttr("dockhand_environment.test", "vulnerability_scanning_enabled", "true"),
					resource.TestCheckResourceAttr("dockhand_environment_scanner_action.test", "action", "install_grype"),
					resource.TestCheckResourceAttr("dockhand_image_scan_action.test", "result", "scan_requested"),
					testAccCheckScannerAvailability("dockhand_environment.test", endpoint, username, password, "grype", true),
				),
			},
			{
				Config: testAccScannerAndImageScanConfig(envName, dindHost, imageName, "remove_grype", "acc-scan-2", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "vulnerability_scanning_enabled", "false"),
					resource.TestCheckResourceAttr("dockhand_environment_scanner_action.test", "action", "remove_grype"),
					testAccCheckScannerActionSuccess("dockhand_environment_scanner_action.test"),
				),
			},
		},
	})
}

func TestAccImagePushActionTerraform(t *testing.T) {
	env := testAccConfigureProviderEnv(t)

	registryURL := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_REGISTRY_URL"))
	if registryURL == "" {
		registryURL = "http://registry:5000"
	}
	registryHostURL := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_REGISTRY_HOST_URL"))
	if registryHostURL == "" {
		registryHostURL = "http://127.0.0.1:25000"
	}

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	registryName := "tf-acc-registry-" + suffix
	imageName := "busybox:1.36.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePushActionConfig(env, registryName, registryURL, imageName, "acc-push-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_registry.test", "url", registryURL),
					resource.TestCheckResourceAttr("dockhand_image_push_action.test", "result", "push_requested"),
					resource.TestCheckResourceAttr("dockhand_image_push_action.test", "trigger", "acc-push-1"),
					testAccCheckRegistryCatalogEventuallyNonEmpty(registryHostURL),
				),
			},
			{
				Config: testAccImagePushActionConfig(env, registryName, registryURL, imageName, "acc-push-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_image_push_action.test", "trigger", "acc-push-2"),
					testAccCheckRegistryCatalogEventuallyNonEmpty(registryHostURL),
				),
			},
		},
	})
}

func TestAccStackAdoptActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	testAccConfigureProviderEnv(t)

	envID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_STACK_ADOPT_ENV_ID"))
	if envID == "" {
		envID = testAccDefaultEnv()
	}
	stackName := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_STACK_ADOPT_NAME"))
	composePath := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_STACK_ADOPT_COMPOSE_PATH"))
	if stackName == "" || composePath == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_STACK_ADOPT_NAME and DOCKHAND_TEST_STACK_ADOPT_COMPOSE_PATH")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStackAdoptActionConfig(envID, stackName, composePath, "acc-adopt-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_adopt_action.test", "trigger", "acc-adopt-1"),
					resource.TestCheckOutput("adopted_contains_target", "true"),
					resource.TestCheckOutput("failed_is_empty", "true"),
					testAccCheckRuntimeStackExists(endpoint, username, password, envID, stackName),
				),
			},
			{
				Config: testAccStackAdoptCleanupConfig(envID, stackName, "acc-adopt-cleanup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_action.cleanup", "action", "down"),
					testAccCheckRuntimeStackQuiesced(endpoint, username, password, envID, stackName),
				),
			},
		},
	})
}

func testAccScannerAndImageScanConfig(envName string, dindHost string, imageName string, scannerAction string, trigger string, enableScanning bool) string {
	imageBlock := ""
	if scannerAction == "install_grype" {
		imageBlock = fmt.Sprintf(`
resource "dockhand_image" "test" {
  env             = dockhand_environment.test.id
  name            = %q
  scan_after_pull = false
}

resource "dockhand_image_scan_action" "test" {
  env        = dockhand_environment.test.id
  image_name = dockhand_image.test.name
  trigger    = %q

  depends_on = [
    dockhand_environment_scanner_action.test,
    dockhand_image.test,
  ]
}
`, imageName, trigger)
	}

	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_environment" "test" {
  name            = %q
  connection_type = "direct"
  host            = %q
  port            = 2375
  protocol        = "http"
  tls_skip_verify = false
  icon            = "globe"

  collect_activity               = true
  collect_metrics                = true
  highlight_changes              = true
  timezone                       = "UTC"
  vulnerability_scanning_enabled = %t
  vulnerability_scanner          = "both"
  ensure_grype_installed         = false
  ensure_trivy_installed         = false
}

resource "dockhand_environment_scanner_action" "test" {
  env     = dockhand_environment.test.id
  action  = %q
  trigger = %q
}
%s
`, envName, dindHost, enableScanning, scannerAction, trigger, imageBlock)
}

func testAccImagePushActionConfig(env string, registryName string, registryURL string, imageName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_registry" "test" {
  name = %q
  url  = %q
}

resource "dockhand_image" "test" {
  env             = %q
  name            = %q
  scan_after_pull = false
}

resource "dockhand_image_push_action" "test" {
  env         = %q
  image_id    = dockhand_image.test.id
  registry_id = tonumber(dockhand_registry.test.id)
  trigger     = %q
}
`, registryName, registryURL, env, imageName, env, trigger)
}

func testAccStackAdoptActionConfig(envID string, stackName string, composePath string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_stack_adopt_action" "test" {
  environment_id = tonumber(%q)
  trigger        = %q

  stacks = [
    {
      name         = %q
      compose_path = %q
    }
  ]
}

output "adopted_contains_target" {
  value = contains(dockhand_stack_adopt_action.test.adopted, %q)
}

output "failed_is_empty" {
  value = length(dockhand_stack_adopt_action.test.failed) == 0
}
`, envID, trigger, stackName, composePath, stackName)
}

func testAccStackAdoptCleanupConfig(env string, stackName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_stack_action" "cleanup" {
  env        = %q
  stack_name = %q
  action     = "down"
  trigger    = %q
}
`, env, stackName, trigger)
}

func testAccCheckScannerAvailability(environmentResourceName string, endpoint string, username string, password string, scanner string, expected bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[environmentResourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", environmentResourceName)
		}
		envID := strings.TrimSpace(rs.Primary.ID)
		if envID == "" {
			return fmt.Errorf("resource %q has empty environment id", environmentResourceName)
		}

		sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
		if err != nil {
			return err
		}

		client, err := NewClient(endpoint, sessionCookie, envID, true)
		if err != nil {
			return err
		}

		deadline := time.Now().Add(2 * time.Minute)
		var lastActual bool
		var lastErr error
		for {
			resp, _, err := client.GetScannerSettings(context.Background(), envID, false)
			if err == nil && resp != nil {
				lastActual = scannerAvailability(resp, scanner)
				if lastActual == expected {
					return nil
				}
			} else if err != nil {
				lastErr = err
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("scanner %q availability for env %s did not become %t (last_actual=%t last_err=%v)", scanner, envID, expected, lastActual, lastErr)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func testAccCheckScannerActionSuccess(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		raw := strings.TrimSpace(rs.Primary.Attributes["result_json"])
		if raw == "" {
			return fmt.Errorf("resource %q has empty result_json", resourceName)
		}

		var payload struct {
			Success bool `json:"success"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("parse %q result_json: %w", resourceName, err)
		}
		if !payload.Success {
			return fmt.Errorf("resource %q reported unsuccessful scanner action: %s", resourceName, raw)
		}
		return nil
	}
}

func testAccCheckRegistryCatalogEventuallyNonEmpty(registryHostURL string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		deadline := time.Now().Add(2 * time.Minute)
		var lastRepos []string
		var lastErr error
		for {
			//nolint:gosec // Acceptance test fixture URL is provisioned by the local harness, not user input.
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, strings.TrimRight(registryHostURL, "/")+"/v2/_catalog", nil)
			if err == nil {
				//nolint:gosec // Acceptance test fixture URL is provisioned by the local harness, not user input.
				res, doErr := http.DefaultClient.Do(req)
				if doErr == nil {
					var payload struct {
						Repositories []string `json:"repositories"`
					}
					decodeErr := json.NewDecoder(res.Body).Decode(&payload)
					closeErr := res.Body.Close()
					switch {
					case decodeErr == nil && closeErr == nil && res.StatusCode >= 200 && res.StatusCode <= 299:
						lastRepos = payload.Repositories
						if len(lastRepos) > 0 {
							return nil
						}
					case decodeErr != nil:
						lastErr = decodeErr
					case closeErr != nil:
						lastErr = closeErr
					default:
						lastErr = fmt.Errorf("registry catalog status %d", res.StatusCode)
					}
				} else {
					lastErr = doErr
				}
			} else {
				lastErr = err
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("registry catalog at %s remained empty after push (last_repositories=%v last_err=%v)", registryHostURL, lastRepos, lastErr)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func testAccCheckRuntimeStackQuiesced(endpoint string, username string, password string, env string, stackName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
		if err != nil {
			return err
		}

		client, err := NewClient(endpoint, sessionCookie, env, true)
		if err != nil {
			return err
		}

		deadline := time.Now().Add(90 * time.Second)
		for {
			item, found, err := client.GetStackByName(context.Background(), env, stackName)
			if err == nil && (!found || (item != nil && len(item.Containers) == 0)) {
				return nil
			}
			if time.Now().After(deadline) {
				if err != nil {
					return fmt.Errorf("check runtime stack quiesced state for %q: %w", stackName, err)
				}
				if item != nil {
					return fmt.Errorf("runtime stack %q still has containers after cleanup (status=%q containers=%v)", stackName, item.Status, item.Containers)
				}
				return fmt.Errorf("runtime stack %q did not quiesce after cleanup", stackName)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
