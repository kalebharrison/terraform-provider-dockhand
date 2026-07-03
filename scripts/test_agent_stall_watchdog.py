#!/usr/bin/env python3
"""Unit tests for agent stall watchdog helpers."""

from __future__ import annotations

import pathlib
import sys
import unittest
from datetime import datetime, timezone

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from agent_stall_watchdog import STALL_HOURS, _hours_since, _parse_time


class AgentStallWatchdogTests(unittest.TestCase):
    def test_hours_since(self) -> None:
        start = datetime(2026, 7, 3, 12, 0, tzinfo=timezone.utc)
        end = datetime(2026, 7, 4, 13, 0, tzinfo=timezone.utc)
        self.assertEqual(_hours_since(start, end), 25.0)

    def test_parse_time(self) -> None:
        parsed = _parse_time("2026-07-03T12:00:00Z")
        self.assertEqual(parsed.tzinfo, timezone.utc)

    def test_stall_threshold(self) -> None:
        self.assertEqual(STALL_HOURS, 24)


if __name__ == "__main__":
    unittest.main()
