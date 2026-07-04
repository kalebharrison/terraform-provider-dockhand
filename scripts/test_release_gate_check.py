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
            {"status": "completed", "conclusion": "failure", "headSha": "abc", "databaseId": 1},
            {"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 2},
        ]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", return_value=True):
                self.assertTrue(gate.release_watch_main_sha_green("abc"))

    def test_workflow_green_on_main_sha_skip_only_run(self) -> None:
        runs = [{"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", return_value=False):
                self.assertFalse(gate.release_watch_main_sha_green("abc"))

    def test_workflow_green_on_main_sha_no_match(self) -> None:
        runs = [{"status": "completed", "conclusion": "success", "headSha": "other", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            self.assertFalse(gate.release_watch_main_sha_green("abc"))

    def test_release_watch_strict_green_requires_validate(self) -> None:
        runs = [{"status": "completed", "conclusion": "success", "headSha": "abc", "databaseId": 1}]
        with mock.patch.object(gate, "workflow_runs", return_value=runs):
            with mock.patch.object(gate, "release_watch_run_validated", return_value=False):
                self.assertFalse(gate.release_watch_strict_green())
            with mock.patch.object(gate, "release_watch_run_validated", return_value=True):
                self.assertTrue(gate.release_watch_strict_green())

    def test_workflow_is_green_main_sha_mode(self) -> None:
        with mock.patch.object(gate, "release_watch_main_sha_green", return_value=True):
            self.assertTrue(
                gate.workflow_is_green(
                    gate.RELEASE_WATCH_WORKFLOW,
                    release_watch_mode="main_sha",
                )
            )

    def test_workflow_is_green_strict_uses_validated_latest(self) -> None:
        with mock.patch.object(gate, "release_watch_strict_green", return_value=False):
            self.assertFalse(
                gate.workflow_is_green(
                    gate.RELEASE_WATCH_WORKFLOW,
                    release_watch_mode="strict",
                )
            )


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


if __name__ == "__main__":
    unittest.main()
