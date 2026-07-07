#!/usr/bin/env python3
"""Unblock open agent/automation PRs stuck on action_required workflow runs."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

HEAD_PREFIXES = ("agent/", "automation/")


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


def is_target_head(head_ref: str) -> bool:
    ref = (head_ref or "").strip()
    return any(ref.startswith(prefix) for prefix in HEAD_PREFIXES)


def open_agent_pull_requests() -> list[dict]:
    data = _gh_json(
        [
            "pr",
            "list",
            "--state",
            "open",
            "--json",
            "number,headRefName,headRefOid,labels",
            "--limit",
            "100",
        ]
    )
    if not isinstance(data, list):
        return []
    return [item for item in data if is_target_head(str(item.get("headRefName", "")))]


def dispatch_pr_approve(pull_number: int, head_sha: str) -> None:
    proc = subprocess.run(
        [
            "gh",
            "workflow",
            "run",
            "agent-pr-approve-ci.yml",
            "--repo",
            _repo(),
            "-f",
            f"head_sha={head_sha}",
            "-f",
            f"pull_number={pull_number}",
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "workflow dispatch failed")


def unblock_pull_requests(*, dispatch_merge: bool = True) -> list[dict]:
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from agent_approve_pr_ci import approve_pull_request

    results: list[dict] = []
    for pr in open_agent_pull_requests():
        number = int(pr["number"])
        head_sha = str(pr["headRefOid"])
        head_ref = str(pr.get("headRefName", ""))
        approval = approve_pull_request(head_sha, retries=4, delay_seconds=10)
        entry = {
            "number": number,
            "head_ref": head_ref,
            "head_sha": head_sha,
            **approval,
        }
        if dispatch_merge:
            try:
                dispatch_pr_approve(number, head_sha)
                entry["merge_dispatch"] = "ok"
            except RuntimeError as err:
                entry["merge_dispatch"] = str(err)
        results.append(entry)
    return results


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--no-dispatch", action="store_true", help="Approve only; skip merge workflow dispatch")
    args = parser.parse_args(argv)

    try:
        results = unblock_pull_requests(dispatch_merge=not args.no_dispatch)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    blocked = [item for item in results if item.get("remaining", 0) > 0]
    failed_dispatch = [
        item for item in results if item.get("merge_dispatch") not in {None, "ok"}
    ]
    payload = {
        "pull_requests": results,
        "count": len(results),
        "blocked_count": len(blocked),
        "dispatch_failures": len(failed_dispatch),
    }
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
