package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleResourceTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccConfigureProviderEnv(t)
	_ = defaultEnv

	scheduleType, scheduleID, originalEnabled := testAccScheduleFixture(t, endpoint, username, password)
	toggledEnabled := !originalEnabled

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleResourceConfig(scheduleType, scheduleID, toggledEnabled),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_schedule.test", "type", scheduleType),
					resource.TestCheckResourceAttr("dockhand_schedule.test", "schedule_id", scheduleID),
					resource.TestCheckResourceAttr("dockhand_schedule.test", "enabled", strconv.FormatBool(toggledEnabled)),
					resource.TestCheckResourceAttrSet("dockhand_schedule.test", "name"),
				),
			},
			{
				Config: testAccScheduleResourceConfig(scheduleType, scheduleID, originalEnabled),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_schedule.test", "enabled", strconv.FormatBool(originalEnabled)),
				),
			},
			{
				ResourceName:      "dockhand_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGitStackWebhookActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	testAccConfigureProviderEnv(t)
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

	stackName := "tf-acc-git-webhook-" + strings.ToLower(time.Now().UTC().Format("20060102150405"))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGitStackWebhookActionConfig(defaultEnv, stackName, repoURL, branch, composePath, credentialID, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack.test", "webhook_enabled", "true"),
					resource.TestCheckResourceAttrSet("dockhand_git_stack.test", "id"),
					resource.TestCheckResourceAttrPair("dockhand_git_stack_webhook_action.test", "stack_id", "dockhand_git_stack.test", "id"),
					testAccCheckRuntimeStackExists(endpoint, username, password, defaultEnv, stackName),
				),
			},
			{
				Config: testAccGitStackWebhookActionConfig(defaultEnv, stackName, repoURL, branch, composePath, credentialID, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_git_stack_webhook_action.test", "trigger", "acc-run-2"),
					resource.TestCheckResourceAttrSet("dockhand_git_stack.test", "id"),
				),
			},
		},
	})
}

func testAccScheduleFixture(t *testing.T, endpoint string, username string, password string) (string, string, bool) {
	t.Helper()

	sessionCookie := testAccLoginSessionCookie(t, endpoint, username, password)
	client, err := NewClient(endpoint, sessionCookie, testAccDefaultEnv(), true)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	schedules, _, err := client.GetSchedules(context.Background())
	if err != nil {
		t.Fatalf("get schedules: %v", err)
	}

	wantedType := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_SCHEDULE_TYPE"))
	wantedID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_SCHEDULE_ID"))
	if wantedType != "" && wantedID != "" {
		for _, sched := range schedules.Schedules {
			if sched.Type == wantedType && strconv.FormatInt(sched.ID, 10) == wantedID {
				return wantedType, wantedID, sched.Enabled
			}
		}
		t.Fatalf("configured acceptance schedule %s:%s was not found", wantedType, wantedID)
	}

	for _, sched := range schedules.Schedules {
		if sched.IsSystem {
			return sched.Type, strconv.FormatInt(sched.ID, 10), sched.Enabled
		}
	}

	if len(schedules.Schedules) > 0 {
		sched := schedules.Schedules[0]
		return sched.Type, strconv.FormatInt(sched.ID, 10), sched.Enabled
	}

	t.Fatal("no schedules available for acceptance testing")
	return "", "", false
}

func testAccScheduleResourceConfig(scheduleType string, scheduleID string, enabled bool) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_schedule" "test" {
  type        = %q
  schedule_id = %q
  enabled     = %t
}
`, scheduleType, scheduleID, enabled)
}

func testAccGitStackWebhookActionConfig(env string, stackName string, repoURL string, branch string, composePath string, credentialID string, trigger string) string {
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
  webhook_enabled = true%s
}

resource "dockhand_git_stack_webhook_action" "test" {
  stack_id = dockhand_git_stack.test.id
  trigger  = %q
}
`, env, stackName, repoURL, branch, composePath, credentialBlock, trigger)
}
