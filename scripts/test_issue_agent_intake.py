#!/usr/bin/env python3
"""Unit tests for issue-agent-intake helpers."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from issue_agent_intake import branch_name, build_prompt, slugify


class IssueAgentIntakeTest(unittest.TestCase):
    def test_slugify(self) -> None:
        self.assertEqual(slugify("Fix batch action status"), "fix-batch-action-status")
        self.assertEqual(slugify("!!!"), "task")

    def test_branch_name(self) -> None:
        self.assertEqual(branch_name(42, "Fix batch action"), "agent/issue-42-fix-batch-action")

    def test_build_prompt_includes_branch_and_issue(self) -> None:
        prompt = build_prompt(7, "Example title", "Problem details", "agent/issue-7-example-title")
        self.assertIn("agent/issue-7-example-title", prompt)
        self.assertIn("Issue #7", prompt)
        self.assertIn("Problem details", prompt)
        self.assertIn("Co-authored-by: Cursor Agent", prompt)
        self.assertIn("Required automated lens sweep", prompt)


if __name__ == "__main__":
    unittest.main()
