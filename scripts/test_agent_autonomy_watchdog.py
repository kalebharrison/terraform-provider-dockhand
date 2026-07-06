#!/usr/bin/env python3
"""Unit tests for agent autonomy watchdog helpers."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from agent_ci_intake_watchdog import is_ci_failure_title
from agent_stuck_pr_watchdog import is_target_head


class AgentAutonomyWatchdogTest(unittest.TestCase):
    def test_target_agent_branch(self) -> None:
        self.assertTrue(is_target_head("agent/issue-42-fix-lint"))

    def test_target_automation_branch(self) -> None:
        self.assertTrue(is_target_head("automation/agent-autonomy-loop"))

    def test_non_target_branch(self) -> None:
        self.assertFalse(is_target_head("feature/manual-fix"))

    def test_ci_failure_title(self) -> None:
        self.assertTrue(is_ci_failure_title("CI failure: Workflow Lint"))

    def test_security_ci_failure_title(self) -> None:
        self.assertTrue(is_ci_failure_title("Security CI failure: Secret Scan"))

    def test_non_ci_title(self) -> None:
        self.assertFalse(is_ci_failure_title("[Automation] Workflow failing: Go CI"))


if __name__ == "__main__":
    unittest.main()
