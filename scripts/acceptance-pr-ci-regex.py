#!/usr/bin/env python3
"""Validate PR CI acceptance suites and print a go test -run regex."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "internal/provider/testdata/acceptance_manifest.json"
PR_CI = ROOT / "internal/provider/testdata/acceptance_pr_ci.json"
TEST_DIR = ROOT / "internal/provider"


def load_json(path: Path) -> dict:
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def discover_testacc_functions() -> set[str]:
    pattern = re.compile(r"func\s+(TestAcc[0-9A-Za-z_]+)\s*\(")
    names: set[str] = set()
    for path in TEST_DIR.glob("*_test.go"):
        text = path.read_text(encoding="utf-8")
        names.update(pattern.findall(text))
    if not names:
        raise SystemExit("no TestAcc functions discovered")
    return names


def manifest_regexes(manifest: dict) -> list[str]:
    regexes: list[str] = []
    for section in ("resources", "data_sources"):
        for entry in manifest.get(section, []):
            regex = str(entry.get("acceptance_test_regex", "")).strip()
            if regex:
                regexes.append(regex)
    return regexes


def main() -> int:
    manifest = load_json(MANIFEST)
    pr_ci = load_json(PR_CI)
    suites = [str(item).strip() for item in pr_ci.get("suites", []) if str(item).strip()]

    if not suites:
        print("acceptance_pr_ci.json must list at least one suite", file=sys.stderr)
        return 1

    test_names = discover_testacc_functions()
    compiled_manifest = [re.compile(regex) for regex in manifest_regexes(manifest)]

    for suite in suites:
        if suite not in test_names:
            print(f"PR CI suite {suite!r} is not a TestAcc function", file=sys.stderr)
            return 1
        if not any(pattern.search(suite) for pattern in compiled_manifest):
            print(
                f"PR CI suite {suite!r} is not covered by any acceptance_manifest.json regex",
                file=sys.stderr,
            )
            return 1

    suffixes: list[str] = []
    for suite in suites:
        if suite.startswith("TestAcc"):
            suffixes.append(suite[len("TestAcc") :])
        else:
            suffixes.append(suite)

    regex = "TestAcc(" + "|".join(suffixes) + ")"
    print(regex)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
