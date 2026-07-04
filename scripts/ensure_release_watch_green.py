#!/usr/bin/env python3
"""Ensure Dockhand Release Watch succeeded on the current main SHA."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import RELEASE_WATCH_WORKFLOW, _gh_json, _repo, workflow_green_on_main_sha

DEFAULT_TIMEOUT_SEC = 2700
POLL_INTERVAL_SEC = 30


def _run(args: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        args,
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )


def current_main_sha() -> str:
    proc = _run(["gh", "api", f"repos/{_repo()}/commits/main", "--jq", ".sha"])
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "failed to resolve main SHA")
    sha = proc.stdout.strip()
    if not sha:
        raise RuntimeError("empty main SHA")
    return sha


def dispatch_release_watch() -> int:
    proc = _run(
        [
            "gh",
            "workflow",
            "run",
            "dockhand-release-watch.yml",
            "--repo",
            _repo(),
            "--ref",
            "main",
        ]
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "workflow dispatch failed")
    time.sleep(10)
    data = _gh_json(
        [
            "run",
            "list",
            "--workflow",
            RELEASE_WATCH_WORKFLOW,
            "--branch",
            "main",
            "--limit",
            "1",
            "--json",
            "databaseId",
        ]
    )
    if not isinstance(data, list) or not data:
        raise RuntimeError("could not resolve Release Watch run id after dispatch")
    return int(data[0]["databaseId"])


def wait_for_run(run_id: int, *, timeout_sec: int) -> str:
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        proc = _run(
            [
                "gh",
                "run",
                "view",
                str(run_id),
                "--repo",
                _repo(),
                "--json",
                "status,conclusion",
            ]
        )
        if proc.returncode != 0:
            raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh run view failed")
        payload = json.loads(proc.stdout)
        status = str(payload.get("status") or "")
        conclusion = str(payload.get("conclusion") or "")
        if status == "completed":
            return conclusion or "unknown"
        time.sleep(POLL_INTERVAL_SEC)
    raise TimeoutError(f"timed out waiting for Release Watch run {run_id}")


def ensure_green(*, retries: int = 1, timeout_sec: int = DEFAULT_TIMEOUT_SEC) -> dict:
    main_sha = current_main_sha()
    if workflow_green_on_main_sha(RELEASE_WATCH_WORKFLOW, main_sha):
        return {"main_sha": main_sha, "action": "already_green", "run_id": None, "conclusion": "success"}

    last_conclusion = ""
    last_run_id: int | None = None
    for attempt in range(retries + 1):
        last_run_id = dispatch_release_watch()
        last_conclusion = wait_for_run(last_run_id, timeout_sec=timeout_sec)
        if last_conclusion == "success" and workflow_green_on_main_sha(RELEASE_WATCH_WORKFLOW, main_sha):
            return {
                "main_sha": main_sha,
                "action": "dispatched",
                "run_id": last_run_id,
                "conclusion": last_conclusion,
                "attempt": attempt + 1,
            }

    raise RuntimeError(
        f"Release Watch did not succeed on main {main_sha} "
        f"(last run {last_run_id}, conclusion={last_conclusion})"
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--retries", type=int, default=1)
    parser.add_argument("--timeout-sec", type=int, default=DEFAULT_TIMEOUT_SEC)
    args = parser.parse_args(argv)

    try:
        result = ensure_green(retries=max(0, args.retries), timeout_sec=args.timeout_sec)
    except (RuntimeError, TimeoutError) as err:
        print(str(err), file=sys.stderr)
        return 1

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
