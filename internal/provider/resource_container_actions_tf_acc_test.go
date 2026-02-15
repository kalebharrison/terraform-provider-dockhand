package provider

import (
	"context"
	"fmt"
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

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	containerID, cleanup := testAccCreateContainerFixture(t, endpoint, username, password, defaultEnv)
	defer cleanup()

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
  enabled = true
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
  restart_policy_name = "no"
  payload_json = jsonencode({})
  trigger      = %q
}
`, env, containerID, trigger)
}

func testAccCreateContainerFixture(t *testing.T, endpoint string, username string, password string, env string) (string, func()) {
	t.Helper()

	ctx := context.Background()
	sessionCookie := testAccLoginSessionCookie(t, endpoint, username, password)
	client, err := NewClient(endpoint, sessionCookie, env, false)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	pullStatus, pullErr := client.PullImage(ctx, env, "nginx:latest", false)
	if pullErr != nil || pullStatus < 200 || pullStatus > 299 {
		t.Fatalf("pull image failed status=%d err=%v", pullStatus, pullErr)
	}

	name := "tf-acc-update-fixture-" + strings.ToLower(time.Now().UTC().Format("20060102150405"))
	created, createStatus, createErr := client.CreateContainer(ctx, env, containerPayload{
		Name:  name,
		Image: "nginx:latest",
	})
	if createErr != nil || createStatus < 200 || createStatus > 299 || created == nil || strings.TrimSpace(created.ID) == "" {
		t.Fatalf("create fixture container failed status=%d err=%v", createStatus, createErr)
	}
	id := created.ID

	if _, startErr := client.StartContainer(ctx, env, id); startErr != nil {
		t.Fatalf("start fixture container failed: %v", startErr)
	}

	cleanup := func() {
		_, _ = client.DeleteContainer(ctx, env, id)
	}
	return id, cleanup
}
