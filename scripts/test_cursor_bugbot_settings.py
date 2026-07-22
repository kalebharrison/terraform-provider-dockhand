#!/usr/bin/env python3
"""Unit tests for Bugbot settings helper."""

from __future__ import annotations

import io
import pathlib
import sys
import unittest
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import cursor_bugbot_settings as bugbot


class CursorBugbotSettingsTest(unittest.TestCase):
    def test_unavailable_error_detection(self) -> None:
        self.assertTrue(
            bugbot.is_unavailable_error(
                RuntimeError('Bugbot API HTTP 401: {"message":"Invalid Team API Key"}')
            )
        )
        self.assertTrue(
            bugbot.is_unavailable_error(RuntimeError("Bugbot API HTTP 403: forbidden"))
        )
        self.assertFalse(
            bugbot.is_unavailable_error(RuntimeError("Bugbot API request failed: timeout"))
        )

    @mock.patch.object(bugbot, "update_bugbot")
    @mock.patch.dict(bugbot.os.environ, {"CURSOR_API_KEY": "k"}, clear=True)
    def test_best_effort_skips_team_key_errors(self, update: mock.Mock) -> None:
        update.side_effect = RuntimeError(
            'Bugbot API HTTP 401: {"message":"Invalid Team API Key"}'
        )
        with mock.patch("sys.stdout", new_callable=io.StringIO) as out:
            with mock.patch("sys.stderr", new_callable=io.StringIO) as err:
                code = bugbot.main(["--disable", "--best-effort", "--json"])
        self.assertEqual(code, 0)
        self.assertIn('"skipped": true', out.getvalue())
        self.assertIn("Bugbot API unavailable", err.getvalue())

    @mock.patch.object(bugbot, "update_bugbot")
    @mock.patch.dict(bugbot.os.environ, {"CURSOR_API_KEY": "k"}, clear=True)
    def test_best_effort_skips_other_errors(self, update: mock.Mock) -> None:
        update.side_effect = RuntimeError("Bugbot API request failed: timeout")
        with mock.patch("sys.stdout", new_callable=io.StringIO) as out:
            code = bugbot.main(["--disable", "--best-effort", "--json"])
        self.assertEqual(code, 0)
        self.assertIn('"skipped": true', out.getvalue())

    @mock.patch.object(bugbot, "update_bugbot")
    @mock.patch.dict(bugbot.os.environ, {"CURSOR_API_KEY": "k"}, clear=True)
    def test_strict_mode_fails_team_key_errors(self, update: mock.Mock) -> None:
        update.side_effect = RuntimeError(
            'Bugbot API HTTP 401: {"message":"Invalid Team API Key"}'
        )
        with mock.patch("sys.stderr", new_callable=io.StringIO):
            code = bugbot.main(["--disable", "--json"])
        self.assertEqual(code, 1)


if __name__ == "__main__":
    unittest.main()
