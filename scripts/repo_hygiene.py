#!/usr/bin/env python3
"""Repo hygiene: stale agent branches, closed automation issues, old workflow runs."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import time
from collections import defaultdict
from datetime import datetime, timedelta, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

DEFAULT_RETAIN_DAYS = 14
DEFAULT_KEEP_MINIMUM_RUNS = 1

# Closed issues matching these title patterns may be deleted when --delete-closed-issues.
DELETE_CLOSED_TITLE_RES = (
    re.compile(r"^agent smoke:", re.I),
    re.compile(r"^Compatibility failure: dockhand v1\.0\.3$", re.I),
    re.compile(r"^\[Automation\] Workflow failing:", re.I),
    re.compile(r"^CI failure:", re.I),
    re.compile(r"^API drift detected: dockhand latest$", re.I),
    re.compile(r"^release: prepare v", re.I),
)

KEEP_BRANCH_NAMES = frozenset()


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


def _parse_run_time(raw: str) -> datetime:
    value = raw.strip()
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    parsed = datetime.fromisoformat(value)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def list_all_workflow_runs() -> list[dict]:
    runs: list[dict] = []
    page = 1
    while True:
        data = _gh_json(
            ["api", f"repos/{_repo()}/actions/runs?per_page=100&page={page}"]
        )
        if not isinstance(data, dict):
            break
        batch = data.get("workflow_runs")
        if not isinstance(batch, list) or not batch:
            break
        runs.extend(batch)
        if len(batch) < 100:
            break
        page += 1
    return runs


def workflow_run_key(run: dict) -> str:
    workflow_id = run.get("workflow_id")
    if workflow_id:
        return f"id:{workflow_id}"
    name = str(run.get("name") or "").strip()
    path = str(run.get("path") or "").strip()
    if path:
        return f"path:{path}"
    return f"name:{name or 'unknown'}"


def select_workflow_runs_to_delete(
    runs: list[dict],
    *,
    retain_days: int,
    keep_minimum: int,
    now: datetime | None = None,
) -> list[dict]:
    now = now or datetime.now(timezone.utc)
    cutoff = now - timedelta(days=retain_days)
    grouped: dict[str, list[dict]] = defaultdict(list)
    for run in runs:
        grouped[workflow_run_key(run)].append(run)

    to_delete: list[dict] = []
    for group in grouped.values():
        group.sort(
            key=lambda item: _parse_run_time(str(item.get("created_at", ""))),
            reverse=True,
        )
        keep_ids: set[int] = set()
        for item in group[: max(keep_minimum, 0)]:
            keep_ids.add(int(item["id"]))
        for item in group:
            created = _parse_run_time(str(item.get("created_at", "")))
            if created >= cutoff:
                keep_ids.add(int(item["id"]))
        for item in group:
            if int(item["id"]) not in keep_ids:
                to_delete.append(item)
    to_delete.sort(key=lambda item: _parse_run_time(str(item.get("created_at", ""))))
    return to_delete


def delete_workflow_run(run_id: int, *, apply: bool) -> bool:
    if not apply:
        return True
    cmd = ["gh", "api", "-X", "DELETE", f"repos/{_repo()}/actions/runs/{run_id}"]
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True, check=False)
    if proc.returncode == 0:
        return True
    err = (proc.stderr or proc.stdout or "").strip()
    if "404" in err or "Not Found" in err:
        return True
    raise RuntimeError(err or f"failed to delete workflow run {run_id}")


def prune_workflow_runs(
    *,
    retain_days: int,
    keep_minimum: int,
    apply: bool,
    pause_ms: int = 0,
    quiet: bool = False,
) -> dict[str, object]:
    runs = list_all_workflow_runs()
    targets = select_workflow_runs_to_delete(
        runs,
        retain_days=retain_days,
        keep_minimum=keep_minimum,
    )
    deleted: list[int] = []
    for run in targets:
        run_id = int(run["id"])
        name = str(run.get("name") or run.get("path") or "unknown")
        created = str(run.get("created_at", ""))[:10]
        if apply:
            delete_workflow_run(run_id, apply=True)
            deleted.append(run_id)
            if len(deleted) % 100 == 0:
                print(
                    f"deleted {len(deleted)}/{len(targets)} workflow runs...",
                    file=sys.stderr,
                    flush=True,
                )
            if pause_ms > 0:
                time.sleep(pause_ms / 1000.0)
        elif not quiet:
            print(f"would delete run {run_id} ({name}, {created})")
    return {
        "total_runs": len(runs),
        "delete_candidates": len(targets),
        "deleted_run_ids": deleted if apply else [],
        "retain_days": retain_days,
        "keep_minimum_runs": keep_minimum,
    }


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
    parser.add_argument(
        "--workflow-runs",
        action="store_true",
        help="Delete old workflow runs (see --retain-days and --keep-minimum-runs)",
    )
    parser.add_argument(
        "--retain-days",
        type=int,
        default=DEFAULT_RETAIN_DAYS,
        help=f"Keep all workflow runs newer than this many days (default {DEFAULT_RETAIN_DAYS})",
    )
    parser.add_argument(
        "--keep-minimum-runs",
        type=int,
        default=DEFAULT_KEEP_MINIMUM_RUNS,
        help=f"Always keep this many newest runs per workflow (default {DEFAULT_KEEP_MINIMUM_RUNS})",
    )
    parser.add_argument(
        "--delete-pause-ms",
        type=int,
        default=0,
        help="Optional pause between workflow run deletions (rate-limit safety)",
    )
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    if not args.branches and not args.delete_closed_issues and not args.workflow_runs:
        args.branches = True

    report: dict[str, object] = {"apply": args.apply, "branches": [], "issues": [], "workflow_runs": {}}

    if args.branches:
        for branch in list_agent_branches():
            if delete_branch(branch, apply=args.apply):
                report["branches"].append(branch)

    if args.delete_closed_issues:
        for item in list_closed_delete_candidates():
            delete_issue(item["number"], apply=args.apply)
            report["issues"].append(item)

    if args.workflow_runs:
        report["workflow_runs"] = prune_workflow_runs(
            retain_days=args.retain_days,
            keep_minimum=args.keep_minimum_runs,
            apply=args.apply,
            pause_ms=args.delete_pause_ms,
            quiet=args.json,
        )
        if args.apply:
            print(
                "deleted "
                f"{report['workflow_runs']['delete_candidates']} workflow run(s); "
                f"{report['workflow_runs']['total_runs']} existed before cleanup"
            )

    if args.json:
        print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        raise SystemExit(1) from err
