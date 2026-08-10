#!/usr/bin/env python3
"""Unit tests for release-loop watchdog decisions."""

from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import agent_release_loop_watchdog as loop
import release_gate_check as gate


class ReleaseLoopWatchdogTests(unittest.TestCase):
    def test_missing_draft_with_unreleased_commits(self) -> None:
        result = gate.GateResult(unreleased_commits_on_main=4)
        self.assertEqual(loop.decide_actions(result), ["release-drafter.yml"])

    def test_ready_for_lens(self) -> None:
        result = gate.GateResult(
            ci_gates_pass=True,
            version="0.1.90",
            tag="v0.1.90",
            unreleased_commits_on_main=2,
        )
        self.assertEqual(loop.decide_actions(result), ["agent-release-orchestrate.yml"])

    def test_ready_to_tag(self) -> None:
        result = gate.GateResult(
            ci_gates_pass=True,
            version="0.1.90",
            tag="v0.1.90",
            lens_verdict_clear=True,
            unreleased_commits_on_main=2,
        )
        self.assertEqual(loop.decide_actions(result), ["agent-release-tag.yml"])


if __name__ == "__main__":
    unittest.main()
