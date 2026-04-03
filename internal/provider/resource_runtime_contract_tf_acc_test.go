package provider

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccContainerRuntimeSurfacesTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	env := testAccDefaultEnv()
	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	imageName := "busybox:1.36.1"
	containerName := "tf-acc-runtime-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRuntimeSurfacesConfig(env, imageName, containerName, "TF_ACC_RUNTIME_STOP", "stop", "acc-run-1", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_image.test", "name", imageName),
					resource.TestCheckResourceAttr("dockhand_container.test", "name", containerName),
					resource.TestCheckResourceAttr("dockhand_container_action.runtime", "action", "stop"),
					resource.TestCheckResourceAttr("dockhand_container_check_updates_action.check", "trigger", "acc-run-1"),
					resource.TestCheckOutput("images_have_target", "true"),
					resource.TestCheckOutput("containers_have_target", "true"),
					resource.TestCheckOutput("pending_updates_env_matches", "true"),
					resource.TestCheckOutput("inspect_has_name", "true"),
					resource.TestCheckOutput("container_state_matches_expectation", "true"),
				),
			},
			{
				Config: testAccContainerRuntimeSurfacesConfig(env, imageName, containerName, "TF_ACC_RUNTIME_START", "start", "acc-run-2", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_container_action.runtime", "action", "start"),
					resource.TestCheckResourceAttr("dockhand_container_action.runtime", "trigger", "acc-run-2"),
					resource.TestCheckResourceAttrSet("dockhand_container_check_updates_action.check", "results_json"),
					resource.TestCheckOutput("images_have_target", "true"),
					resource.TestCheckOutput("containers_have_target", "true"),
					resource.TestCheckOutput("pending_updates_env_matches", "true"),
					resource.TestCheckOutput("inspect_has_name", "true"),
					resource.TestCheckOutput("container_state_matches_expectation", "true"),
					resource.TestCheckOutput("logs_have_marker", "true"),
					resource.TestCheckOutput("shells_have_sh", "true"),
					resource.TestCheckOutput("stats_have_target", "true"),
				),
			},
		},
	})
}

func TestAccNetworkConnectionActionTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	env := testAccDefaultEnv()
	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	containerName := "tf-acc-network-ctr-" + suffix
	networkName := "tf-acc-network-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConnectionActionConfig(env, containerName, networkName, "connect", "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_network_connection_action.test", "action", "connect"),
					resource.TestCheckOutput("inspect_contains_network", "true"),
					resource.TestCheckOutput("networks_have_target", "true"),
				),
			},
			{
				Config: testAccNetworkConnectionActionConfig(env, containerName, networkName, "disconnect", "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_network_connection_action.test", "action", "disconnect"),
					resource.TestCheckResourceAttr("dockhand_network_connection_action.test", "trigger", "acc-run-2"),
					resource.TestCheckOutput("inspect_contains_network", "false"),
					resource.TestCheckOutput("networks_have_target", "true"),
				),
			},
		},
	})
}

func TestAccVolumeCloneActionTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	env := testAccDefaultEnv()
	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	sourceName := "tf-acc-volume-src-" + suffix
	targetName := "tf-acc-volume-clone-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeCloneActionConfig(env, sourceName, targetName, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_volume_clone_action.test", "source_name", sourceName),
					resource.TestCheckResourceAttr("dockhand_volume_clone_action.test", "target_name", targetName),
					resource.TestCheckOutput("clone_present", "true"),
				),
			},
		},
	})
}

func TestAccStackInventoryAndScanTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	env := testAccDefaultEnv()
	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	stackName := "tf-acc-stack-inventory-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInventoryAndScanConfig(env, stackName, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack.test", "name", stackName),
					resource.TestCheckResourceAttr("dockhand_stack_scan_action.test", "trigger", "acc-run-1"),
					resource.TestCheckResourceAttrSet("dockhand_stack_scan_action.test", "result_json"),
					resource.TestCheckOutput("stacks_have_target", "true"),
					resource.TestCheckOutput("stack_sources_have_target", "true"),
				),
			},
			{
				Config: testAccStackInventoryAndScanConfig(env, stackName, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_stack_scan_action.test", "trigger", "acc-run-2"),
					resource.TestCheckOutput("stacks_have_target", "true"),
					resource.TestCheckOutput("stack_sources_have_target", "true"),
				),
			},
		},
	})
}

func testAccContainerRuntimeSurfacesConfig(env string, imageName string, containerName string, marker string, action string, trigger string, includeRunningData bool) string {
	runningData := ""
	if includeRunningData {
		runningData = `
data "dockhand_container_stats" "stats" {
  env = "` + env + `"

  depends_on = [
    dockhand_container_action.runtime,
  ]
}

data "dockhand_container_shells" "shells" {
  env          = "` + env + `"
  container_id = dockhand_container.test.id

  depends_on = [
    dockhand_container_action.runtime,
  ]
}

data "dockhand_container_logs" "logs" {
  env          = "` + env + `"
  container_id = dockhand_container.test.id
  tail         = 200

  depends_on = [
    dockhand_container_action.runtime,
  ]
}

output "stats_have_target" {
  value = contains([for s in data.dockhand_container_stats.stats.stats : s.id], dockhand_container.test.id)
}

output "shells_have_sh" {
  value = contains(data.dockhand_container_shells.shells.shells, "/bin/sh")
}

output "logs_have_marker" {
  value = strcontains(data.dockhand_container_logs.logs.logs, "` + marker + `")
}
`
	}

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
  command = %q
  enabled = true
  labels = {
    "com.example.owner" = "terraform-acceptance"
  }
}

