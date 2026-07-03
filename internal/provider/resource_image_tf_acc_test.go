package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccImageResourceImportTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	endpoint, username, password := testAccEnv(t)
	env := testAccDefaultEnv()
	imageName := "busybox:1.36.1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		CheckDestroy:             testAccCheckImageDestroyed(endpoint, username, password, env),
		Steps: []resource.TestStep{
			{
				Config: testAccImageResourceConfig(env, imageName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_image.test", "name", imageName),
					resource.TestCheckResourceAttrSet("dockhand_image.test", "id"),
				),
			},
			{
				ResourceName:      "dockhand_image.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccImageResourceConfig(env string, imageName string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_image" "test" {
  env             = %q
  name            = %q
  scan_after_pull = false
}
`, env, imageName)
}

func testAccCheckImageDestroyed(endpoint string, username string, password string, env string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		client, err := testAccDestroyClient(endpoint, username, password)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "dockhand_image" {
				continue
			}
			images, status, err := client.ListImages(context.Background(), env)
			if err != nil {
				return fmt.Errorf("list images after destroy: %w", err)
			}
			if status >= 400 {
				return fmt.Errorf("list images after destroy: status=%d", status)
			}
			for _, img := range images {
				if img.ID == rs.Primary.ID {
					return fmt.Errorf("image still exists: id=%s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
