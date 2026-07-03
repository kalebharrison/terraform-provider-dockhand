#!/usr/bin/env python3
"""Evaluate automated release gate readiness (GitHub CLI required)."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
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
    "Acceptance Full",
    "Dockhand Release Watch",
)


@dataclass
class GateResult:
    ci_gates_pass: bool = False
    version: str = ""
    tag: str = ""
    tier: str = "patch"
    blockers: list[str] = field(default_factory=list)
    awaiting_release_issues: list[int] = field(default_factory=list)
    lens_verdict_clear: bool = False
    open_release_issue: int | None = None

    @property
    def ready_for_lens_dispatch(self) -> bool:
        return (
            self.ci_gates_pass
            and not self.lens_verdict_clear
            and self.open_release_issue is None
            and bool(self.awaiting_release_issues)
            and bool(self.version)
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
            "lens_verdict_clear": self.lens_verdict_clear,
            "open_release_issue": self.open_release_issue,
        }


def _gh_json(args: list[str]) -> object:
    cmd = ["gh", *args, "--repo", _repo()]
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
        ["gh", "release", "view", tag, "--repo", _repo()],
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
    query = f"repo:{_repo()} is:issue label:awaiting-release -label:released"
    data = _gh_json(["search", "issues", query, "--json", "number", "--limit", "100"])
    if not isinstance(data, list):
        return []
    return [int(item["number"]) for item in data]


def workflow_green(name: str) -> bool:
    data = _gh_json(
        [
            "run",
            "list",
            "--workflow",
            name,
            "--branch",
            "main",
            "--limit",
            "1",
            "--json",
            "conclusion,status",
        ]
    )
    if not isinstance(data, list) or not data:
        return False
    run = data[0]
    return run.get("status") == "completed" and run.get("conclusion") == "success"


def find_open_release_issue(version: str) -> int | None:
    query = f'repo:{_repo()} is:issue is:open in:title "release: prepare v{version}"'
    data = _gh_json(["search", "issues", query, "--json", "number", "--limit", "1"])
    if isinstance(data, list) and data:
        return int(data[0]["number"])
    return None


def evaluate_gate() -> GateResult:
    result = GateResult()

    for issue_number in open_compatibility_issues():
        result.blockers.append(f"open compatibility issue #{issue_number}")

    workflow_failures = [name for name in REQUIRED_WORKFLOWS if not workflow_green(name)]
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
    result.open_release_issue = find_open_release_issue(result.version)

    review_log = ROOT / "docs/reports/agent-review-log.md"
    if review_log.exists():
        result.lens_verdict_clear = is_cleared_for_version(
            review_log.read_text(encoding="utf-8"),
            result.version,
        )

    result.ci_gates_pass = not result.blockers
    return result


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
        result = evaluate_gate()
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