resource "dockhand_container_action" "runtime" {
  env          = %q
  container_id = dockhand_container.test.id
  action       = %q
  trigger      = %q
}

resource "dockhand_container_check_updates_action" "check" {
  env     = %q
  trigger = %q
}

data "dockhand_images" "all" {
  env = %q

  depends_on = [
    dockhand_image.test,
  ]
}

data "dockhand_containers" "all" {
  env = %q

  depends_on = [
    dockhand_container_action.runtime,
  ]
}

data "dockhand_container_pending_updates" "pending" {
  env = %q

  depends_on = [
    dockhand_container_check_updates_action.check,
  ]
}

data "dockhand_container_inspect" "inspect" {
  env          = %q
  container_id = dockhand_container.test.id

  depends_on = [
    dockhand_container_action.runtime,
  ]
}

locals {
  target_container = try([for c in data.dockhand_containers.all.containers : c if c.id == dockhand_container.test.id][0], null)
  target_state     = lower(try(local.target_container.state, ""))
}

output "images_have_target" {
  value = contains(data.dockhand_images.all.ids, dockhand_image.test.id)
}

output "containers_have_target" {
  value = contains(data.dockhand_containers.all.ids, dockhand_container.test.id)
}

output "pending_updates_env_matches" {
  value = data.dockhand_container_pending_updates.pending.environment_id == %q
}

output "inspect_has_name" {
  value = strcontains(data.dockhand_container_inspect.inspect.inspect_json, dockhand_container.test.name)
}

output "container_state_matches_expectation" {
  value = contains(%s, local.target_state)
}

%s
`, env, imageName, env, containerName, "sh -c 'echo "+marker+"; sleep 3600'", env, action, trigger, env, trigger, env, env, env, env, env, expectedContainerStates(action), runningData)
}

func testAccNetworkConnectionActionConfig(env string, containerName string, networkName string, action string, trigger string) string {
	expectedPresent := action == "connect"

	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_image" "test" {
  env             = %q
  name            = "busybox:1.36.1"
  scan_after_pull = false
}

resource "dockhand_container" "test" {
  env     = %q
  name    = %q
  image   = dockhand_image.test.name
  command = "sleep 3600"
  enabled = true
}

resource "dockhand_network" "test" {
  env  = %q
  name = %q
}

resource "dockhand_network_connection_action" "test" {
  env          = %q
  network_id   = dockhand_network.test.id
  container_id = dockhand_container.test.id
  action       = %q
  trigger      = %q
}

data "dockhand_networks" "all" {
  env = %q

  depends_on = [
    dockhand_network_connection_action.test,
  ]
}

data "dockhand_container_inspect" "inspect" {
  env          = %q
  container_id = dockhand_container.test.id

  depends_on = [
    dockhand_network_connection_action.test,
  ]
}

output "networks_have_target" {
  value = contains(data.dockhand_networks.all.names, dockhand_network.test.name)
}

output "inspect_contains_network" {
  value = strcontains(data.dockhand_container_inspect.inspect.inspect_json, dockhand_network.test.name) == %t
}
`, env, env, containerName, env, networkName, env, action, trigger, env, env, expectedPresent)
}

func testAccVolumeCloneActionConfig(env string, sourceName string, targetName string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_volume" "source" {
  env  = %q
  name = %q
}

resource "dockhand_volume_clone_action" "test" {
  env         = %q
  source_name = dockhand_volume.source.name
  target_name = %q
  trigger     = %q
}

data "dockhand_volumes" "all" {
  env = %q

  depends_on = [
    dockhand_volume_clone_action.test,
  ]
}

output "clone_present" {
  value = contains(data.dockhand_volumes.all.names, %q)
}
`, env, sourceName, env, targetName, trigger, env, targetName)
}

func testAccStackInventoryAndScanConfig(env string, stackName string, trigger string) string {
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

resource "dockhand_stack_scan_action" "test" {
  trigger = %q

  depends_on = [
    dockhand_stack.test,
  ]
}

data "dockhand_stacks" "all" {
  env = %q

  depends_on = [
    dockhand_stack_scan_action.test,
  ]
}

data "dockhand_stack_sources" "all" {
  depends_on = [
    dockhand_stack_scan_action.test,
  ]
}

output "stacks_have_target" {
  value = contains([for s in data.dockhand_stacks.all.stacks : s.name], dockhand_stack.test.name)
}

output "stack_sources_have_target" {
  value = contains([for s in data.dockhand_stack_sources.all.sources : s.stack_name], dockhand_stack.test.name)
}
`, env, stackName, trigger, env)
}

func expectedContainerStates(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "stop":
		return `["created", "dead", "exited", "stopped"]`
	case "pause":
		return `["paused"]`
	default:
		return `["running"]`
	}
}
