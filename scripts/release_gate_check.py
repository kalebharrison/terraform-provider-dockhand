#!/usr/bin/env python3
"""Evaluate automated release gate readiness (GitHub CLI required)."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_semver import Version, release_tier
from release_verdict import is_cleared_for_version

REQUIRED_WORKFLOWS = (
    "Go CI",
    "Govulncheck",
    "Workflow Lint",
    "Gitleaks",
    "Dependency Review",
    "Acceptance Full",
    "Dockhand Release Watch",
)

RELEASE_WATCH_WORKFLOW = "Dockhand Release Watch"
RELEASE_WATCH_VALIDATE_JOB = "Validate Provider Against Dockhand Release"


@dataclass
class GateResult:
    ci_gates_pass: bool = False
    version: str = ""
    tag: str = ""
    tier: str = "patch"
    blockers: list[str] = field(default_factory=list)
    awaiting_release_issues: list[int] = field(default_factory=list)
    unreleased_commits_on_main: int = 0
    lens_verdict_clear: bool = False
    open_release_issue: int | None = None

    @property
    def has_release_work(self) -> bool:
        return bool(self.awaiting_release_issues) or self.unreleased_commits_on_main > 0

    @property
    def release_trigger(self) -> str | None:
        if self.awaiting_release_issues:
            return "awaiting_release_issues"
        if self.unreleased_commits_on_main > 0:
            return "unreleased_main_commits"
        return None

    @property
    def ready_for_lens_dispatch(self) -> bool:
        return (
            self.ci_gates_pass
            and not self.lens_verdict_clear
            and self.open_release_issue is None
            and bool(self.version)
            and self.has_release_work
        )

    @property
    def ready_to_tag(self) -> bool:
        return self.ci_gates_pass and self.lens_verdict_clear and bool(self.tag)

    def as_json(self) -> dict:
        return {
            "ci_gates_pass": self.ci_gates_pass,
            "ready_for_lens_dispatch": self.ready_for_lens_dispatch,
            "ready_to_tag": self.ready_to_tag,
            "version": self.version,
            "tag": self.tag,
            "tier": self.tier,
            "blockers": self.blockers,
            "awaiting_release_issues": self.awaiting_release_issues,
            "unreleased_commits_on_main": self.unreleased_commits_on_main,
            "has_release_work": self.has_release_work,
            "release_trigger": self.release_trigger,
            "lens_verdict_clear": self.lens_verdict_clear,
            "open_release_issue": self.open_release_issue,
        }


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


def latest_release_tag() -> Version | None:
    data = _gh_json(["release", "list", "--limit", "20", "--json", "tagName,isDraft"])
    if not isinstance(data, list):
        return None
    for item in data:
        if item.get("isDraft"):
            continue
        version = Version.parse(str(item.get("tagName", "")))
        if version:
            return version
    return None


def draft_release_tag() -> Version | None:
    data = _gh_json(["release", "list", "--limit", "20", "--json", "tagName,isDraft"])
    if not isinstance(data, list):
        return None
    for item in data:
        if not item.get("isDraft"):
            continue
        version = Version.parse(str(item.get("tagName", "")))
        if version:
            return version
    return None


def tag_exists(tag: str) -> bool:
    proc = subprocess.run(
        ["gh", "api", f"repos/{_repo()}/git/refs/tags/{tag}"],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    return proc.returncode == 0


def open_compatibility_issues() -> list[int]:
    data = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "open",
            "--label",
            "compatibility",
            "--json",
            "number,title",
            "--limit",
            "100",
        ]
    )
    if not isinstance(data, list):
        return []
    numbers: list[int] = []
    for item in data:
        title = str(item.get("title", "")).lower()
        if title.startswith("release:"):
            continue
        numbers.append(int(item["number"]))
    return numbers


def awaiting_release_issues() -> list[int]:
    data = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "open",
            "--label",
            "awaiting-release",
            "--json",
            "number,labels",
            "--limit",
            "100",
        ]
    )
    if not isinstance(data, list):
        return []
    numbers: list[int] = []
    for item in data:
        labels = {label.get("name", "") for label in item.get("labels", [])}
        if "released" in labels:
            continue
        numbers.append(int(item["number"]))
    return numbers


def commits_ahead_of_latest_release() -> int:
    latest = latest_release_tag()
    if latest is None:
        proc = subprocess.run(
            ["gh", "api", f"repos/{_repo()}/commits/main", "--jq", ".sha"],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        return 1 if proc.returncode == 0 and proc.stdout.strip() else 0

    data = _gh_json(["api", f"repos/{_repo()}/compare/{latest.tag()}...main"])
    if not isinstance(data, dict):
        return 0
    return max(int(data.get("ahead_by", 0)), 0)


def workflow_runs(name: str, *, limit: int = 15) -> list[dict]:
    data = _gh_json(
        [
            "run",
            "list",
            "--workflow",
            name,
            "--branch",
            "main",
            "--limit",
            str(limit),
            "--json",
            "conclusion,status,headSha,createdAt,databaseId",
        ]
    )
    if not isinstance(data, list):
        return []
    return [item for item in data if isinstance(item, dict)]


def release_watch_run_validated(run: dict) -> bool:
    run_id = run.get("databaseId")
    if not run_id:
        return False
    proc = subprocess.run(
        [
            "gh",
            "run",
            "view",
            str(run_id),
            "--repo",
            _repo(),
            "--json",
            "jobs",
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        return False
    payload = json.loads(proc.stdout)
    jobs = payload.get("jobs") if isinstance(payload, dict) else None
    if not isinstance(jobs, list):
        return False
    for job in jobs:
        if not isinstance(job, dict):
            continue
        if job.get("name") != RELEASE_WATCH_VALIDATE_JOB:
            continue
        return job.get("status") == "completed" and job.get("conclusion") == "success"
    return False


def release_watch_chain_green() -> bool:
    """True when the latest Release Watch run succeeded and a validated run exists
    in the current skip-only chain (unchanged Dockhand tag) without failures."""
    for run in workflow_runs(RELEASE_WATCH_WORKFLOW, limit=30):
        if run.get("status") != "completed":
            return False
        if run.get("conclusion") != "success":
            return False
        if release_watch_run_validated(run):
            return True
    return False


def release_watch_strict_green() -> bool:
    return release_watch_chain_green()


def release_watch_main_sha_green(main_sha: str | None = None) -> bool:
    _ = main_sha
    return release_watch_chain_green()


def current_main_sha() -> str:
    proc = subprocess.run(
        ["gh", "api", f"repos/{_repo()}/commits/main", "--jq", ".sha"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "failed to resolve main SHA")
    sha = proc.stdout.strip()
    if not sha:
        raise RuntimeError("empty main SHA")
    return sha


def workflow_green(name: str) -> bool:
    if name == RELEASE_WATCH_WORKFLOW:
        return release_watch_strict_green()
    data = workflow_runs(name, limit=1)
    if not data:
        return False
    run = data[0]
    return run.get("status") == "completed" and run.get("conclusion") == "success"


def workflow_green_on_main_sha(name: str, main_sha: str | None = None) -> bool:
    if name == RELEASE_WATCH_WORKFLOW:
        try:
            return release_watch_main_sha_green(main_sha)
        except RuntimeError:
            return False
    sha = main_sha or current_main_sha()
    for run in workflow_runs(name, limit=20):
        if run.get("status") != "completed":
            continue
        if run.get("headSha") != sha:
            continue
        if run.get("conclusion") == "success":
            return True
    return False


def workflow_is_green(name: str, *, release_watch_mode: str = "strict") -> bool:
    if name == RELEASE_WATCH_WORKFLOW and release_watch_mode == "main_sha":
        try:
            return workflow_green_on_main_sha(name)
        except RuntimeError:
            return False
    return workflow_green(name)


def find_open_release_issue(version: str) -> int | None:
    query = f'repo:{_repo()} is:issue is:open in:title "release: prepare v{version}"'
    data = _gh_json(["search", "issues", query, "--json", "number", "--limit", "1"])
    if isinstance(data, list) and data:
        return int(data[0]["number"])
    return None


def evaluate_gate(*, release_watch_mode: str = "strict") -> GateResult:
    result = GateResult()

    for issue_number in open_compatibility_issues():
        result.blockers.append(f"open compatibility issue #{issue_number}")

    workflow_failures = [
        name
        for name in REQUIRED_WORKFLOWS
        if not workflow_is_green(name, release_watch_mode=release_watch_mode)
    ]
    for name in workflow_failures:
        result.blockers.append(f"workflow not green on main: {name}")

    draft = draft_release_tag()
    if draft is None:
        result.blockers.append("no draft release from Release Drafter")
        return result

    if tag_exists(draft.tag()):
        result.blockers.append(f"tag already published: {draft.tag()}")
        return result

    previous = latest_release_tag()
    result.version = str(draft)
    result.tag = draft.tag()
    result.tier = release_tier(previous, draft)
    result.awaiting_release_issues = awaiting_release_issues()
    result.unreleased_commits_on_main = commits_ahead_of_latest_release()
    result.open_release_issue = find_open_release_issue(result.version)

    review_log = ROOT / "docs/reports/agent-review-log.md"
    if review_log.exists():
        result.lens_verdict_clear = is_cleared_for_version(
            review_log.read_text(encoding="utf-8"),
            result.version,
        )

    result.ci_gates_pass = not result.blockers
    return result


RELEASE_WATCH_ENSURE_TIMEOUT_SEC = 2700
RELEASE_WATCH_ENSURE_POLL_SEC = 30


def release_watch_needs_force_validate() -> bool:
    """True when discover would skip validation (unchanged Dockhand) but the chain is red."""
    if release_watch_chain_green():
        return False
    runs = workflow_runs(RELEASE_WATCH_WORKFLOW, limit=5)
    if not runs:
        return True
    latest = runs[0]
    if latest.get("status") != "completed":
        return False
    return latest.get("conclusion") == "failure"


def dispatch_release_watch(*, force: bool | None = None) -> int:
    if force is None:
        force = release_watch_needs_force_validate()
    cmd = [
        "gh",
        "workflow",
        "run",
        "dockhand-release-watch.yml",
        "--repo",
        _repo(),
        "--ref",
        "main",
    ]
    if force:
        cmd.extend(["-f", "force_validate=true"])
    proc = subprocess.run(
        cmd,
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
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


def wait_for_release_watch_run(run_id: int, *, timeout_sec: int) -> str:
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        proc = subprocess.run(
            [
                "gh",
                "run",
                "view",
                str(run_id),
                "--repo",
                _repo(),
                "--json",
                "status,conclusion",
            ],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if proc.returncode != 0:
            raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or "gh run view failed")
        payload = json.loads(proc.stdout)
        status = str(payload.get("status") or "")
        conclusion = str(payload.get("conclusion") or "")
        if status == "completed":
            return conclusion or "unknown"
        time.sleep(RELEASE_WATCH_ENSURE_POLL_SEC)
    raise TimeoutError(f"timed out waiting for Release Watch run {run_id}")


def ensure_release_watch_green(*, retries: int = 1, timeout_sec: int = RELEASE_WATCH_ENSURE_TIMEOUT_SEC) -> dict:
    main_sha = current_main_sha()
    if release_watch_main_sha_green(main_sha):
        return {"main_sha": main_sha, "action": "already_green", "run_id": None, "conclusion": "success"}

    last_conclusion = ""
    last_run_id: int | None = None
    for attempt in range(retries + 1):
        force = attempt > 0 or release_watch_needs_force_validate()
        last_run_id = dispatch_release_watch(force=force)
        last_conclusion = wait_for_release_watch_run(last_run_id, timeout_sec=timeout_sec)
        if last_conclusion == "success" and release_watch_main_sha_green(main_sha):
            return {
                "main_sha": main_sha,
                "action": "dispatched",
                "run_id": last_run_id,
                "conclusion": last_conclusion,
                "attempt": attempt + 1,
                "force_validate": force,
            }

    raise RuntimeError(
        f"Release Watch did not validate on main {main_sha} "
        f"(last run {last_run_id}, conclusion={last_conclusion})"
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument(
        "--mode",
        choices=("lens", "tag", "status"),
        default="status",
        help="lens=ready to open release issue; tag=ready to publish tag",
    )
    args = parser.parse_args(argv)

    try:
        release_watch_mode = "main_sha" if args.mode == "tag" else "strict"
        result = evaluate_gate(release_watch_mode=release_watch_mode)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    payload = result.as_json()
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))

    if args.mode == "lens":
        return 0 if result.ready_for_lens_dispatch else 1
    if args.mode == "tag":
        return 0 if result.ready_to_tag else 1
    return 0 if result.ci_gates_pass else 1


if __name__ == "__main__":
    raise SystemExit(main())
