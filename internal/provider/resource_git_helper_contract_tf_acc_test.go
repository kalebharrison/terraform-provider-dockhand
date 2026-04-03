package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGitHelperSurfacesTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	repoURL := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_HELPER_REPO_URL"))
	if repoURL == "" {
		repoURL = "https://github.com/docker/awesome-compose.git"
	}
	branch := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_HELPER_BRANCH"))
	if branch == "" {
		branch = "master"
	}
	composePath := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_GIT_HELPER_COMPOSE_PATH"))
	if composePath == "" {
		composePath = "nginx-flask-mysql/compose.yaml"
	}

	suffix := strings.ToLower(time.Now().UTC().Format("20060102150405"))
	repoName := "tf-acc-git-helper-" + suffix

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGitHelperSurfacesConfig(repoName, repoURL, branch, composePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_repository_test_action.inline", "success", "true"),
					resource.TestCheckResourceAttr("dockhand_git_repository_test_action.inline", "resolved_branch", branch),
					resource.TestCheckResourceAttrSet("dockhand_git_repository_test_action.inline", "last_commit"),
					resource.TestCheckResourceAttr("data.dockhand_git_preview_env.inline", "compose_path", composePath),
					resource.TestCheckResourceAttrSet("data.dockhand_git_preview_env.inline", "vars_json"),
					resource.TestCheckResourceAttrSet("data.dockhand_git_preview_env.inline", "sources_json"),
					resource.TestCheckResourceAttr("dockhand_git_repository_test_action.saved", "success", "true"),
					resource.TestCheckResourceAttr("dockhand_git_repository_test_action.saved", "resolved_branch", branch),
					resource.TestCheckResourceAttrSet("dockhand_git_repository_test_action.saved", "last_commit"),
					resource.TestCheckResourceAttr("data.dockhand_git_preview_env.saved", "compose_path", composePath),
					resource.TestCheckResourceAttrSet("data.dockhand_git_preview_env.saved", "vars_json"),
					resource.TestCheckResourceAttrSet("data.dockhand_git_preview_env.saved", "sources_json"),
				),
			},
		},
	})
}

func testAccGitHelperSurfacesConfig(repoName string, repoURL string, branch string, composePath string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_git_repository" "saved" {
  name   = %q
  url    = %q
  branch = %q
}

resource "dockhand_git_repository_test_action" "inline" {
  url           = %q
  branch        = %q
  fail_on_error = true
  trigger       = "inline"
}

data "dockhand_git_preview_env" "inline" {
  url          = %q
  branch       = %q
  compose_path = %q
}

resource "dockhand_git_repository_test_action" "saved" {
  repository_id = dockhand_git_repository.saved.id
  fail_on_error = true
  trigger       = "saved"
}

data "dockhand_git_preview_env" "saved" {
  repository_id = dockhand_git_repository.saved.id
  compose_path  = %q
}
`, repoName, repoURL, branch, repoURL, branch, repoURL, branch, composePath, composePath)
}
