package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSettingsSingletonsTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	endpoint, username, password := testAccEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSettingsSingletonsConfig(
					"24h",
					"YYYY-MM-DD",
					true,
					true,
					"UTC",
					7200,
					stringPtr("/tmp/dockhand-acc-primary-a"),
					[]string{"/tmp/dockhand-acc-ext-a", "/tmp/dockhand-acc-ext-b"},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "id", "general"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "time_format", "24h"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "date_format", "YYYY-MM-DD"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "show_stopped_containers", "true"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "highlight_updates", "true"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "default_timezone", "UTC"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "primary_stack_location", "/tmp/dockhand-acc-primary-a"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "external_stack_paths.#", "2"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "external_stack_paths.0", "/tmp/dockhand-acc-ext-a"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "external_stack_paths.1", "/tmp/dockhand-acc-ext-b"),
					resource.TestCheckResourceAttr("dockhand_auth_settings.test", "id", "auth"),
					resource.TestCheckResourceAttr("dockhand_auth_settings.test", "auth_enabled", "true"),
					resource.TestCheckResourceAttr("dockhand_auth_settings.test", "default_provider", "local"),
					resource.TestCheckResourceAttr("dockhand_auth_settings.test", "session_timeout", "7200"),
					resource.TestCheckResourceAttr("dockhand_license.test", "id", "license"),
					resource.TestCheckResourceAttrSet("dockhand_license.test", "valid"),
					resource.TestCheckResourceAttrSet("dockhand_license.test", "active"),
					resource.TestCheckOutput("health_ok", "true"),
					resource.TestCheckOutput("activity_ok", "true"),
					resource.TestCheckOutput("auth_providers_ok", "true"),
				),
			},
			{
				Config: testAccSettingsSingletonsConfig(
					"12h",
					"MM/DD/YYYY",
					false,
					false,
					"America/New_York",
					14400,
					stringPtr("/tmp/dockhand-acc-primary-b"),
					[]string{"/tmp/dockhand-acc-ext-c"},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "time_format", "12h"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "date_format", "MM/DD/YYYY"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "show_stopped_containers", "false"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "highlight_updates", "false"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "default_timezone", "America/New_York"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "primary_stack_location", "/tmp/dockhand-acc-primary-b"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "external_stack_paths.#", "1"),
					resource.TestCheckResourceAttr("dockhand_settings_general.test", "external_stack_paths.0", "/tmp/dockhand-acc-ext-c"),
					resource.TestCheckResourceAttr("dockhand_auth_settings.test", "session_timeout", "14400"),
					resource.TestCheckOutput("health_ok", "true"),
					resource.TestCheckOutput("activity_ok", "true"),
					resource.TestCheckOutput("auth_providers_ok", "true"),
				),
			},
		},
	})
	_ = endpoint
	_ = username
	_ = password
}

func TestAccNotificationAndConfigSetResourcesTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	endpoint, username, password := testAccEnv(t)
	suffix := time.Now().UTC().Format("20060102150405")
	notificationName := "tf-acc-notification-" + suffix
	configSetName := "tf-acc-config-set-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		CheckDestroy:             testAccCheckNotificationAndConfigSetDestroyed(endpoint, username, password),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAndConfigSetConfig(
					notificationName,
					true,
					"smtp1.example.invalid",
					"ops1@example.invalid",
					"user1",
					"pass1",
					configSetName,
					"Config set 1",
					"UTC",
					"terraform-a",
					18080,
					true,
					"no",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_notification.test", "name", notificationName),
					resource.TestCheckResourceAttr("dockhand_notification.test", "type", "smtp"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_host", "smtp1.example.invalid"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_from_email", "dockhand@example.invalid"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_to_emails.0", "ops1@example.invalid"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_username", "user1"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_skip_tls_verify", "true"),
					resource.TestCheckResourceAttrSet("dockhand_notification.test", "id"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "name", configSetName),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "description", "Config set 1"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "env_vars.TZ", "UTC"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "labels.com.example.owner", "terraform-a"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "ports.0.host_port", "18080"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "volumes.0.read_only", "true"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "restart_policy", "no"),
					resource.TestCheckResourceAttrSet("dockhand_config_set.test", "id"),
					resource.TestCheckOutput("notifications_data_source_ok", "true"),
					resource.TestCheckOutput("config_sets_data_source_ok", "true"),
				),
			},
			{
				Config: testAccNotificationAndConfigSetConfig(
					notificationName,
					false,
					"smtp2.example.invalid",
					"ops2@example.invalid",
					"user2",
					"pass2",
					configSetName,
					"Config set 2",
					"America/New_York",
					"terraform-b",
					18081,
					false,
					"unless-stopped",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_notification.test", "enabled", "false"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_host", "smtp2.example.invalid"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_to_emails.0", "ops2@example.invalid"),
					resource.TestCheckResourceAttr("dockhand_notification.test", "smtp_username", "user2"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "description", "Config set 2"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "env_vars.TZ", "America/New_York"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "labels.com.example.owner", "terraform-b"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "ports.0.host_port", "18081"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "volumes.0.read_only", "false"),
					resource.TestCheckResourceAttr("dockhand_config_set.test", "restart_policy", "unless-stopped"),
					resource.TestCheckOutput("notifications_data_source_ok", "true"),
					resource.TestCheckOutput("config_sets_data_source_ok", "true"),
				),
			},
			{
				ResourceName:            "dockhand_notification.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"smtp_password", "created_at", "updated_at"},
			},
			{
				ResourceName:            "dockhand_config_set.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
		},
	})
}

