#!/usr/bin/env python3
"""Unit tests for merge-cleanup watchdog helpers."""

from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import agent_merge_cleanup_watchdog as watchdog


class MergeCleanupWatchdogTests(unittest.TestCase):
    def test_linked_issue_numbers(self) -> None:
        body = "Fixes #252\n\nAlso resolves #10."
        self.assertEqual(watchdog.linked_issue_numbers(body), [10, 252])

    def test_agent_branch_is_target(self) -> None:
        self.assertTrue(
            watchdog.is_cleanup_target_pr({"headRefName": "agent/issue-252-api-drift", "author": {}})
        )

    def test_automation_branch_skipped(self) -> None:
        self.assertFalse(
            watchdog.is_cleanup_target_pr({"headRefName": "automation/compat-reports-sync", "author": {}})
        )

    def test_open_issue_needs_cleanup(self) -> None:
        issue = {"state": "open", "labels": [{"name": "compatibility"}]}
        self.assertTrue(watchdog.issue_needs_cleanup(issue, []))

    def test_closed_awaiting_release_with_comment_is_done(self) -> None:
        issue = {"state": "closed", "labels": [{"name": "awaiting-release"}]}
        comments = [{"body": "hello\n\n---\n_Posted by Agent Merge Cleanup._\n"}]
        self.assertFalse(watchdog.issue_needs_cleanup(issue, comments))


if __name__ == "__main__":
    unittest.main()
