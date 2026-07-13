#!/usr/bin/env python3
"""Unit tests for automation settings verification."""

import pathlib
import sys
import unittest
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import check_automation_settings as settings


class CheckAutomationSettingsTest(unittest.TestCase):
    @mock.patch.object(settings, "_repo", return_value="owner/repo")
    @mock.patch.object(settings, "_gh_api")
    def test_integration_403_is_unverifiable_not_misconfigured(
        self, api: mock.Mock, _repo: mock.Mock
    ) -> None:
        api.side_effect = RuntimeError(
            "gh: Resource not accessible by integration (HTTP 403)"
        )

        result = settings.check_settings()

        self.assertTrue(result["ok"])
        self.assertFalse(result["settings_readable"])
        self.assertEqual(result["blockers"], [])

    @mock.patch.object(settings, "_repo", return_value="owner/repo")
    @mock.patch.object(settings, "_gh_api")
    def test_readable_bad_settings_are_blockers(
        self, api: mock.Mock, _repo: mock.Mock
    ) -> None:
        api.side_effect = [
            {"enabled": True},
            {
                "default_workflow_permissions": "read",
                "can_approve_pull_request_reviews": False,
            },
        ]

        result = settings.check_settings()

        self.assertFalse(result["ok"])
        self.assertTrue(result["settings_readable"])
        self.assertEqual(len(result["blockers"]), 2)


if __name__ == "__main__":
    unittest.main()
