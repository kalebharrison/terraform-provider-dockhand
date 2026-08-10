#!/usr/bin/env python3
"""Recover release-loop steps skipped after GITHUB_TOKEN merges to main."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
MERGE_CLEANUP_MARKER = "Posted by Agent Merge Cleanup"
AGENT_HEAD_PREFIX = "agent/"


def _repo() -> str:
    import os

    env = os.environ.get("GITHUB_REPOSITORY", "").strip()
    if env:
        return env
    proc = subprocess.run(
        ["gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=True,
    )
    return proc.stdout.strip()


def _gh_json(args: list[str]) -> object:
    cmd = ["gh", *args]
    if not args or args[0] != "api":
        cmd.extend(["--repo", _repo()])
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True, check=False)
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh failed")
    if not proc.stdout.strip():
        return None
    return json.loads(proc.stdout)


def dispatch_workflow(workflow_file: str, *, inputs: dict[str, str] | None = None) -> None:
    cmd = [
        "gh",
        "workflow",
        "run",
        workflow_file,
        "--repo",
        _repo(),
        "--ref",
        "main",
    ]
    for key, value in (inputs or {}).items():
        cmd.extend(["-f", f"{key}={value}"])
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True, check=False)
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "workflow dispatch failed")


def issue_has_cleanup_comment(issue_number: int) -> bool:
    data = _gh_json(
        [
            "issue",
            "view",
            str(issue_number),
            "--json",
            "comments",
        ]
    )
    if not isinstance(data, dict):
        return False
    comments = data.get("comments")
    if not isinstance(comments, list):
        return False
    return any(MERGE_CLEANUP_MARKER in str(item.get("body", "")) for item in comments)


def linked_issue_numbers(pr_body: str) -> set[int]:
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from release_gate_check import _linked_issue_numbers

    return _linked_issue_numbers(pr_body)


def find_missed_merge_cleanups(*, lookback_days: int = 14) -> list[dict]:
    data = _gh_json(
        [
            "pr",
            "list",
            "--state",
            "merged",
            "--json",
            "number,mergedAt,body,headRefName",
            "--limit",
            "100",
        ]
    )
    if not isinstance(data, list):
        return []

    cutoff = datetime.now(timezone.utc) - timedelta(days=lookback_days)
    pending: list[dict] = []
    for pr in data:
        head_ref = str(pr.get("headRefName", ""))
        if not head_ref.startswith(AGENT_HEAD_PREFIX):
            continue
        merged_at = str(pr.get("mergedAt", ""))
        if merged_at:
            merged_time = datetime.fromisoformat(merged_at.replace("Z", "+00:00"))
            if merged_time < cutoff:
                continue
        issue_numbers = linked_issue_numbers(str(pr.get("body", "")))
        if not issue_numbers:
            continue
        needs_cleanup = False
        for issue_number in sorted(issue_numbers):
            if not issue_has_cleanup_comment(issue_number):
                needs_cleanup = True
                break
        if not needs_cleanup:
            continue
        pending.append(
            {
                "pull_number": int(pr["number"]),
                "head_ref": head_ref,
                "linked_issues": sorted(issue_numbers),
            }
        )
    return pending


def recover_release_loop(*, dry_run: bool = False) -> dict:
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from release_gate_check import draft_release_tag, evaluate_gate

    actions: list[dict] = []

    missed = find_missed_merge_cleanups()
    for item in missed:
        entry = {"action": "merge_cleanup", **item}
        if dry_run:
            entry["dispatch"] = "dry-run"
        else:
            try:
                dispatch_workflow(
                    "agent-merge-cleanup.yml",
                    inputs={"pull_number": str(item["pull_number"])},
                )
                entry["dispatch"] = "ok"
            except RuntimeError as err:
                entry["dispatch"] = str(err)
        actions.append(entry)

    if draft_release_tag() is None:
        entry = {"action": "release_drafter"}
        if dry_run:
            entry["dispatch"] = "dry-run"
        else:
            try:
                dispatch_workflow("release-drafter.yml")
                entry["dispatch"] = "ok"
            except RuntimeError as err:
                entry["dispatch"] = str(err)
        actions.append(entry)

    try:
        gate = evaluate_gate().as_json()
    except RuntimeError as err:
        return {
            "actions": actions,
            "gate_error": str(err),
            "dispatch_failures": len([a for a in actions if a.get("dispatch") not in {"ok", "dry-run"}]),
        }

    if gate.get("ready_for_lens_dispatch") and gate.get("open_release_issue") is None:
        entry = {
            "action": "release_orchestrate",
            "version": gate.get("version"),
        }
        if dry_run:
            entry["dispatch"] = "dry-run"
        else:
            try:
                dispatch_workflow("agent-release-orchestrate.yml")
                entry["dispatch"] = "ok"
            except RuntimeError as err:
                entry["dispatch"] = str(err)
        actions.append(entry)

    if gate.get("ready_to_tag"):
        entry = {
            "action": "release_tag",
            "tag": gate.get("tag"),
        }
        if dry_run:
            entry["dispatch"] = "dry-run"
        else:
            try:
                dispatch_workflow("agent-release-tag.yml")
                entry["dispatch"] = "ok"
            except RuntimeError as err:
                entry["dispatch"] = str(err)
        actions.append(entry)

    failures = [item for item in actions if item.get("dispatch") not in {"ok", "dry-run"}]
    return {
        "actions": actions,
        "gate": gate,
        "missed_merge_cleanups": len(missed),
        "dispatch_failures": len(failures),
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args(argv)

    try:
        payload = recover_release_loop(dry_run=args.dry_run)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