func TestAccRegistryAndGitCredentialResourcesTerraform(t *testing.T) {
	testAccConfigureProviderEnv(t)
	endpoint, username, password := testAccEnv(t)
	suffix := time.Now().UTC().Format("20060102150405")
	registryName := "tf-acc-registry-" + suffix
	credentialName := "tf-acc-git-credential-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		CheckDestroy:             testAccCheckRegistryAndGitCredentialDestroyed(endpoint, username, password),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAndGitCredentialConfig(
					registryName,
					"https://registry-acc-1.example.invalid",
					credentialName,
					"git-user-1",
					"git-pass-1",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_registry.test", "name", registryName),
					resource.TestCheckResourceAttr("dockhand_registry.test", "url", "https://registry-acc-1.example.invalid"),
					resource.TestCheckResourceAttr("dockhand_registry.test", "is_default", "false"),
					resource.TestCheckResourceAttr("dockhand_registry.test", "has_credentials", "false"),
					resource.TestCheckResourceAttrSet("dockhand_registry.test", "id"),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "name", credentialName),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "auth_type", "password"),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "username", "git-user-1"),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "has_password", "true"),
					resource.TestCheckResourceAttrSet("dockhand_git_credential.test", "id"),
					resource.TestCheckOutput("registries_data_source_ok", "true"),
					resource.TestCheckOutput("git_credentials_data_source_ok", "true"),
				),
			},
			{
				Config: testAccRegistryAndGitCredentialConfig(
					registryName,
					"https://registry-acc-2.example.invalid",
					credentialName,
					"git-user-2",
					"git-pass-2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_registry.test", "url", "https://registry-acc-2.example.invalid"),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "username", "git-user-2"),
					resource.TestCheckResourceAttr("dockhand_git_credential.test", "has_password", "true"),
					resource.TestCheckOutput("registries_data_source_ok", "true"),
					resource.TestCheckOutput("git_credentials_data_source_ok", "true"),
				),
			},
			{
				ResourceName:            "dockhand_registry.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				ResourceName:            "dockhand_git_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "ssh_key", "created_at", "updated_at"},
			},
		},
	})
}

