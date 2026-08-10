#!/usr/bin/env python3
"""Unit tests for release loop watchdog helpers."""

import pathlib
import sys
import unittest
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from agent_release_loop_watchdog import linked_issue_numbers


class ReleaseLoopWatchdogTest(unittest.TestCase):
    def test_linked_issue_numbers_from_pr_body(self) -> None:
        body = "Fixes #256\n\n## What was fixed\n\nrelease loop"
        self.assertEqual(linked_issue_numbers(body), {256})

    def test_recover_dispatches_merge_cleanup(self) -> None:
        import agent_release_loop_watchdog as watchdog

        missed = [{"pull_number": 257, "head_ref": "agent/issue-256-fix", "linked_issues": [256]}]
        fake_version = mock.Mock()
        with mock.patch.object(watchdog, "find_missed_merge_cleanups", return_value=missed):
            with mock.patch("release_gate_check.draft_release_tag", return_value=fake_version):
                with mock.patch(
                    "release_gate_check.evaluate_gate",
                    return_value=mock.Mock(
                        as_json=lambda: {"ready_for_lens_dispatch": False, "ready_to_tag": False}
                    ),
                ):
                    with mock.patch.object(watchdog, "dispatch_workflow") as dispatch:
                        payload = watchdog.recover_release_loop(dry_run=False)
                        dispatch.assert_any_call(
                            "agent-merge-cleanup.yml",
                            inputs={"pull_number": "257"},
                        )
                        self.assertEqual(payload["missed_merge_cleanups"], 1)


if __name__ == "__main__":
    unittest.main()
