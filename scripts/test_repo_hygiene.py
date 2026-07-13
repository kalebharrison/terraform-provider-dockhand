#!/usr/bin/env python3
"""Unit tests for repo hygiene branch pruning helpers."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from repo_hygiene import KEEP_BRANCH_NAMES, is_prunable_branch


class RepoHygieneBranchTest(unittest.TestCase):
    def test_agent_branch_is_prunable(self) -> None:
        self.assertTrue(is_prunable_branch("agent/issue-42-fix"))

    def test_automation_branch_is_prunable(self) -> None:
        self.assertTrue(is_prunable_branch("automation/pr-merge-retry"))

    def test_cursor_branch_is_prunable(self) -> None:
        self.assertTrue(is_prunable_branch("cursor/repository-hygiene-automation-919d"))

    def test_main_is_not_prunable(self) -> None:
        self.assertFalse(is_prunable_branch("main"))

    def test_compat_sync_branch_is_kept(self) -> None:
        self.assertIn("automation/compat-reports-sync", KEEP_BRANCH_NAMES)


if __name__ == "__main__":
    unittest.main()
