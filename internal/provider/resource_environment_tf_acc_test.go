package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccEnvironmentResourceDirectDinDTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	dindHost := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_DIND_HOST"))
	if dindHost == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_DIND_HOST")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	envName := "tf-acc-dind-env-" + suffix
	containerName := "tf-acc-dind-container-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDirectDinDConfig(envName, dindHost, "globe", containerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "name", envName),
					resource.TestCheckResourceAttr("dockhand_environment.test", "connection_type", "direct"),
					resource.TestCheckResourceAttr("dockhand_environment.test", "host", dindHost),
					resource.TestCheckResourceAttr("dockhand_environment.test", "port", "2375"),
					resource.TestCheckResourceAttr("dockhand_environment.test", "protocol", "http"),
					resource.TestCheckResourceAttr("dockhand_environment.test", "icon", "globe"),
					resource.TestCheckResourceAttrSet("dockhand_environment.test", "id"),
					resource.TestCheckResourceAttr("dockhand_container.test", "name", containerName),
					resource.TestCheckResourceAttrSet("dockhand_container.test", "id"),
				),
			},
			{
				Config: testAccEnvironmentDirectDinDConfig(envName, dindHost, "server", containerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "icon", "server"),
				),
			},
		},
	})
}

func testAccEnvironmentDirectDinDConfig(envName string, dindHost string, icon string, containerName string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_environment" "test" {
  name            = %q
  connection_type = "direct"
  host            = %q
  port            = 2375
  protocol        = "http"
  tls_skip_verify = false
  icon            = %q

  collect_activity  = true
  collect_metrics   = true
  highlight_changes = true
  timezone          = "UTC"
}

resource "dockhand_image" "test" {
  env             = dockhand_environment.test.id
  name            = "busybox:1.36.1"
  scan_after_pull = false
}

resource "dockhand_container" "test" {
  env     = dockhand_environment.test.id
  name    = %q
  image   = dockhand_image.test.name
  enabled = false
}
`, envName, dindHost, icon, containerName)
}

func TestAccEnvironmentResourceAgentTokenTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	agentToken := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_AGENT_TOKEN"))
	if agentToken == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_AGENT_TOKEN")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	envName := "tf-acc-agent-env-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentAgentConfig(envName, agentToken, "globe"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "name", envName),
					resource.TestCheckResourceAttr("dockhand_environment.test", "connection_type", "agent"),
					resource.TestCheckResourceAttr("dockhand_environment.test", "agent_token", agentToken),
					resource.TestCheckResourceAttr("dockhand_environment.test", "icon", "globe"),
					resource.TestCheckResourceAttrSet("dockhand_environment.test", "id"),
					testAccCheckHawserConnected("dockhand_environment.test", endpoint, username, password, agentToken),
				),
			},
			{
				Config: testAccEnvironmentAgentConfig(envName, agentToken, "server"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment.test", "icon", "server"),
					testAccCheckHawserConnected("dockhand_environment.test", endpoint, username, password, agentToken),
				),
			},
		},
	})
}

func testAccEnvironmentAgentConfig(envName string, token string, icon string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_environment" "test" {
  name            = %q
  connection_type = "agent"
  agent_token     = %q
  icon            = %q

  collect_activity  = true
  collect_metrics   = true
  highlight_changes = true
  timezone          = "UTC"
}
`, envName, token, icon)
}

