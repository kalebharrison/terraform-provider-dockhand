#!/usr/bin/env python3
"""Unit tests for automated lens selection."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from issue_agent_intake_lenses import (
    LENS_API,
    LENS_OPS,
    LENS_ACCEPTANCE,
    LENS_SENIOR,
    build_lens_instructions,
    select_lenses,
)
from issue_agent_intake import build_prompt


class IssueAgentIntakeLensesTest(unittest.TestCase):
    def test_compatibility_issue_lenses(self) -> None:
        lenses = select_lenses(["bug", "compatibility", "agent"], "Compatibility failure: dockhand latest")
        self.assertEqual(lenses[:2], [LENS_API, LENS_OPS])
        self.assertIn(LENS_ACCEPTANCE, lenses)

    def test_feature_issue_lenses(self) -> None:
        lenses = select_lenses(["enhancement"], "[Feature]: new resource")
        self.assertIn("Terraform schema & state", lenses)

    def test_lens_instructions_include_review_log(self) -> None:
        text = build_lens_instructions([LENS_API], issue_number=135, title="Compatibility failure")
        self.assertIn("docs/reports/agent-review-log.md", text)
        self.assertIn("API compatibility", text)

    def test_build_prompt_includes_lens_block(self) -> None:
        prompt = build_prompt(
            135,
            "Compatibility failure: dockhand latest",
            "body",
            "agent/issue-135-compatibility-failure-dockhand-latest",
            ["bug", "compatibility"],
        )
        self.assertIn("Required automated lens sweep", prompt)
        self.assertIn("AGENT_REVIEW_LENSES.md", prompt)

    def test_release_candidate_uses_core_or_full_lenses(self) -> None:
        body = "Tier: patch\n\n## Problem\nRelease\n\n## Done when\n- clear"
        lenses = select_lenses(["release-candidate", "agent"], "release: prepare v0.2.0", body)
        self.assertEqual(len(lenses), 5)
        body_minor = "Tier: minor\n\n## Problem\nRelease\n\n## Done when\n- clear"
        lenses_minor = select_lenses(["release-candidate"], "release: prepare v0.3.0", body_minor)
        self.assertEqual(len(lenses_minor), 11)


if __name__ == "__main__":
    unittest.main()
