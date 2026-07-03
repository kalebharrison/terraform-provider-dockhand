#!/usr/bin/env python3
"""Fail when manifest-mapped TestAcc suites skip during acceptance runs."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


def load_manifest_regexes(manifest_path: Path) -> list[re.Pattern[str]]:
    data = json.loads(manifest_path.read_text(encoding="utf-8"))
    patterns: list[re.Pattern[str]] = []
    for section in ("resources", "data_sources"):
        for entry in data.get(section, []):
            raw = str(entry.get("acceptance_test_regex", "")).strip()
            if raw:
                patterns.append(re.compile(raw))
    return patterns


def is_manifest_mapped(test_name: str, patterns: list[re.Pattern[str]]) -> bool:
    return any(pattern.search(test_name) for pattern in patterns)


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: check-acceptance-skips.py <go-test-jsonl>", file=sys.stderr)
        return 2

    jsonl_path = Path(sys.argv[1])
    if not jsonl_path.is_file():
        print(f"acceptance test output not found: {jsonl_path}", file=sys.stderr)
        return 1

    manifest = Path("internal/provider/testdata/acceptance_manifest.json")
    if not manifest.is_file():
        manifest = Path(__file__).resolve().parents[1] / "internal/provider/testdata/acceptance_manifest.json"
    patterns = load_manifest_regexes(manifest)

    skipped: list[str] = []
    for line in jsonl_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("Action") != "skip":
            continue
        test_name = str(event.get("Test", "")).strip()
        if not test_name:
            continue
        if is_manifest_mapped(test_name, patterns):
            skipped.append(test_name)

    if not skipped:
        print("acceptance skip check passed (no manifest-mapped skips)")
        return 0

    unique = sorted(set(skipped))
    print("manifest-mapped acceptance tests skipped:", file=sys.stderr)
    for name in unique:
        print(f"  - {name}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
