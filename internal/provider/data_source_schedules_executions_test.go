package provider

import "testing"

func TestMatchesScheduleExecutionFilters(t *testing.T) {
	status := "success"
	envID := int64(2)
	item := scheduleExecutionItemResponse{
		ScheduleType:  "system_cleanup",
		ScheduleID:    2,
		EnvironmentID: &envID,
		Status:        &status,
	}

	if !matchesScheduleExecutionFilters(item, "system_cleanup", "2", "success", "2") {
		t.Fatalf("expected item to match all filters")
	}
	if matchesScheduleExecutionFilters(item, "container_update", "", "", "") {
		t.Fatalf("expected type mismatch")
	}
	if matchesScheduleExecutionFilters(item, "", "3", "", "") {
		t.Fatalf("expected id mismatch")
	}
	if matchesScheduleExecutionFilters(item, "", "", "failed", "") {
		t.Fatalf("expected status mismatch")
	}
	if matchesScheduleExecutionFilters(item, "", "", "", "99") {
		t.Fatalf("expected environment mismatch")
	}
}
