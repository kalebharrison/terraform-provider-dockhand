#!/usr/bin/env python3
"""Detect substantive Compat Reports Sync changes (ignore volatile metadata)."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
TRACKED_RELPATHS = (
    "docs/reports/endpoint-probe.md",
    "docs/reports/endpoint-probe.csv",
    "docs/reports/webui-api-endpoints.txt",
    "docs/reports/docs-reference-api-endpoints.txt",
    "docs/reports/dockhand-last-tested.json",
    "docs/non-present-endpoints.md",
)
VOLATILE_JSON_KEYS = frozenset({"updated_at", "source_run"})
DATE_LINE_RE = re.compile(r"(?m)^- Date: .*$")


def normalize_content(relpath: str, text: str) -> str:
    """Strip fields that change every green CI run without real baseline drift."""
    name = Path(relpath).name
    if name == "dockhand-last-tested.json":
        data = json.loads(text) if text.strip() else {}
        if not isinstance(data, dict):
            return text
        for key in VOLATILE_JSON_KEYS:
            data.pop(key, None)
        return json.dumps(data, sort_keys=True, indent=2) + "\n"
    if name == "non-present-endpoints.md":
        return DATE_LINE_RE.sub("- Date: <normalized>", text)
    return text


def _git_show(repo: Path, rev: str, relpath: str) -> str | None:
    proc = subprocess.run(
        ["git", "-C", str(repo), "show", f"{rev}:{relpath}"],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        return None
    return proc.stdout


def substantive_paths_changed(
    *,
    repo: Path = ROOT,
    base_rev: str = "HEAD",
    relpaths: tuple[str, ...] = TRACKED_RELPATHS,
) -> list[str]:
    """Return tracked paths whose normalized content differs from base_rev."""
    changed: list[str] = []
    for relpath in relpaths:
        path = repo / relpath
        if path.is_file():
            working = path.read_text(encoding="utf-8")
        else:
            working = None
        base = _git_show(repo, base_rev, relpath)
        if working is None and base is None:
            continue
        if working is None or base is None:
            changed.append(relpath)
            continue
        if normalize_content(relpath, working) != normalize_content(relpath, base):
            changed.append(relpath)
    return changed


def restore_tracked_files(*, repo: Path = ROOT, relpaths: tuple[str, ...] = TRACKED_RELPATHS) -> None:
    """Discard working-tree edits for tracked sync paths."""
    subprocess.run(
        ["git", "-C", str(repo), "checkout", "--", *relpaths],
        check=False,
        capture_output=True,
        text=True,
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--check-working-tree",
        action="store_true",
        help="Exit 0 if substantive changes vs HEAD; 1 if metadata-only/no-op. "
        "Restores tracked files when unchanged.",
    )
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--repo", type=Path, default=ROOT)
    parser.add_argument("--base-rev", default="HEAD")
    args = parser.parse_args(argv)

    if not args.check_working_tree:
        parser.error("pass --check-working-tree")

    changed = substantive_paths_changed(repo=args.repo, base_rev=args.base_rev)
    payload = {"changed": bool(changed), "paths": changed}
    if args.json:
        print(json.dumps(payload))
    else:
        if changed:
            print(f"substantive compat report changes: {', '.join(changed)}")
        else:
            print("compat reports unchanged after normalizing volatile metadata")

    if not changed:
        restore_tracked_files(repo=args.repo)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
