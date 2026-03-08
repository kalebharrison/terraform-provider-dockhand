package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"dockhand": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestAccContainerFileDirectoryResourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	containerID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_FILE_CONTAINER_ID"))
	if containerID == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_FILE_CONTAINER_ID")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	path := fmt.Sprintf("/tmp/tf-acc-dir-%s", suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDirectoryConfig(defaultEnv, containerID, path),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_file.test", "path", path),
					resource.TestCheckResourceAttr("dockhand_container_file.test", "type", "directory"),
					resource.TestCheckNoResourceAttr("dockhand_container_file.test", "content"),
				),
			},
		},
	})
}

func TestAccContainerProcessesDataSourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	containerID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_FILE_CONTAINER_ID"))
	if containerID == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_FILE_CONTAINER_ID")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerProcessesConfig(defaultEnv, containerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockhand_container_processes.test", "container_id", containerID),
					resource.TestCheckResourceAttrSet("data.dockhand_container_processes.test", "id"),
				),
			},
		},
	})
}

func TestAccStackActionDownTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	stackName := "tf-acc-down-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStackActionDownConfig(defaultEnv, stackName, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_action.down", "action", "down"),
					resource.TestCheckResourceAttrSet("dockhand_stack_action.down", "id"),
				),
			},
			{
				Config: testAccStackActionDownConfig(defaultEnv, stackName, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_action.down", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func TestAccStackEnvResourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	stackName := "tf-acc-env-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStackEnvConfig(defaultEnv, stackName, "API_KEY=abc\n", "TOKEN", "secret-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "stack_name", stackName),
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "raw_content", "API_KEY=abc\n"),
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "secret_variables.0.key", "TOKEN"),
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "secret_variables.0.value", "secret-1"),
				),
			},
			{
				Config: testAccStackEnvConfig(defaultEnv, stackName, "API_KEY=xyz\n", "TOKEN", "secret-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "raw_content", "API_KEY=xyz\n"),
					resource.TestCheckResourceAttr("dockhand_stack_env.test", "secret_variables.0.value", "secret-2"),
				),
			},
		},
	})
}

func TestAccGitStackEnvFileResourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	stackID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_ID"))
	envPath := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_ENV_PATH"))
	if stackID == "" || envPath == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_GIT_STACK_ID and DOCKHAND_TEST_GIT_STACK_ENV_PATH")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGitStackEnvFileConfig(stackID, envPath, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack_env_file.test", "stack_id", stackID),
					resource.TestCheckResourceAttr("dockhand_git_stack_env_file.test", "path", envPath),
					resource.TestCheckResourceAttrSet("dockhand_git_stack_env_file.test", "vars_json"),
				),
			},
			{
				Config: testAccGitStackEnvFileConfig(stackID, envPath, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack_env_file.test", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func TestAccGitRepositoryEnvironmentIDNoDriftTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	testEnvID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_REPO_ENV_ID"))
	if testEnvID == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_GIT_REPO_ENV_ID")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	repoName := "tf-acc-git-repo-" + suffix
	repoURL := "https://github.com/docker-library/hello-world.git"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGitRepositoryConfig(repoName, repoURL, "main", testEnvID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_repository.test", "name", repoName),
					resource.TestCheckResourceAttr("dockhand_git_repository.test", "environment_id", testEnvID),
					resource.TestCheckResourceAttrSet("dockhand_git_repository.test", "id"),
				),
			},
			{
				ResourceName:            "dockhand_git_repository.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_secret", "last_sync", "last_commit", "sync_status", "sync_error", "created_at", "updated_at"},
			},
			{
				Config:             testAccGitRepositoryConfig(repoName, repoURL, "main", testEnvID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccBatchActionAndJobDataSourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	containerName := "tf-acc-batch-target-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccBatchActionConfig(defaultEnv, containerName, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "entity_type", "containers"),
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "operation", "restart"),
					resource.TestMatchResourceAttr("dockhand_batch_action.test", "job_status", regexp.MustCompile("(?i)^(done|queued|pending|running|processing|failed|error|cancelled|canceled)$")),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "result_json"),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "lines_json"),
				),
			},
			{
				Config: testAccBatchActionConfig(defaultEnv, containerName, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "trigger", "acc-run-2"),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "result_json"),
				),
			},
		},
	})
}

func TestAccBatchActionNoWaitTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	containerName := "tf-acc-no-wait-target-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccBatchActionNoWaitConfig(defaultEnv, containerName, "acc-no-wait-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "entity_type", "containers"),
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "operation", "restart"),
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "wait_for_completion", "false"),
					resource.TestMatchResourceAttr("dockhand_batch_action.test", "job_status", regexp.MustCompile("(?i)^(submitted|queued|pending|running|processing|done|failed|error|cancelled|canceled)$")),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "result_json"),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "lines_json"),
				),
			},
			{
				Config: testAccBatchActionNoWaitConfig(defaultEnv, containerName, "acc-no-wait-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_batch_action.test", "trigger", "acc-no-wait-2"),
					resource.TestCheckResourceAttrSet("dockhand_batch_action.test", "result_json"),
				),
			},
		},
	})
}

func TestAccJobDataSourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
	if err != nil {
		t.Fatalf("login for job data source fixture: %v", err)
	}
	client, err := NewClient(endpoint, sessionCookie, defaultEnv, false)
	if err != nil {
		t.Fatalf("new client for job data source fixture: %v", err)
	}

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	submitted, _, err := client.SubmitBatch(context.Background(), defaultEnv, "containers", "restart", []string{"tf-acc-job-fixture-" + suffix})
	if err != nil {
		t.Fatalf("submit batch fixture: %v", err)
	}
	if strings.TrimSpace(submitted.JobID) == "" {
		t.Skip("Dockhand returned inline batch completion without async job id; skipping dockhand_job data source acceptance")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDataSourceConfig(submitted.JobID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dockhand_job.test", "job_id", submitted.JobID),
					resource.TestCheckResourceAttrSet("data.dockhand_job.test", "id"),
					resource.TestCheckResourceAttrSet("data.dockhand_job.test", "status"),
					resource.TestCheckResourceAttrSet("data.dockhand_job.test", "result_json"),
					resource.TestCheckResourceAttrSet("data.dockhand_job.test", "lines_json"),
				),
			},
		},
	})
}

func TestAccEnvironmentTestActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	dindHost := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_DIND_HOST"))
	if dindHost == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_DIND_HOST")
	}
	dindPort := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_DIND_PORT"))
	if dindPort == "" {
		dindPort = "2375"
	}
	if !regexp.MustCompile(`^\d+$`).MatchString(dindPort) {
		t.Skip("acceptance test requires numeric DOCKHAND_TEST_DIND_PORT")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentTestActionConfig(dindHost, dindPort, "acc-env-test-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "connection_type", "direct"),
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "host", dindHost),
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "port", dindPort),
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "protocol", "http"),
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "success", "true"),
					resource.TestCheckResourceAttrSet("dockhand_environment_test_action.test", "server_version"),
					resource.TestCheckResourceAttrSet("dockhand_environment_test_action.test", "info_json"),
				),
			},
			{
				Config: testAccEnvironmentTestActionConfig(dindHost, dindPort, "acc-env-test-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "trigger", "acc-env-test-2"),
					resource.TestCheckResourceAttr("dockhand_environment_test_action.test", "success", "true"),
				),
			},
		},
	})
}

func TestAccEnvironmentDetectSocketDataSourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDetectSocketDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dockhand_environment_detect_socket.test", "id"),
					resource.TestCheckResourceAttrSet("data.dockhand_environment_detect_socket.test", "home_dir"),
					resource.TestCheckResourceAttrSet("data.dockhand_environment_detect_socket.test", "sockets_json"),
				),
			},
		},
	})
}

func TestAccNotificationTestActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTestActionConfig("acc-notif-test-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_notification_test_action.test", "type", "smtp"),
					resource.TestCheckResourceAttr("dockhand_notification_test_action.test", "fail_on_error", "false"),
					resource.TestCheckResourceAttr("dockhand_notification_test_action.test", "success", "false"),
					resource.TestMatchResourceAttr("dockhand_notification_test_action.test", "error", regexp.MustCompile(`(?i)smtp`)),
					resource.TestCheckResourceAttrSet("dockhand_notification_test_action.test", "result_json"),
				),
			},
			{
				Config: testAccNotificationTestActionConfig("acc-notif-test-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_notification_test_action.test", "trigger", "acc-notif-test-2"),
					resource.TestCheckResourceAttr("dockhand_notification_test_action.test", "success", "false"),
				),
			},
		},
	})
}

