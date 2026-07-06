#!/usr/bin/env python3
"""Inspect a GitHub Actions workflow run for a named job conclusion."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
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


def list_jobs(run_id: int) -> list[dict]:
    jobs: list[dict] = []
    page = 1
    while True:
        proc = subprocess.run(
            [
                "gh",
                "api",
                f"repos/{_repo()}/actions/runs/{run_id}/jobs?per_page=100&page={page}",
            ],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if proc.returncode != 0:
            raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh api failed")
        payload = json.loads(proc.stdout)
        if not isinstance(payload, dict):
            break
        batch = payload.get("jobs")
        if not isinstance(batch, list) or not batch:
            break
        jobs.extend(batch)
        if len(batch) < 100:
            break
        page += 1
    return jobs


def job_conclusion(run_id: int, job_name: str) -> str | None:
    matches = [job for job in list_jobs(run_id) if str(job.get("name", "")) == job_name]
    if not matches:
        return None
    matches.sort(key=lambda item: str(item.get("completed_at") or ""), reverse=True)
    conclusion = str(matches[0].get("conclusion") or "").strip()
    return conclusion or None


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run-id", type=int, required=True)
    parser.add_argument("--job-name", required=True)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    try:
        conclusion = job_conclusion(args.run_id, args.job_name)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    payload = {
        "run_id": args.run_id,
        "job_name": args.job_name,
        "conclusion": conclusion,
        "found": conclusion is not None,
        "succeeded": conclusion == "success",
    }
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
