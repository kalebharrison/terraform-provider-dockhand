#!/usr/bin/env python3
"""Unit tests for release watch cache/skip state."""

import pathlib
import sys
import tempfile
import unittest
from datetime import datetime, timedelta, timezone
from pathlib import Path

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import release_watch_state as state


class ReleaseWatchStateTest(unittest.TestCase):
    def test_decide_force_validate(self) -> None:
        should_run, reason = state.decide_should_run(
            tag="1.0.0",
            digest="sha256:abc",
            state={"last_tag": "1.0.0", "last_digest": "sha256:abc"},
            event_name="workflow_dispatch",
            force_validate=True,
            manual_image_tag=False,
            main_sha="abc123",
        )
        self.assertTrue(should_run)
        self.assertEqual(reason, "gate_required")

    def test_decide_skip_unchanged(self) -> None:
        should_run, reason = state.decide_should_run(
            tag="1.0.0",
            digest="sha256:abc",
            state={
                "last_tag": "1.0.0",
                "last_digest": "sha256:abc",
                "updated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
                "last_provider_sha": "abc123",
            },
            event_name="workflow_dispatch",
            force_validate=False,
            manual_image_tag=False,
            main_sha="abc123",
        )
        self.assertFalse(should_run)
        self.assertEqual(reason, "unchanged_tag_digest")

    def test_decide_provider_main_changed(self) -> None:
        should_run, reason = state.decide_should_run(
            tag="1.0.0",
            digest="sha256:abc",
            state={
                "last_tag": "1.0.0",
                "last_digest": "sha256:abc",
                "last_provider_sha": "oldsha",
            },
            event_name="workflow_dispatch",
            force_validate=False,
            manual_image_tag=False,
            main_sha="newsha",
        )
        self.assertTrue(should_run)
        self.assertEqual(reason, "provider_main_changed")

    def test_decide_schedule_skips_when_unchanged(self) -> None:
        stale = (datetime.now(timezone.utc) - timedelta(days=8)).strftime("%Y-%m-%dT%H:%M:%SZ")
        should_run, reason = state.decide_should_run(
            tag="1.0.0",
            digest="sha256:abc",
            state={
                "last_tag": "1.0.0",
                "last_digest": "sha256:abc",
                "updated_at": stale,
                "last_provider_sha": "abc123",
            },
            event_name="schedule",
            force_validate=False,
            manual_image_tag=False,
            main_sha="abc123",
        )
        self.assertFalse(should_run)
        self.assertEqual(reason, "unchanged_tag_digest")

    def test_decide_schedule_skips_without_updated_at(self) -> None:
        should_run, reason = state.decide_should_run(
            tag="1.0.0",
            digest="sha256:abc",
            state={
                "last_tag": "1.0.0",
                "last_digest": "sha256:abc",
                "last_provider_sha": "abc123",
            },
            event_name="schedule",
            force_validate=False,
            manual_image_tag=False,
            main_sha="abc123",
        )
        self.assertFalse(should_run)
        self.assertEqual(reason, "unchanged_tag_digest")

    def test_write_and_load_roundtrip(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "state.env"
            state.write_state(
                last_tag="0.28.0",
                last_digest="sha256:deadbeef",
                source_run="https://example/run/1",
                last_provider_sha="sha1",
                path=path,
            )
            loaded = state.load_state(path)
            self.assertEqual(loaded["last_tag"], "0.28.0")
            self.assertEqual(loaded["last_digest"], "sha256:deadbeef")
            self.assertEqual(loaded["last_provider_sha"], "sha1")


if __name__ == "__main__":
    unittest.main()
