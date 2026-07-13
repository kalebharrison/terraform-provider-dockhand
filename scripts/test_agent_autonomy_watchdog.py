#!/usr/bin/env python3
"""Unit tests for agent autonomy watchdog helpers."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from agent_stuck_pr_watchdog import is_target_head, is_target_pull_request


class AgentAutonomyWatchdogTest(unittest.TestCase):
    def test_target_agent_branch(self) -> None:
        self.assertTrue(is_target_head("agent/issue-42-fix-lint"))

    def test_target_automation_branch(self) -> None:
        self.assertTrue(is_target_head("automation/agent-autonomy-loop"))

    def test_non_target_branch(self) -> None:
        self.assertFalse(is_target_head("feature/manual-fix"))

    def test_cursor_bot_pr_is_targeted(self) -> None:
        self.assertTrue(
            is_target_pull_request(
                {
                    "headRefName": "cursor/repository-hygiene-automation-919d",
                    "author": {"login": "app/cursor"},
                }
            )
        )

    def test_untrusted_cursor_named_pr_is_not_targeted(self) -> None:
        self.assertFalse(
            is_target_pull_request(
                {
                    "headRefName": "cursor/not-from-cursor",
                    "author": {"login": "random-user"},
                }
            )
        )


if __name__ == "__main__":
    unittest.main()
