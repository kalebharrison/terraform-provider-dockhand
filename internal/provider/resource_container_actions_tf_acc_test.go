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

func TestAccContainerRenameActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	containerName := "tf-acc-rename-" + suffix
	imageName := "busybox:1.36.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"dockhand": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRenameActionConfig(defaultEnv, imageName, containerName, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_rename_action.test", "name", containerName),
					resource.TestCheckResourceAttrSet("dockhand_container_rename_action.test", "id"),
				),
			},
			{
				Config: testAccContainerRenameActionConfig(defaultEnv, imageName, containerName, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_rename_action.test", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func TestAccContainerUpdateActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	containerID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_UPDATE_CONTAINER_ID"))
	if containerID == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_UPDATE_CONTAINER_ID")
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
				Config: testAccContainerUpdateActionConfig(defaultEnv, containerID, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_update_action.test", "payload_json", "{}"),
					resource.TestCheckResourceAttrSet("dockhand_container_update_action.test", "result_json"),
				),
			},
			{
				Config: testAccContainerUpdateActionConfig(defaultEnv, containerID, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_update_action.test", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func testAccContainerRenameActionConfig(env string, imageName string, containerName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_image" "test" {
  env             = %q
  name            = %q
  scan_after_pull = false
}

resource "dockhand_container" "test" {
  env     = %q
  name    = %q
  image   = dockhand_image.test.name
  enabled = false
}

resource "dockhand_container_rename_action" "test" {
  env          = %q
  container_id = dockhand_container.test.id
  name         = %q
  trigger      = %q
}
`, env, imageName, env, containerName, env, containerName, trigger)
}

func testAccContainerUpdateActionConfig(env string, containerID string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_container_update_action" "test" {
  env          = %q
  container_id = %q
  payload_json = jsonencode({})
  trigger      = %q
}
`, env, containerID, trigger)
}
