#!/usr/bin/env python3
"""Recover Agent Merge Cleanup after GITHUB_TOKEN squash-merges.

Bot merges do not trigger ``push`` / ``pull_request`` workflows, so linked
compatibility issues can stay open and block the release gate. A scheduled
watchdog (trusted trigger) can dispatch merge cleanup by PR number.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

HEAD_PREFIXES = ("agent/",)
AUTOMATION_PREFIXES = ("automation/",)
CURSOR_HEAD_PREFIX = "cursor/"
CURSOR_AUTHORS = frozenset({"app/cursor", "cursor[bot]"})
CLOSE_RE = re.compile(r"\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+#(\d+)\b", re.I)
CLEANUP_MARKER = "Posted by Agent Merge Cleanup"
LOOKBACK_HOURS = 72


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


def _parse_time(raw: str) -> datetime | None:
    text = (raw or "").strip()
    if not text:
        return None
    if text.endswith("Z"):
        text = text[:-1] + "+00:00"
    try:
        parsed = datetime.fromisoformat(text)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        return parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def linked_issue_numbers(body: str) -> list[int]:
    return sorted({int(match) for match in CLOSE_RE.findall(body or "")})


def is_cleanup_target_pr(item: dict) -> bool:
    head_ref = str(item.get("headRefName", "")).strip()
    if any(head_ref.startswith(prefix) for prefix in HEAD_PREFIXES):
        return True
    if any(head_ref.startswith(prefix) for prefix in AUTOMATION_PREFIXES):
        return False
    if not head_ref.startswith(CURSOR_HEAD_PREFIX):
        return False
    author = item.get("author") if isinstance(item.get("author"), dict) else {}
    return str(author.get("login", "")).strip().lower() in CURSOR_AUTHORS


def issue_needs_cleanup(issue: dict, comments: list[dict]) -> bool:
    state = str(issue.get("state", "")).lower()
    labels = {
        (label.get("name") if isinstance(label, dict) else str(label))
        for label in (issue.get("labels") or [])
    }
    notified = any(CLEANUP_MARKER in str(comment.get("body") or "") for comment in comments)
    if state != "closed":
        return True
    if "awaiting-release" not in labels and "released" not in labels:
        return True
    if not notified:
        return True
    return False


def recently_merged_agent_prs(*, lookback_hours: int = LOOKBACK_HOURS) -> list[dict]:
    data = _gh_json(
        [
            "pr",
            "list",
            "--state",
            "merged",
            "--limit",
            "40",
            "--json",
            "number,title,body,headRefName,mergedAt,author",
        ]
    )
    if not isinstance(data, list):
        return []
    cutoff = datetime.now(timezone.utc) - timedelta(hours=lookback_hours)
    out: list[dict] = []
    for item in data:
        if not isinstance(item, dict) or not is_cleanup_target_pr(item):
            continue
        merged_at = _parse_time(str(item.get("mergedAt") or ""))
        if merged_at is None or merged_at < cutoff:
            continue
        out.append(item)
    return out


def pr_needs_cleanup(pr: dict) -> bool:
    issues = linked_issue_numbers(str(pr.get("body") or ""))
    if not issues:
        return False
    for number in issues:
        issue = _gh_json(
            ["issue", "view", str(number), "--json", "number,state,labels,comments"]
        )
        if not isinstance(issue, dict):
            continue
        comment_items = [
            c for c in (issue.get("comments") or []) if isinstance(c, dict)
        ]
        if issue_needs_cleanup(issue, comment_items):
            return True
    return False


def dispatch_merge_cleanup(pull_number: int) -> None:
    proc = subprocess.run(
        [
            "gh",
            "workflow",
            "run",
            "agent-merge-cleanup.yml",
            "--repo",
            _repo(),
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


def recover_missed_cleanup(*, dispatch: bool = True, lookback_hours: int = LOOKBACK_HOURS) -> dict:
    recovered: list[dict] = []
    for pr in recently_merged_agent_prs(lookback_hours=lookback_hours):
        number = int(pr["number"])
        if not pr_needs_cleanup(pr):
            continue
        entry: dict = {"number": number, "title": pr.get("title", "")}
        if dispatch:
            try:
                dispatch_merge_cleanup(number)
                entry["dispatch"] = "ok"
            except RuntimeError as err:
                entry["dispatch"] = str(err)
        else:
            entry["dispatch"] = "skipped"
        recovered.append(entry)
    return {"pull_requests": recovered, "count": len(recovered)}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--no-dispatch", action="store_true")
    parser.add_argument("--lookback-hours", type=int, default=LOOKBACK_HOURS)
    args = parser.parse_args(argv)

    try:
        payload = recover_missed_cleanup(
            dispatch=not args.no_dispatch,
            lookback_hours=args.lookback_hours,
        )
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
