#!/usr/bin/env python3
"""Verify GitHub Actions settings required for hands-off agent automation."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import _repo


def _gh_api(path: str) -> dict:
    proc = subprocess.run(
        ["gh", "api", path],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh api failed")
    return json.loads(proc.stdout)


def check_settings() -> dict:
    repo = _repo()
    blockers: list[str] = []

    permissions = _gh_api(f"repos/{repo}/actions/permissions")
    if not permissions.get("enabled", True):
        blockers.append("GitHub Actions is disabled for this repository")

    try:
        workflow_permissions = _gh_api(f"repos/{repo}/actions/permissions/workflow")
    except RuntimeError as err:
        blockers.append(f"could not read workflow permissions: {err}")
        workflow_permissions = {}

    if workflow_permissions.get("default_workflow_permissions") != "write":
        blockers.append(
            "workflow default permissions must be 'write' "
            f"(current: {workflow_permissions.get('default_workflow_permissions')!r})"
        )

    if not workflow_permissions.get("can_approve_pull_requests_reviews"):
        blockers.append(
            "Enable Settings → Actions → General → "
            "'Allow GitHub Actions to create and approve pull requests'"
        )

    return {
        "ok": not blockers,
        "blockers": blockers,
        "workflow_permissions": workflow_permissions,
        "actions_enabled": permissions.get("enabled", True),
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    try:
        payload = check_settings()
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))

    return 0 if payload["ok"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
