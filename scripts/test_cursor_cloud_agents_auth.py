#!/usr/bin/env python3
"""Unit tests for Cloud Agents API key verification."""

from __future__ import annotations

import io
import pathlib
import sys
import unittest
import urllib.error
from unittest import mock

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import cursor_cloud_agents_auth as auth


class CursorCloudAgentsAuthTest(unittest.TestCase):
    @mock.patch.object(auth.urllib.request, "urlopen")
    def test_verify_api_key_success(self, urlopen: mock.Mock) -> None:
        response = mock.MagicMock()
        response.read.return_value = (
            b'{"apiKeyName":"ci","userId":1,"userEmail":"a@example.com",'
            b'"createdAt":"2026-01-01T00:00:00.000Z"}'
        )
        response.__enter__.return_value = response
        response.__exit__.return_value = False
        urlopen.return_value = response

        result = auth.verify_api_key("test-key")

        self.assertEqual(result["apiKeyName"], "ci")
        request = urlopen.call_args.args[0]
        self.assertEqual(request.full_url, auth.ME_URL)
        self.assertTrue(request.get_header("Authorization").startswith("Basic "))

    @mock.patch.object(auth.urllib.request, "urlopen")
    def test_verify_api_key_http_error(self, urlopen: mock.Mock) -> None:
        err = urllib.error.HTTPError(
            auth.ME_URL, 401, "Unauthorized", hdrs=None, fp=io.BytesIO(b"bad key")
        )
        urlopen.side_effect = err

        with self.assertRaisesRegex(RuntimeError, "Cloud Agents API HTTP 401"):
            auth.verify_api_key("bad")

    @mock.patch.dict(auth.os.environ, {}, clear=True)
    def test_main_missing_key(self) -> None:
        self.assertEqual(auth.main(["--json"]), 2)

    @mock.patch.object(auth, "verify_api_key")
    @mock.patch.dict(auth.os.environ, {"CURSOR_API_KEY": "k"}, clear=True)
    def test_main_redacts_email(self, verify: mock.Mock) -> None:
        verify.return_value = {
            "apiKeyName": "ci",
            "userEmail": "secret@example.com",
            "userId": 9,
            "createdAt": "2026-01-01T00:00:00.000Z",
        }
        with mock.patch("sys.stdout", new_callable=io.StringIO) as out:
            code = auth.main(["--json"])
        self.assertEqual(code, 0)
        printed = out.getvalue()
        self.assertNotIn("secret@example.com", printed)
        self.assertIn('"apiKeyName": "ci"', printed)


if __name__ == "__main__":
    unittest.main()
