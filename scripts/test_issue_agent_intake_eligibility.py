#!/usr/bin/env python3
"""Unit tests for issue agent intake eligibility."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from issue_agent_intake_eligibility import has_acceptance_section, intake_eligible


class IssueAgentIntakeEligibilityTest(unittest.TestCase):
    def test_human_opened_bug_is_eligible(self) -> None:
        ok, reason = intake_eligible(
            title="[Bug]: plan fails",
            labels=["bug"],
            author_type="User",
            author_login="random-user",
            trigger="opened",
            body="Provider version: v1.0\n\nTerraform config...\n" * 5,
            has_dispatched=False,
            has_regression=False,
        )
        self.assertTrue(ok)
        self.assertIsNone(reason)

    def test_released_label_skips(self) -> None:
        ok, reason = intake_eligible(
            title="[Feature]: wait",
            labels=["enhancement", "released"],
            author_type="User",
            author_login="random-user",
            trigger="opened",
            body="Long enough body " * 10,
            has_dispatched=False,
            has_regression=False,
        )
        self.assertFalse(ok)
        self.assertIn("released", reason or "")

    def test_automation_tracker_skips(self) -> None:
        ok, _ = intake_eligible(
            title="[Automation] Workflow failing: Go CI",
            labels=["ci", "maintenance", "no-agent"],
            author_type="Bot",
            author_login="github-actions[bot]",
            trigger="opened",
            body="automation body",
            has_dispatched=False,
            has_regression=False,
        )
        self.assertFalse(ok)

    def test_compatibility_bot_issue_eligible(self) -> None:
        body = "\n".join(
            [
                "## Problem",
                "Validation failed.",
                "",
                "## Done when",
                "- Release watch passes.",
            ]
        )
        ok, reason = intake_eligible(
            title="Compatibility failure: dockhand latest",
            labels=["bug", "compatibility", "agent"],
            author_type="Bot",
            author_login="github-actions[bot]",
            trigger="opened",
            body=body,
            has_dispatched=False,
            has_regression=False,
        )
        self.assertTrue(ok)
        self.assertIsNone(reason)

    def test_bot_bug_without_compatibility_skips(self) -> None:
        ok, reason = intake_eligible(
            title="Some bot issue",
            labels=["bug"],
            author_type="Bot",
            author_login="github-actions[bot]",
            trigger="opened",
            body="short",
            has_dispatched=False,
            has_regression=False,
        )
        self.assertFalse(ok)
        self.assertIn("bot-opened", reason or "")

    def test_acceptance_section_detection(self) -> None:
        self.assertTrue(has_acceptance_section("### Problem statement\nSomething broke"))
        self.assertTrue(has_acceptance_section("## Done when\n- fixed"))

    def test_ci_failure_bot_issue_ineligible(self) -> None:
        body = "\n".join(
            [
                "## Problem",
                "Go CI failed on main.",
                "",
                "## Done when",
                "- Go CI is green on main.",
            ]
        )
        ok, reason = intake_eligible(
            title="CI failure: Go CI",
            labels=["ci", "agent", "maintenance"],
            author_type="Bot",
            author_login="github-actions[bot]",
            trigger="opened",
            body=body,
            has_dispatched=False,
            has_regression=False,
        )
        self.assertFalse(ok)
        self.assertEqual(reason, "automation tracker (not agent work)")

    def test_edited_issue_eligible(self) -> None:
        ok, reason = intake_eligible(
            title="[Bug]: plan fails",
            labels=["bug"],
            author_type="User",
            author_login="random-user",
            trigger="edited",
            body="## Problem\nbroken\n\n## Done when\n- fixed",
            has_dispatched=False,
            has_regression=False,
        )
        self.assertTrue(ok)
        self.assertIsNone(reason)

    def test_automation_health_tracker_skips(self) -> None:
        ok, reason = intake_eligible(
            title="[Automation] Release gate blocked",
            labels=["maintenance"],
            author_type="Bot",
            author_login="github-actions[bot]",
            trigger="opened",
            body="Gate blocked for 24 hours.",
            has_dispatched=False,
            has_regression=False,
        )
        self.assertFalse(ok)
        self.assertIn("automation tracker", reason or "")


if __name__ == "__main__":
    unittest.main()
