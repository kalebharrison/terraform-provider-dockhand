#!/usr/bin/env python3
"""Unit tests for agent PR CI approval helpers."""

from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from agent_approve_pr_ci import needs_approval


class AgentApprovePrCiTests(unittest.TestCase):
    def test_waiting_run_needs_approval(self) -> None:
        self.assertTrue(needs_approval({"status": "waiting", "conclusion": None}))

    def test_action_required_run_needs_approval(self) -> None:
        self.assertTrue(
            needs_approval({"status": "completed", "conclusion": "action_required"})
        )

    def test_success_run_does_not_need_approval(self) -> None:
        self.assertFalse(needs_approval({"status": "completed", "conclusion": "success"}))


if __name__ == "__main__":
    unittest.main()