func testAccContainerDirectoryConfig(env string, containerID string, path string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_container_file" "test" {
  env          = %q
  container_id = %q
  path         = %q
  type         = "directory"
}
`, env, containerID, path)
}

func testAccContainerProcessesConfig(env string, containerName string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

data "dockhand_container_processes" "test" {
  env          = %q
  container_id = %q
}
`, env, containerName)
}

func testAccStackActionDownConfig(env string, stackName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_stack" "test" {
  env  = %q
  name = %q
  compose = <<-YAML
services:
  app:
    image: busybox:1.36.1
    command: ["sleep", "3600"]
YAML
  enabled = true
}

resource "dockhand_stack_action" "down" {
  env        = %q
  stack_name = dockhand_stack.test.name
  action     = "down"
  trigger    = %q
}
`, env, stackName, env, trigger)
}

func testAccStackEnvConfig(env string, stackName string, raw string, key string, value string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_stack" "test" {
  env  = %q
  name = %q
  compose = <<-YAML
services:
  app:
    image: busybox:1.36.1
    command: ["sleep", "3600"]
YAML
  enabled = true
}

resource "dockhand_stack_env" "test" {
  env        = %q
  stack_name = dockhand_stack.test.name
  raw_content = %q
  secret_variables = [
    {
      key       = %q
      value     = %q
      is_secret = true
    }
  ]
}
`, env, stackName, env, raw, key, value)
}

func testAccGitStackEnvFileConfig(stackID string, path string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_git_stack_env_file" "test" {
  stack_id = %q
  path     = %q
  trigger  = %q
}
`, stackID, path, trigger)
}

func testAccGitRepositoryConfig(name string, url string, branch string, environmentID string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_git_repository" "test" {
  name           = %q
  url            = %q
  branch         = %q
  environment_id = %q
}
`, name, url, branch, environmentID)
}

func testAccBatchActionConfig(env string, containerName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_image" "batch_target" {
  env             = %q
  name            = "busybox:1.36.1"
  scan_after_pull = false
}

resource "dockhand_container" "batch_target" {
  env     = %q
  name    = %q
  image   = dockhand_image.batch_target.name
  enabled = false
}

resource "dockhand_batch_action" "test" {
  env                 = %q
  entity_type         = "containers"
  operation           = "restart"
  item_ids            = [dockhand_container.batch_target.id]
  wait_for_completion = true
  timeout_seconds     = 30
  poll_interval_ms    = 500
  trigger             = %q
}
	`, env, env, containerName, env, trigger)
}

func testAccBatchActionNoWaitConfig(env string, containerName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_image" "batch_target" {
  env             = %q
  name            = "busybox:1.36.1"
  scan_after_pull = false
}

resource "dockhand_container" "batch_target" {
  env     = %q
  name    = %q
  image   = dockhand_image.batch_target.name
  enabled = false
}

resource "dockhand_batch_action" "test" {
  env                 = %q
  entity_type         = "containers"
  operation           = "restart"
  item_ids            = [dockhand_container.batch_target.id]
  wait_for_completion = false
  trigger             = %q
}
`, env, env, containerName, env, trigger)
}

func testAccEnvironmentTestActionConfig(host string, port string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_environment_test_action" "test" {
  connection_type = "direct"
  host            = %q
  port            = %s
  protocol        = "http"
  tls_skip_verify = false
  fail_on_error   = true
  trigger         = %q
}
`, host, port, trigger)
}

func testAccEnvironmentDetectSocketDataSourceConfig() string {
	return `
provider "dockhand" {}

data "dockhand_environment_detect_socket" "test" {}
`
}

func testAccNotificationTestActionConfig(trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_notification_test_action" "test" {
  type = "smtp"
  config_json = jsonencode({
    host       = "smtp.example.local"
    port       = 25
    from_email = "dockhand@example.local"
    to_emails  = ["ops@example.local"]
  })
  fail_on_error = false
  trigger       = %q
}
`, trigger)
}

func testAccJobDataSourceConfig(jobID string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

data "dockhand_job" "test" {
  job_id = %q
}
`, jobID)
}
