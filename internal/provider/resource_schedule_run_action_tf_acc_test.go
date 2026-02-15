package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleRunActionTerraform(t *testing.T) {
	endpoint, username, password := testAccEnv(t)
	defaultEnv := testAccDefaultEnv()

	t.Setenv("DOCKHAND_ENDPOINT", endpoint)
	t.Setenv("DOCKHAND_USERNAME", username)
	t.Setenv("DOCKHAND_PASSWORD", password)
	t.Setenv("DOCKHAND_DEFAULT_ENV", defaultEnv)

	scheduleType := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_SCHEDULE_TYPE"))
	if scheduleType == "" {
		scheduleType = "system_cleanup"
	}
	scheduleID := strings.TrimSpace(os.Getenv("DOCKHAND_TEST_SCHEDULE_ID"))
	if scheduleID == "" {
		scheduleID = "2"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleRunActionConfig(scheduleType, scheduleID, "acc-run-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_schedule_run_action.test", "type", scheduleType),
					resource.TestCheckResourceAttr("dockhand_schedule_run_action.test", "schedule_id", scheduleID),
				),
			},
			{
				Config: testAccScheduleRunActionConfig(scheduleType, scheduleID, "acc-run-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dockhand_schedule_run_action.test", "trigger", "acc-run-2"),
				),
			},
		},
	})
}

func testAccScheduleRunActionConfig(scheduleType string, scheduleID string, trigger string) string {
	return fmt.Sprintf(`
provider "dockhand" {}

resource "dockhand_schedule_run_action" "test" {
  type        = %q
  schedule_id = %q
  trigger     = %q
}
`, scheduleType, scheduleID, trigger)
}
