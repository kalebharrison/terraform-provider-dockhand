#!/usr/bin/env python3
"""One-shot and scheduled repo hygiene: stale agent branches, closed automation issues."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# Closed issues matching these title patterns may be deleted when --delete-closed-issues.
DELETE_CLOSED_TITLE_RES = (
    re.compile(r"^agent smoke:", re.I),
    re.compile(r"^Compatibility failure: dockhand v1\.0\.3$", re.I),
    re.compile(r"^\[Automation\] Workflow failing:", re.I),
    re.compile(r"^CI failure:", re.I),
    re.compile(r"^API drift detected: dockhand latest$", re.I),
    re.compile(r"^release: prepare v", re.I),
)

KEEP_BRANCH_NAMES = frozenset({"agent/compat-reports-sync"})


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


def list_agent_branches() -> list[str]:
    data = _gh_json(["api", f"repos/{_repo()}/git/matching-refs/heads/agent/"])
    if not isinstance(data, list):
        return []
    out: list[str] = []
    for item in data:
        ref = str(item.get("ref", ""))
        if ref.startswith("refs/heads/"):
            out.append(ref.removeprefix("refs/heads/"))
    return sorted(out)


def branch_has_open_pr(branch: str) -> bool:
    data = _gh_json(["pr", "list", "--state", "open", "--head", branch, "--json", "number"])
    return isinstance(data, list) and len(data) > 0


def _gh_run(args: list[str]) -> None:
    cmd = ["gh", *args]
    if not args or args[0] != "api":
        cmd.extend(["--repo", _repo()])
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True, check=False)
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh failed")


def delete_branch(branch: str, *, apply: bool) -> bool:
    if branch in KEEP_BRANCH_NAMES:
        return False
    if branch_has_open_pr(branch):
        return False
    if not apply:
        print(f"would delete branch {branch}")
        return True
    _gh_run(["api", "-X", "DELETE", f"repos/{_repo()}/git/refs/heads/{branch}"])
    print(f"deleted branch {branch}")
    return True


def list_closed_delete_candidates() -> list[dict]:
    issues = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "closed",
            "--limit",
            "200",
            "--json",
            "number,title,labels",
        ]
    )
    if not isinstance(issues, list):
        return []
    out: list[dict] = []
    for issue in issues:
        title = str(issue.get("title", ""))
        labels = {str(l.get("name", "")).lower() for l in issue.get("labels", [])}
        if "security" in labels or "pinned" in labels:
            continue
        if not any(p.search(title) for p in DELETE_CLOSED_TITLE_RES):
            continue
        out.append({"number": int(issue["number"]), "title": title})
    return out


def delete_issue(number: int, *, apply: bool) -> None:
    if not apply:
        print(f"would delete issue #{number}")
        return
    proc = subprocess.run(
        ["gh", "issue", "delete", str(number), "--yes", "--repo", _repo()],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or f"failed to delete #{number}")
    print(f"deleted issue #{number}")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--apply",
        action="store_true",
        help="Perform deletes (default is dry-run)",
    )
    parser.add_argument(
        "--branches",
        action="store_true",
        help="Delete stale agent/* branches without open PRs",
    )
    parser.add_argument(
        "--delete-closed-issues",
        action="store_true",
        help="Delete closed automation/smoke/drift tracker issues (see script patterns)",
    )
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    if not args.branches and not args.delete_closed_issues:
        args.branches = True

    report: dict[str, object] = {"apply": args.apply, "branches": [], "issues": []}

    if args.branches:
        for branch in list_agent_branches():
            if delete_branch(branch, apply=args.apply):
                report["branches"].append(branch)

    if args.delete_closed_issues:
        for item in list_closed_delete_candidates():
            delete_issue(item["number"], apply=args.apply)
            report["issues"].append(item)

    if args.json:
        print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        raise SystemExit(1) from err
