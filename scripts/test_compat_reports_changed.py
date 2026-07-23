#!/usr/bin/env python3
"""Unit tests for Compat Reports Sync substantive-change detection."""

from __future__ import annotations

import json
import pathlib
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import compat_reports_changed as crc


class CompatReportsChangedTest(unittest.TestCase):
    def test_normalize_strips_volatile_json_keys(self) -> None:
        raw = json.dumps(
            {
                "digest": "sha256:abc",
                "image": "fnsys/dockhand:latest",
                "updated_at": "2026-07-23T08:47:46Z",
                "source_run": "https://example/run/1",
                "tag": "latest",
            }
        )
        normalized = crc.normalize_content("docs/reports/dockhand-last-tested.json", raw)
        data = json.loads(normalized)
        self.assertEqual(data["digest"], "sha256:abc")
        self.assertNotIn("updated_at", data)
        self.assertNotIn("source_run", data)

    def test_normalize_date_line_in_non_present(self) -> None:
        text = "# Non-present\n\n- Date: July 22, 2026 (CI Acceptance Full / Release Watch)\n- Other: keep\n"
        a = crc.normalize_content("docs/non-present-endpoints.md", text)
        b = crc.normalize_content(
            "docs/non-present-endpoints.md",
            text.replace("July 22, 2026", "July 23, 2026"),
        )
        self.assertEqual(a, b)
        self.assertIn("- Other: keep", a)

    def test_normalize_leaves_probe_body(self) -> None:
        text = "# probe\nroute A\n"
        self.assertEqual(crc.normalize_content("docs/reports/endpoint-probe.md", text), text)

    def test_substantive_paths_changed_ignores_metadata_only(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            reports = repo / "docs" / "reports"
            reports.mkdir(parents=True)
            (repo / "docs").mkdir(exist_ok=True)

            base_json = {
                "digest": "sha256:same",
                "image": "fnsys/dockhand:latest",
                "tag": "latest",
                "updated_at": "2026-07-22T00:00:00Z",
                "source_run": "https://example/old",
            }
            new_json = dict(base_json)
            new_json["updated_at"] = "2026-07-23T00:00:00Z"
            new_json["source_run"] = "https://example/new"

            json_path = reports / "dockhand-last-tested.json"
            json_path.write_text(json.dumps(base_json, indent=2) + "\n", encoding="utf-8")
            non_present = repo / "docs" / "non-present-endpoints.md"
            non_present.write_text(
                "- Date: July 22, 2026 (CI Acceptance Full / Release Watch)\n",
                encoding="utf-8",
            )
            probe = reports / "endpoint-probe.md"
            probe.write_text("stable probe\n", encoding="utf-8")

            # Fake HEAD via a "base" file store by monkeypatching _git_show.
            base_files = {
                "docs/reports/dockhand-last-tested.json": json_path.read_text(encoding="utf-8"),
                "docs/non-present-endpoints.md": non_present.read_text(encoding="utf-8"),
                "docs/reports/endpoint-probe.md": probe.read_text(encoding="utf-8"),
            }

            def fake_show(_repo: Path, _rev: str, relpath: str) -> str | None:
                return base_files.get(relpath)

            # Metadata-only edits in working tree.
            json_path.write_text(json.dumps(new_json, indent=2) + "\n", encoding="utf-8")
            non_present.write_text(
                "- Date: July 23, 2026 (CI Acceptance Full / Release Watch)\n",
                encoding="utf-8",
            )

            original = crc._git_show
            crc._git_show = fake_show  # type: ignore[assignment]
            try:
                self.assertEqual(
                    crc.substantive_paths_changed(repo=repo, base_rev="HEAD"),
                    [],
                )
                # Digests diverge → substantive.
                new_json["digest"] = "sha256:other"
                json_path.write_text(json.dumps(new_json, indent=2) + "\n", encoding="utf-8")
                self.assertEqual(
                    crc.substantive_paths_changed(repo=repo, base_rev="HEAD"),
                    ["docs/reports/dockhand-last-tested.json"],
                )
            finally:
                crc._git_show = original  # type: ignore[assignment]


if __name__ == "__main__":
    unittest.main()
