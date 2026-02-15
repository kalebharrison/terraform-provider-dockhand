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

func TestAccContainerFileResourceTerraform(t *testing.T) {
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
	path := fmt.Sprintf("/tmp/tf-acc-file-%s.txt", suffix)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"dockhand": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccContainerFileConfig(defaultEnv, containerID, path, "hello-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_file.test", "path", path),
					resource.TestCheckResourceAttr("dockhand_container_file.test", "content", "hello-1"),
					resource.TestCheckResourceAttrSet("dockhand_container_file.test", "id"),
				),
			},
			{
				Config: testAccContainerFileConfig(defaultEnv, containerID, path, "hello-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_file.test", "content", "hello-2"),
				),
			},
		},
	})
}

func TestAccGitStackDeployActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	stackID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_ID"))
	if stackID == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_GIT_STACK_ID")
	}

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"dockhand": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGitStackDeployActionConfig(stackID, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack_deploy_action.test", "stack_id", stackID),
					resource.TestCheckResourceAttr("dockhand_git_stack_deploy_action.test", "result", "deploy_requested"),
					resource.TestCheckResourceAttrSet("dockhand_git_stack_deploy_action.test", "id"),
				),
			},
			{
				Config: testAccGitStackDeployActionConfig(stackID, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack_deploy_action.test", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func testAccContainerFileConfig(env string, containerID string, path string, content string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_container_file" "test" {
  env          = %q
  container_id = %q
  path         = %q
  content      = %q
}
`, env, containerID, path, content)
}

func testAccGitStackDeployActionConfig(stackID string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_git_stack_deploy_action" "test" {
  stack_id = %q
  trigger  = %q
}
`, stackID, trigger)
}
