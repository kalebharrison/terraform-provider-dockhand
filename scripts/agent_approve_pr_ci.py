#!/usr/bin/env python3
"""Approve or re-run PR workflow runs blocked on maintainer approval."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


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


def _gh(args: list[str], *, use_repo_flag: bool = True) -> subprocess.CompletedProcess[str]:
    import os

    cmd = ["gh", *args]
    env = os.environ.copy()
    if use_repo_flag and args and args[0] != "api":
        cmd.extend(["--repo", _repo()])
    else:
        env.setdefault("GH_REPO", _repo())
    return subprocess.run(
        cmd,
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
        env=env,
    )


def runs_for_sha(head_sha: str) -> list[dict]:
    proc = _gh(["api", f"repos/{_repo()}/actions/runs?head_sha={head_sha}&per_page=100"], use_repo_flag=False)
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh failed")
    data = json.loads(proc.stdout)
    if not isinstance(data, dict):
        return []
    runs = data.get("workflow_runs")
    return runs if isinstance(runs, list) else []


def needs_approval(run: dict) -> bool:
    status = str(run.get("status", "")).lower()
    conclusion = str(run.get("conclusion") or "").lower()
    if status == "waiting":
        return True
    return status == "completed" and conclusion == "action_required"


def approve_run(run_id: int) -> bool:
    proc = _gh(["api", "--method", "POST", f"repos/{_repo()}/actions/runs/{run_id}/approve"], use_repo_flag=False)
    return proc.returncode == 0


def rerun_run(run_id: int) -> bool:
    proc = _gh(["run", "rerun", str(run_id)])
    return proc.returncode == 0


def approve_pull_request(head_sha: str, *, retries: int = 12, delay_seconds: int = 15) -> dict:
    approved = 0
    rerun = 0
    skipped = 0
    errors: list[str] = []

    for attempt in range(retries):
        pending = [run for run in runs_for_sha(head_sha) if needs_approval(run)]
        if not pending and attempt > 0:
            break
        for run in pending:
            run_id = int(run["id"])
            name = str(run.get("name", run_id))
            status = str(run.get("status", "")).lower()
            if status == "waiting" and approve_run(run_id):
                approved += 1
                continue
            if approve_run(run_id):
                approved += 1
                continue
            if rerun_run(run_id):
                rerun += 1
                continue
            errors.append(f"{name} ({run_id})")
        if attempt + 1 < retries:
            time.sleep(delay_seconds)

    remaining = [run for run in runs_for_sha(head_sha) if needs_approval(run)]
    return {
        "head_sha": head_sha,
        "approved": approved,
        "rerun": rerun,
        "skipped": skipped,
        "remaining": len(remaining),
        "errors": errors,
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--head-sha", required=True)
    parser.add_argument("--retries", type=int, default=12)
    parser.add_argument("--delay-seconds", type=int, default=15)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    try:
        result = approve_pull_request(
            args.head_sha,
            retries=args.retries,
            delay_seconds=args.delay_seconds,
        )
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))

    return 0 if result["remaining"] == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
