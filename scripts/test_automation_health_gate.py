#!/usr/bin/env python3
"""Unit tests for automation health gate helpers."""

from __future__ import annotations

import pathlib
import sys
import tempfile
import unittest
from datetime import datetime, timedelta, timezone
from pathlib import Path

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from automation_health_gate import (
    ALERT_HOURS,
    blockers_fingerprint,
    effective_blockers,
    evaluate_health,
    hours_blocked,
    should_track,
)


class AutomationHealthGateTests(unittest.TestCase):
    def test_effective_blockers_adds_orchestrator_gap(self) -> None:
        gate = {
            "blockers": [],
            "ready_for_lens_dispatch": True,
            "open_release_issue": None,
        }
        blockers = effective_blockers(gate)
        self.assertIn("release orchestrator has not opened release lens issue", blockers)
        self.assertTrue(should_track(gate))

    def test_fingerprint_changes_when_blockers_change(self) -> None:
        first = blockers_fingerprint(["workflow not green on main: Go CI"])
        second = blockers_fingerprint(["open compatibility issue #135"])
        self.assertNotEqual(first, second)

    def test_alert_after_threshold(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            state_path = Path(tmp) / "state.env"
            now = datetime(2026, 7, 4, 12, 0, tzinfo=timezone.utc)
            past = now - timedelta(hours=ALERT_HOURS + 1)
            state_path.write_text(
                "FIRST_SEEN=2026-07-03T11:00:00Z\nBLOCKERS_FP=deadbeef\n",
                encoding="utf-8",
            )

            gate = {
                "ci_gates_pass": False,
                "ready_for_lens_dispatch": False,
                "ready_to_tag": False,
                "version": "0.1.85",
                "tag": "v0.1.85",
                "tier": "patch",
                "blockers": ["open compatibility issue #135"],
                "awaiting_release_issues": [132],
                "lens_verdict_clear": False,
                "open_release_issue": None,
            }

            from automation_health_gate import (
                blockers_fingerprint,
                format_issue_body,
                parse_state,
                write_state,
            )

            fp = blockers_fingerprint(effective_blockers(gate))
            write_state(state_path, past, fp)
            first_seen, _ = parse_state(state_path)
            self.assertIsNotNone(first_seen)
            hours = hours_blocked(first_seen, now)
            self.assertGreaterEqual(hours, ALERT_HOURS)
            body = format_issue_body(gate, gate["blockers"], first_seen, hours)
            self.assertIn("open compatibility issue #135", body)

    def test_clear_when_gate_healthy(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            state_path = Path(tmp) / "state.env"
            state_path.write_text("FIRST_SEEN=2026-07-03T11:00:00Z\nBLOCKERS_FP=abc\n", encoding="utf-8")

            original = evaluate_health.__globals__["evaluate_gate"]

            class Healthy:
                def as_json(self) -> dict:
                    return {
                        "ci_gates_pass": True,
                        "ready_for_lens_dispatch": False,
                        "ready_to_tag": True,
                        "version": "0.1.85",
                        "tag": "v0.1.85",
                        "tier": "patch",
                        "blockers": [],
                        "awaiting_release_issues": [],
                        "lens_verdict_clear": True,
                        "open_release_issue": None,
                    }

            evaluate_health.__globals__["evaluate_gate"] = lambda: Healthy()
            try:
                result = evaluate_health(state_path=state_path)
            finally:
                evaluate_health.__globals__["evaluate_gate"] = original

            self.assertTrue(result["should_close_issue"])
            self.assertFalse(state_path.exists())


if __name__ == "__main__":
    raise SystemExit(unittest.main())