func testAccCheckHawserConnected(resourceName string, endpoint string, username string, password string, agentToken string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if err := testAccEnsureHawserRunning(agentToken); err != nil {
			return err
		}

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}
		envID := strings.TrimSpace(rs.Primary.ID)
		if envID == "" {
			return fmt.Errorf("resource %q has empty environment id", resourceName)
		}

		deadline := time.Now().Add(3 * time.Minute)
		var lastConnections int64
		var lastStatus string
		var lastErr error
		var lastEnvTestErr string
		var lastHawserSummary string

		for time.Now().Before(deadline) {
			sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
			if err != nil {
				lastErr = err
				time.Sleep(3 * time.Second)
				continue
			}

			client, err := NewClient(endpoint, sessionCookie, testAccDefaultEnv(), false)
			if err != nil {
				lastErr = err
				time.Sleep(3 * time.Second)
				continue
			}

			if status, _, err := client.GetHawserStatus(context.Background()); err == nil && status != nil {
				lastConnections = status.ActiveConnections
				lastStatus = status.Status
				lastHawserSummary = strings.TrimSpace(status.Message)
			}

			out, _, err := client.TestEnvironmentConnectionByID(context.Background(), envID)
			if err != nil {
				lastErr = err
				time.Sleep(3 * time.Second)
				continue
			}

			if out.Success {
				return nil
			}
			lastEnvTestErr = strings.TrimSpace(out.Error)
			if lastEnvTestErr == "" {
				lastEnvTestErr = "environment test returned success=false without error"
			}

			time.Sleep(3 * time.Second)
		}

		return fmt.Errorf(
			"hawser-backed environment did not pass /api/environments/%s/test before timeout (hawser_status=%q active_connections=%d hawser_message=%q env_test_error=%q last_err=%v hawser_logs_tail=%q)",
			envID,
			lastStatus,
			lastConnections,
			lastHawserSummary,
			lastEnvTestErr,
			lastErr,
			testAccTailHawserLogs(),
		)
	}
}

func testAccEnsureHawserRunning(agentToken string) error {
	dockerHost := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_DOCKER_HOST"))
	if dockerHost == "" {
		dockerHost = "tcp://127.0.0.1:23750"
	}

	containerName := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_CONTAINER"))
	if containerName == "" {
		return fmt.Errorf("acceptance test requires DOCKHAND_TEST_HAWSER_CONTAINER")
	}

	serverURL := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_SERVER_URL"))
	if serverURL == "" {
		return fmt.Errorf("acceptance test requires DOCKHAND_TEST_HAWSER_SERVER_URL")
	}

	image := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_IMAGE"))
	if image == "" {
		image = "ghcr.io/finsys/hawser:latest"
	}

	agentName := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_AGENT_NAME"))
	if agentName == "" {
		agentName = "ci-hawser"
	}

	inspectCtx, inspectCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer inspectCancel()
	// #nosec G204 -- acceptance test intentionally launches the docker CLI with harness-provided values.
	inspectOut, inspectErr := exec.CommandContext(inspectCtx, "docker", "--host", dockerHost, "inspect", "-f", "{{.State.Running}}", containerName).CombinedOutput()
	if inspectErr == nil && strings.TrimSpace(string(inspectOut)) == "true" {
		return nil
	}

	rmCtx, rmCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer rmCancel()
	// #nosec G204 -- acceptance test intentionally launches the docker CLI with harness-provided values.
	_ = exec.CommandContext(rmCtx, "docker", "--host", dockerHost, "rm", "-f", containerName).Run()

	runCtx, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer runCancel()
	args := []string{
		"--host", dockerHost,
		"run", "-d",
		"--name", containerName,
		"--restart", "unless-stopped",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-e", "DOCKHAND_SERVER_URL=" + serverURL,
		"-e", "TOKEN=" + agentToken,
		"-e", "AGENT_NAME=" + agentName,
		image,
	}
	// #nosec G204 -- acceptance test intentionally launches the docker CLI with harness-provided values.
	out, err := exec.CommandContext(runCtx, "docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("start hawser agent container %q: %w (output=%s)", containerName, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func testAccTailHawserLogs() string {
	dockerHost := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_DOCKER_HOST"))
	if dockerHost == "" {
		dockerHost = "tcp://127.0.0.1:23750"
	}
	containerName := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_HAWSER_CONTAINER"))
	if containerName == "" {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// #nosec G204 -- acceptance test intentionally launches the docker CLI with harness-provided values.
	out, err := exec.CommandContext(ctx, "docker", "--host", dockerHost, "logs", "--tail", "20", containerName).CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	return strings.TrimSpace(string(out))
}
