#!/usr/bin/env python3
"""Re-dispatch Issue Agent Intake for CI failure issues that never left the queue."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

CI_ISSUE_PREFIXES = ("CI failure:", "Security CI failure:")


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


def label_names(raw_labels: object) -> set[str]:
    if not isinstance(raw_labels, list):
        return set()
    names: set[str] = set()
    for item in raw_labels:
        if isinstance(item, dict):
            names.add(str(item.get("name", "")).strip().lower())
        else:
            names.add(str(item).strip().lower())
    return {name for name in names if name}


def is_ci_failure_title(title: str) -> bool:
    text = (title or "").strip()
    return any(text.startswith(prefix) for prefix in CI_ISSUE_PREFIXES)


def dispatch_intake(issue_number: int) -> None:
    proc = subprocess.run(
        [
            "gh",
            "workflow",
            "run",
            "issue-agent-intake.yml",
            "--repo",
            _repo(),
            "-f",
            f"issue_number={issue_number}",
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "workflow dispatch failed")


def find_undispatched_ci_issues() -> list[dict]:
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from issue_agent_intake_eligibility import intake_eligible

    data = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "open",
            "--label",
            "agent",
            "--json",
            "number,title,labels,body,author",
            "--limit",
            "100",
        ]
    )
    if not isinstance(data, list):
        return []

    pending: list[dict] = []
    for issue in data:
        number = int(issue["number"])
        title = str(issue.get("title", ""))
        if not is_ci_failure_title(title):
            continue

        labels = label_names(issue.get("labels"))
        if "agent-dispatched" in labels or "no-agent" in labels:
            continue
        if not (labels & {"ci", "security"}):
            continue

        author = issue.get("author") if isinstance(issue.get("author"), dict) else {}
        eligible, reason = intake_eligible(
            title=title,
            labels=sorted(labels),
            author_type=str(author.get("__typename", author.get("type", "Bot"))),
            author_login=str(author.get("login", "github-actions[bot]")),
            trigger="opened",
            body=str(issue.get("body", "")),
            has_dispatched=False,
            has_regression=False,
        )
        if not eligible:
            continue

        pending.append({"number": number, "title": title, "labels": sorted(labels)})

    return pending


def redispatch(*, dry_run: bool = False) -> list[dict]:
    results: list[dict] = []
    for item in find_undispatched_ci_issues():
        entry = dict(item)
        if dry_run:
            entry["intake_dispatch"] = "dry-run"
        else:
            try:
                dispatch_intake(int(item["number"]))
                entry["intake_dispatch"] = "ok"
            except RuntimeError as err:
                entry["intake_dispatch"] = str(err)
        results.append(entry)
    return results


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args(argv)

    try:
        results = redispatch(dry_run=args.dry_run)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    failed = [item for item in results if item.get("intake_dispatch") not in {"ok", "dry-run"}]
    payload = {
        "issues": results,
        "count": len(results),
        "dispatch_failures": len(failed),
    }
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