func TestAccNetworkAndVolumeResourcesTerraform(t *testing.T) {
	defaultEnv := testAccConfigureProviderEnv(t)
	endpoint, username, password := testAccEnv(t)
	suffix := time.Now().UTC().Format("20060102150405")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		CheckDestroy:             testAccCheckNetworkAndVolumeDestroyed(endpoint, username, password),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkAndVolumeConfig(
					defaultEnv,
					"tf-acc-network-"+suffix+"-a",
					false,
					true,
					"tf-acc-volume-"+suffix+"-a",
					"terraform-a",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_network.test", "env", defaultEnv),
					resource.TestCheckResourceAttr("dockhand_network.test", "name", "tf-acc-network-"+suffix+"-a"),
					resource.TestCheckResourceAttr("dockhand_network.test", "driver", "bridge"),
					resource.TestCheckResourceAttr("dockhand_network.test", "internal", "false"),
					resource.TestCheckResourceAttr("dockhand_network.test", "attachable", "true"),
					resource.TestCheckResourceAttrSet("dockhand_network.test", "id"),
					resource.TestCheckResourceAttr("dockhand_volume.test", "env", defaultEnv),
					resource.TestCheckResourceAttr("dockhand_volume.test", "name", "tf-acc-volume-"+suffix+"-a"),
					resource.TestCheckResourceAttr("dockhand_volume.test", "driver", "local"),
					resource.TestCheckResourceAttr("dockhand_volume.test", "labels.com.example.owner", "terraform-a"),
					resource.TestCheckResourceAttrSet("dockhand_volume.test", "id"),
					resource.TestCheckOutput("networks_data_source_ok", "true"),
					resource.TestCheckOutput("volumes_data_source_ok", "true"),
				),
			},
			{
				Config: testAccNetworkAndVolumeConfig(
					defaultEnv,
					"tf-acc-network-"+suffix+"-b",
					true,
					false,
					"tf-acc-volume-"+suffix+"-b",
					"terraform-b",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_network.test", "name", "tf-acc-network-"+suffix+"-b"),
					resource.TestCheckResourceAttr("dockhand_network.test", "internal", "true"),
					resource.TestCheckResourceAttr("dockhand_network.test", "attachable", "false"),
					resource.TestCheckResourceAttr("dockhand_volume.test", "name", "tf-acc-volume-"+suffix+"-b"),
					resource.TestCheckResourceAttr("dockhand_volume.test", "labels.com.example.owner", "terraform-b"),
					resource.TestCheckOutput("networks_data_source_ok", "true"),
					resource.TestCheckOutput("volumes_data_source_ok", "true"),
				),
			},
		},
	})
}

