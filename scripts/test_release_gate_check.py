#!/usr/bin/env python3
"""Unit tests for release gate workflow evaluation helpers."""

import pathlib
import sys
import unittest
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import release_gate_check as gate


class WorkflowGreenTest(unittest.TestCase):
    def test_workflow_green_on_main_sha_success(self) -> None:
        runs = [
            {"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 1},
            {"status": "completed", "conclusion": "success", "headSha": "other", "databaseId": 2},
        ]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", side_effect=[False, True]):
                self.assertTrue(gate.release_watch_chain_green())

    def test_workflow_green_on_main_sha_skip_only_chain(self) -> None:
        runs = [{"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", return_value=False):
                self.assertFalse(gate.release_watch_chain_green())

    def test_workflow_green_on_main_sha_failure_breaks_chain(self) -> None:
        runs = [{"status": "completed", "conclusion": "failure", "headSha": "abc", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            self.assertFalse(gate.release_watch_chain_green())

    def test_release_watch_strict_green_requires_validated_in_chain(self) -> None:
        runs = [{"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", return_value=True):
                self.assertTrue(gate.release_watch_strict_green())

    def test_workflow_is_green_main_sha_mode(self) -> None:
        with mock.patch.object(gate, "release_watch_chain_green", return_value=True):
            self.assertTrue(
                gate.workflow_is_green(
                    gate.RELEASE_WATCH_WORKFLOW,
                    release_watch_mode="main_sha",
                )
            )

    def test_workflow_is_green_strict_uses_validated_latest(self) -> None:
        with mock.patch.object(gate, "release_watch_chain_green", return_value=False):
            self.assertFalse(
                gate.workflow_is_green(
                    gate.RELEASE_WATCH_WORKFLOW,
                    release_watch_mode="strict",
                )
            )


class GateResultTest(unittest.TestCase):
    def test_ready_for_lens_with_unreleased_commits(self) -> None:
        result = gate.GateResult(
            ci_gates_pass=True,
            version="0.1.87",
            tag="v0.1.87",
            unreleased_commits_on_main=3,
        )
        self.assertTrue(result.ready_for_lens_dispatch)
        self.assertEqual(result.release_trigger, "unreleased_main_commits")

    def test_ready_for_lens_without_release_work(self) -> None:
        result = gate.GateResult(ci_gates_pass=True, version="0.1.87", tag="v0.1.87")
        self.assertFalse(result.ready_for_lens_dispatch)

    def test_awaiting_release_takes_precedence(self) -> None:
        result = gate.GateResult(
            ci_gates_pass=True,
            version="0.1.87",
            tag="v0.1.87",
            awaiting_release_issues=[42],
            unreleased_commits_on_main=5,
        )
        self.assertEqual(result.release_trigger, "awaiting_release_issues")


class CompatibilityIssueFilterTest(unittest.TestCase):
    def test_linked_issue_numbers_parses_fixes_clause(self) -> None:
        text = "Fixes #252\n\nAlso closes #99"
        self.assertEqual(gate._linked_issue_numbers(text), {252, 99})

    def test_issue_has_merged_closing_pr_true(self) -> None:
        with mock.patch.object(
            gate,
            "_gh_json",
            return_value=[{"number": 254, "body": "Fixes #252", "title": "fix(probe): navigation"}],
        ):
            self.assertTrue(gate.issue_has_merged_closing_pr(252))
            self.assertEqual(gate.merged_closing_pull_request(252), 254)

    def test_issue_has_merged_closing_pr_false_when_unrelated(self) -> None:
        with mock.patch.object(
            gate,
            "_gh_json",
            return_value=[{"number": 300, "body": "Fixes #999", "title": "unrelated"}],
        ):
            self.assertFalse(gate.issue_has_merged_closing_pr(252))

    def test_open_compatibility_issues_skips_merged_fix(self) -> None:
        with mock.patch.object(
            gate,
            "_gh_json",
            return_value=[{"number": 252, "title": "API drift detected"}],
        ):
            with mock.patch.object(gate, "issue_has_merged_closing_pr", return_value=True):
                self.assertEqual(gate.open_compatibility_issues(), [])

    def test_open_compatibility_issues_keeps_unresolved(self) -> None:
        with mock.patch.object(
            gate,
            "_gh_json",
            return_value=[{"number": 252, "title": "API drift detected"}],
        ):
            with mock.patch.object(gate, "issue_has_merged_closing_pr", return_value=False):
                self.assertEqual(gate.open_compatibility_issues(), [252])


class CutChangelogTest(unittest.TestCase):
    def test_cut_changelog_inserts_version_section(self) -> None:
        from release_housekeeping import cut_changelog

        changelog = gate.ROOT / "CHANGELOG.md"
        original = changelog.read_text(encoding="utf-8")
        try:
            text = "## [Unreleased]\n\n### Added\n\n- item\n"
            changelog.write_text(text, encoding="utf-8")
            self.assertTrue(cut_changelog("0.9.9"))
            updated = changelog.read_text(encoding="utf-8")
            self.assertIn("## [0.9.9]", updated)
            self.assertIn("## [Unreleased]", updated)
        finally:
            changelog.write_text(original, encoding="utf-8")


class ReleaseIssueNotifyTest(unittest.TestCase):
    def test_build_comment_body_includes_marker(self) -> None:
        from release_issue_notify import NOTIFY_MARKER, build_comment_body

        body = build_comment_body(tag="v1.2.3", release_url="https://example/release")
        self.assertIn("v1.2.3", body)
        self.assertIn(NOTIFY_MARKER, body)

    def test_section_from_body(self) -> None:
        from release_issue_notify import section_from_body

        text = "## What was fixed\n\nhello\n\n## User impact\n\nworld"
        self.assertEqual(section_from_body("What was fixed", text), "hello")
        self.assertEqual(section_from_body("User impact", text), "world")


if __name__ == "__main__":
    unittest.main()
