#!/usr/bin/env python3
"""Post-release housekeeping: label awaiting-release issues and cut CHANGELOG."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import _gh_json, _repo


def awaiting_release_without_released() -> list[int]:
    data = _gh_json(
        [
            "issue",
            "list",
            "--state",
            "all",
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


def label_issue_released(issue_number: int) -> None:
    subprocess.run(
        [
            "gh",
            "issue",
            "edit",
            str(issue_number),
            "--repo",
            _repo(),
            "--add-label",
            "released",
        ],
        cwd=ROOT,
        check=True,
    )
    try:
        subprocess.run(
            [
                "gh",
                "issue",
                "edit",
                str(issue_number),
                "--repo",
                _repo(),
                "--remove-label",
                "awaiting-release",
            ],
            cwd=ROOT,
            check=True,
        )
    except subprocess.CalledProcessError:
        pass


def cut_changelog(version: str) -> bool:
    changelog = ROOT / "CHANGELOG.md"
    if not changelog.exists():
        return False
    text = changelog.read_text(encoding="utf-8")
    version = version.lstrip("v")
    heading = f"## [{version}]"
    if heading in text:
        return False
    today = date.today().isoformat()
    replacement = f"## [Unreleased]\n\n### Added\n\n- _Nothing yet._\n\n{heading} - {today}"
    updated, count = re.subn(r"## \[Unreleased\]", replacement, text, count=1)
    if count != 1:
        return False
    changelog.write_text(updated, encoding="utf-8")
    return True


CHANGELOG_BRANCH = "automation/release-changelog"


def commit_changelog(version: str) -> dict[str, object]:
    """Commit CHANGELOG cut; push main, or open a PR if main is protected."""
    changelog = ROOT / "CHANGELOG.md"
    if not changelog.exists():
        return {"pushed_main": False, "pull_request": None}
    version = version.lstrip("v")
    title = f"chore(release): cut CHANGELOG for v{version}"
    subprocess.run(["git", "add", str(changelog)], cwd=ROOT, check=True)
    subprocess.run(
        [
            "git",
            "-c",
            "user.name=github-actions[bot]",
            "-c",
            "user.email=41898282+github-actions[bot]@users.noreply.github.com",
            "commit",
            "-m",
            title,
        ],
        cwd=ROOT,
        check=True,
    )
    push_main = subprocess.run(
        ["git", "push", "origin", "HEAD:main"],
        cwd=ROOT,
        check=False,
        capture_output=True,
        text=True,
    )
    if push_main.returncode == 0:
        return {"pushed_main": True, "pull_request": None}

    subprocess.run(["git", "checkout", "-B", CHANGELOG_BRANCH], cwd=ROOT, check=True)
    subprocess.run(
        ["git", "push", "origin", f"HEAD:{CHANGELOG_BRANCH}", "--force"],
        cwd=ROOT,
        check=True,
    )
    body = "\n".join(
        [
            f"## Automated CHANGELOG cut for v{version}",
            "",
            "Docs-only housekeeping after a successful Release publish.",
            "Direct push to `main` was blocked by branch protection.",
            "",
            f"- Version: v{version}",
        ]
    )
    existing = subprocess.run(
        [
            "gh",
            "pr",
            "list",
            "--repo",
            _repo(),
            "--head",
            CHANGELOG_BRANCH,
            "--state",
            "open",
            "--json",
            "number,url",
            "--limit",
            "1",
        ],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    open_prs = json.loads(existing.stdout or "[]")
    if open_prs:
        pr_number = int(open_prs[0]["number"])
        subprocess.run(
            [
                "gh",
                "pr",
                "edit",
                str(pr_number),
                "--repo",
                _repo(),
                "--title",
                title,
                "--body",
                body,
            ],
            cwd=ROOT,
            check=True,
        )
        pr_url = str(open_prs[0].get("url") or "")
    else:
        created = subprocess.run(
            [
                "gh",
                "pr",
                "create",
                "--repo",
                _repo(),
                "--base",
                "main",
                "--head",
                CHANGELOG_BRANCH,
                "--title",
                title,
                "--body",
                body,
                "--label",
                "maintenance",
                "--label",
                "skip-changelog",
            ],
            cwd=ROOT,
            check=True,
            capture_output=True,
            text=True,
        )
        pr_url = (created.stdout or "").strip()
        pr_number = int(pr_url.rstrip("/").split("/")[-1]) if pr_url else 0
    if pr_number:
        subprocess.run(
            [
                "gh",
                "pr",
                "merge",
                str(pr_number),
                "--repo",
                _repo(),
                "--squash",
                "--auto",
            ],
            cwd=ROOT,
            check=False,
        )
    return {"pushed_main": False, "pull_request": pr_url or None}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--release-url", default="", help="Unused; kept for workflow compatibility")
    parser.add_argument("--cut-changelog", action="store_true")
    parser.add_argument("--commit-changelog", action="store_true")
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    labeled: list[int] = []
    errors: list[str] = []
    for issue_number in awaiting_release_without_released():
        try:
            label_issue_released(issue_number)
            labeled.append(issue_number)
        except subprocess.CalledProcessError as err:
            errors.append(f"#{issue_number}: {err}")

    changelog_cut = False
    changelog_commit: dict[str, object] = {"pushed_main": False, "pull_request": None}
    if args.cut_changelog:
        changelog_cut = cut_changelog(args.version)
        if args.commit_changelog and changelog_cut:
            changelog_commit = commit_changelog(args.version)

    payload = {
        "labeled_issues": labeled,
        "errors": errors,
        "changelog_cut": changelog_cut,
        "changelog_commit": changelog_commit,
    }
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))

    return 1 if errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