func testAccCheckNotificationAndConfigSetDestroyed(endpoint string, username string, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		client, err := testAccDestroyClient(endpoint, username, password)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			switch rs.Type {
			case "dockhand_notification":
				_, status, err := client.GetNotification(context.Background(), rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("notification still exists: id=%s", rs.Primary.ID)
				}
				if status != 404 {
					return fmt.Errorf("unexpected status checking notification destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			case "dockhand_config_set":
				_, status, err := client.GetConfigSet(context.Background(), rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("config set still exists: id=%s", rs.Primary.ID)
				}
				if status != 404 {
					return fmt.Errorf("unexpected status checking config set destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			}
		}

		return nil
	}
}

func testAccCheckRegistryAndGitCredentialDestroyed(endpoint string, username string, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		client, err := testAccDestroyClient(endpoint, username, password)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			switch rs.Type {
			case "dockhand_registry":
				_, status, err := client.GetRegistry(context.Background(), rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("registry still exists: id=%s", rs.Primary.ID)
				}
				if status != 404 {
					return fmt.Errorf("unexpected status checking registry destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			case "dockhand_git_credential":
				_, status, err := client.GetGitCredential(context.Background(), rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("git credential still exists: id=%s", rs.Primary.ID)
				}
				if status != 404 {
					return fmt.Errorf("unexpected status checking git credential destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			}
		}

		return nil
	}
}

func testAccCheckNetworkAndVolumeDestroyed(endpoint string, username string, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		client, err := testAccDestroyClient(endpoint, username, password)
		if err != nil {
			return err
		}

		env := testAccDefaultEnv()
		for _, rs := range state.RootModule().Resources {
			switch rs.Type {
			case "dockhand_network":
				_, status, err := client.GetNetworkInspect(context.Background(), env, rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("network still exists: id=%s", rs.Primary.ID)
				}
				if status != 404 && !isMissingNetworkInspectError(status, err) {
					return fmt.Errorf("unexpected status checking network destroy: id=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			case "dockhand_volume":
				_, status, err := client.GetVolumeInspect(context.Background(), env, rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("volume still exists: name=%s", rs.Primary.ID)
				}
				if status != 404 && !isMissingVolumeInspectError(status, err) {
					return fmt.Errorf("unexpected status checking volume destroy: name=%s status=%d err=%v", rs.Primary.ID, status, err)
				}
			}
		}

		return nil
	}
}

func testAccDestroyClient(endpoint string, username string, password string) (*Client, error) {
	sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
	if err != nil {
		return nil, err
	}
	return NewClient(endpoint, sessionCookie, testAccDefaultEnv(), true)
}

func testAccSettingsSingletonsConfig(timeFormat string, dateFormat string, showStopped bool, highlight bool, timezone string, sessionTimeout int64, primaryStackLocation *string, externalStackPaths []string) string {
	primaryLit := "null"
	if primaryStackLocation != nil {
		primaryLit = fmt.Sprintf("%q", *primaryStackLocation)
	}

	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_settings_general" "test" {
  time_format             = %q
  date_format             = %q
  show_stopped_containers = %t
  highlight_updates       = %t
  default_timezone        = %q
  primary_stack_location  = %s
  external_stack_paths    = %s
}

resource "dockhand_auth_settings" "test" {
  auth_enabled     = true
  default_provider = "local"
  session_timeout  = %d
}

resource "dockhand_license" "test" {}

data "dockhand_health" "test" {}

data "dockhand_activity" "test" {
  limit = 5
}

data "dockhand_auth_providers" "test" {}

output "health_ok" {
  value = data.dockhand_health.test.id == "dockhand-health" && data.dockhand_health.test.checked_at != "" && data.dockhand_health.test.status != ""
}

output "activity_ok" {
  value = data.dockhand_activity.test.id == "dockhand-activity" && data.dockhand_activity.test.limit == 5
}

output "auth_providers_ok" {
  value = data.dockhand_auth_providers.test.id == "auth-providers" && data.dockhand_auth_providers.test.default_provider == dockhand_auth_settings.test.default_provider && contains([for p in data.dockhand_auth_providers.test.providers : p.id], "local")
}
`, timeFormat, dateFormat, showStopped, highlight, timezone, primaryLit, terraformStringList(externalStackPaths), sessionTimeout)
}

func testAccNotificationAndConfigSetConfig(notificationName string, notificationEnabled bool, smtpHost string, recipient string, smtpUsername string, smtpPassword string, configSetName string, description string, timezone string, owner string, hostPort int, readOnly bool, restartPolicy string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_notification" "test" {
  name                  = %q
  type                  = "smtp"
  enabled               = %t
  smtp_host             = %q
  smtp_port             = 25
  smtp_from_email       = "dockhand@example.invalid"
  smtp_to_emails        = [%q]
  smtp_username         = %q
  smtp_password         = %q
  smtp_use_tls          = false
  smtp_starttls         = false
  smtp_skip_tls_verify  = true
}

resource "dockhand_config_set" "test" {
  name        = %q
  description = %q

  env_vars = {
    TZ = %q
  }

  labels = {
    "com.example.owner" = %q
  }

  ports = [
    {
      container_port = 80
      host_port      = %d
      protocol       = "tcp"
    }
  ]

  volumes = [
    {
      source    = "/tmp"
      target    = "/data"
      type      = "bind"
      read_only = %t
    }
  ]

  network_mode   = "bridge"
  restart_policy = %q
}

data "dockhand_notifications" "test" {}
data "dockhand_config_sets" "test" {}

output "notifications_data_source_ok" {
  value = try(length(data.dockhand_notifications.test.ids) >= 0, false)
}

output "config_sets_data_source_ok" {
  value = try(length(data.dockhand_config_sets.test.ids) >= 0, false)
}
`, notificationName, notificationEnabled, smtpHost, recipient, smtpUsername, smtpPassword, configSetName, description, timezone, owner, hostPort, readOnly, restartPolicy)
}

func testAccRegistryAndGitCredentialConfig(registryName string, registryURL string, credentialName string, gitUsername string, gitPassword string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_registry" "test" {
  name       = %q
  url        = %q
  is_default = false
}

resource "dockhand_git_credential" "test" {
  name      = %q
  auth_type = "password"
  username  = %q
  password  = %q
}

data "dockhand_registries" "test" {}
data "dockhand_git_credentials" "test" {}

output "registries_data_source_ok" {
  value = try(length(data.dockhand_registries.test.ids) >= 0, false)
}

output "git_credentials_data_source_ok" {
  value = try(length(data.dockhand_git_credentials.test.ids) >= 0, false)
}
`, registryName, registryURL, credentialName, gitUsername, gitPassword)
}

func testAccNetworkAndVolumeConfig(env string, networkName string, internal bool, attachable bool, volumeName string, owner string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_network" "test" {
  env        = %q
  name       = %q
  driver     = "bridge"
  internal   = %t
  attachable = %t
}

resource "dockhand_volume" "test" {
  env   = %q
  name  = %q
  driver = "local"
  labels = {
    "com.example.owner" = %q
  }
}

data "dockhand_networks" "test" {
  env = %q
}

data "dockhand_volumes" "test" {
  env = %q
}

output "networks_data_source_ok" {
  value = try(length(data.dockhand_networks.test.ids) >= 0, false)
}

output "volumes_data_source_ok" {
  value = try(length(data.dockhand_volumes.test.names) >= 0, false)
}
`, env, networkName, internal, attachable, env, volumeName, owner, env, env)
}

func terraformStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func stringPtr(v string) *string {
	return &v
}
