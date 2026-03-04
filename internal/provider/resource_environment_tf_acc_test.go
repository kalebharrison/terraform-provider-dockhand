package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
