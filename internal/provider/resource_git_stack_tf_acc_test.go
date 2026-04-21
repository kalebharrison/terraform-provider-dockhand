package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGitStackResourceDestroyRemovesRuntimeTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()
	repoURL := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_REPO_URL"))
	composePath := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH"))
	if repoURL == "" || composePath == "" {
		t.Skip("acceptance test requires DOCKHAND_TEST_GIT_STACK_REPO_URL and DOCKHAND_TEST_GIT_STACK_COMPOSE_PATH")
	}

	branch := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_BRANCH"))
	if branch == "" {
		branch = "main"
	}
	credentialID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_STACK_CREDENTIAL_ID"))

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	stackName := "tf-acc-git-stack-" + strings.ToLower(time.Now().UTC().Format("20060102150405"))
	resourceName := "dockhand_git_stack.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		CheckDestroy:             testAccCheckGitStackDestroyed(endpoint, username, password),
		Steps: []resource.TestStep{
			{
				Config: testAccGitStackConfig(defaultEnv, stackName, repoURL, branch, composePath, credentialID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env", defaultEnv),
					resource.TestCheckResourceAttr(resourceName, "stack_name", stackName),
					resource.TestCheckResourceAttr(resourceName, "compose_path", composePath),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testAccCheckRuntimeStackExists(endpoint, username, password, defaultEnv, stackName),
				),
			},
		},
	})
}

func testAccGitStackConfig(env string, stackName string, repoURL string, branch string, composePath string, credentialID string) string {
	credentialBlock := ""
	if strings.TrimSpace(credentialID) != "" {
		credentialBlock = fmt.Sprintf("\n  credential_id = %q", credentialID)
	}

	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_git_stack" "test" {
  env             = %q
  stack_name      = %q
  url             = %q
  branch          = %q
  compose_path    = %q
  deploy_now      = true
  build_on_deploy = true
  repull_images   = false
  force_redeploy  = false%s
}
`, env, stackName, repoURL, branch, composePath, credentialBlock)
}

func testAccCheckRuntimeStackExists(endpoint string, username string, password string, env string, stackName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
		if err != nil {
			return err
		}

		client, err := NewClient(endpoint, sessionCookie, env, true)
		if err != nil {
			return err
		}

		deadline := time.Now().Add(90 * time.Second)
		for {
			_, found, err := client.GetStackByName(context.Background(), env, stackName)
			if err == nil && found {
				return nil
			}
			if time.Now().After(deadline) {
				if err != nil {
					return fmt.Errorf("check runtime stack existence for %q: %w", stackName, err)
				}
				return fmt.Errorf("runtime stack %q was not observed after deploy", stackName)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func testAccCheckGitStackDestroyed(endpoint string, username string, password string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		sessionCookie, err := testAccLoginSessionCookieForDestroy(endpoint, username, password)
		if err != nil {
			return err
		}

		client, err := NewClient(endpoint, sessionCookie, "", true)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "dockhand_git_stack" {
				continue
			}

			env := strings.TrimSpace(rs.Primary.Attributes["env"])
			if env == "" {
				env = testAccDefaultEnv()
			}
			stackName := strings.TrimSpace(rs.Primary.Attributes["stack_name"])

			item, _, err := client.GetGitStackByID(context.Background(), env, rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("check git stack destroy: %w", err)
			}
			if item != nil {
				return fmt.Errorf("git stack still exists: id=%s", rs.Primary.ID)
			}

			_, found, err := client.GetStackByName(context.Background(), env, stackName)
			if err != nil {
				return fmt.Errorf("check runtime stack destroy: %w", err)
			}
			if found {
				return fmt.Errorf("runtime stack still exists: env=%s name=%s", env, stackName)
			}
		}

		return nil
	}
}
