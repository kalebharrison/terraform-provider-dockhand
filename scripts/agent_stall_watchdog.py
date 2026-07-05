#!/usr/bin/env python3
"""Find open agent-dispatched issues with stalled Cloud Agent progress."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
STALL_HOURS = 24
DISPATCH_MARKER = "## Agent dispatched"


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


def _parse_time(raw: str) -> datetime:
    return datetime.fromisoformat(raw.replace("Z", "+00:00")).astimezone(timezone.utc)


def _hours_since(when: datetime, now: datetime) -> float:
    return max((now - when.astimezone(timezone.utc)).total_seconds() / 3600.0, 0.0)


def main_sha() -> str:
    data = _gh_json(["api", f"repos/{_repo()}/git/ref/heads/main"])
    if not isinstance(data, dict):
        raise RuntimeError("could not resolve main ref")
    return str(data["object"]["sha"])


def matching_branch(issue_number: int) -> str | None:
    data = _gh_json(["api", f"repos/{_repo()}/git/matching-refs/heads/agent/issue-{issue_number}-"])
    if not isinstance(data, list) or not data:
        return None
    ref = str(data[-1].get("ref", ""))
    if not ref.startswith("refs/heads/"):
        return None
    return ref.removeprefix("refs/heads/")


def branch_tip_sha(branch: str) -> str:
    data = _gh_json(["api", f"repos/{_repo()}/git/ref/heads/{branch}"])
    if not isinstance(data, dict):
        raise RuntimeError(f"could not resolve branch {branch}")
    return str(data["object"]["sha"])


def commit_time(sha: str) -> datetime:
    data = _gh_json(["api", f"repos/{_repo()}/commits/{sha}"])
    if not isinstance(data, dict):
        raise RuntimeError(f"could not resolve commit time for {sha}")
    return _parse_time(str(data["commit"]["committer"]["date"]))


def commits_ahead(base_sha: str, head_sha: str) -> int:
    data = _gh_json(["api", f"repos/{_repo()}/compare/{base_sha}...{head_sha}"])
    if not isinstance(data, dict):
        return 0
    return int(data.get("ahead_by", 0))


def has_open_pr(issue_number: int) -> bool:
    data = _gh_json(["pr", "list", "--state", "open", "--json", "headRefName", "--limit", "100"])
    if not isinstance(data, list):
        return False
    prefix = f"agent/issue-{issue_number}-"
    return any(str(item.get("headRefName", "")).startswith(prefix) for item in data)


def last_dispatch_time(issue_number: int) -> datetime | None:
    data = _gh_json(["api", f"repos/{_repo()}/issues/{issue_number}/comments"])
    if not isinstance(data, list):
        return None
    for comment in reversed(data):
        body = str(comment.get("body", ""))
        if DISPATCH_MARKER in body:
            created = comment.get("created_at")
            if created:
                return _parse_time(str(created))
    return None


def find_stalled(*, now: datetime | None = None) -> list[dict]:
    now = now or datetime.now(timezone.utc)
    base = main_sha()
    issues = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "open",
            "--label",
            "agent-dispatched",
            "--json",
            "number,title,updatedAt",
            "--limit",
            "100",
        ]
    )
    if not isinstance(issues, list):
        return []

    stalled: list[dict] = []
    for issue in issues:
        number = int(issue["number"])
        if has_open_pr(number):
            continue

        branch = matching_branch(number)
        dispatch_time = last_dispatch_time(number)
        if dispatch_time is None and issue.get("updatedAt"):
            dispatch_time = _parse_time(str(issue["updatedAt"]))

        if branch is None:
            if dispatch_time and _hours_since(dispatch_time, now) >= STALL_HOURS:
                stalled.append(
                    {
                        "number": number,
                        "title": issue.get("title", ""),
                        "branch": None,
                        "reason": "no agent branch after dispatch",
                    }
                )
            continue

        head = branch_tip_sha(branch)
        ahead = commits_ahead(base, head)
        if ahead == 0:
            if dispatch_time and _hours_since(dispatch_time, now) >= STALL_HOURS:
                stalled.append(
                    {
                        "number": number,
                        "title": issue.get("title", ""),
                        "branch": branch,
                        "reason": "agent branch has no commits beyond main",
                    }
                )
            continue

        last_activity = commit_time(head)
        if _hours_since(last_activity, now) >= STALL_HOURS:
            stalled.append(
                {
                    "number": number,
                    "title": issue.get("title", ""),
                    "branch": branch,
                    "reason": f"no agent activity on {branch} for {STALL_HOURS}h",
                }
            )
    return stalled


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    try:
        result = {"stalled": find_stalled(), "stall_hours": STALL_HOURS}
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
