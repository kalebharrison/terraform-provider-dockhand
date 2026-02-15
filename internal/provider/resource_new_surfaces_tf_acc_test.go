package provider

import (
	"fmt"
	"os"
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
