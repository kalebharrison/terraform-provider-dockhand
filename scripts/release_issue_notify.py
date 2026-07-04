#!/usr/bin/env python3
"""Post release comments on issues linked from GitHub release notes."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import _gh_json, _repo
from release_housekeeping import awaiting_release_without_released

NOTIFY_MARKER = "Posted by Release Issue Notify."
CLOSE_REGEX = re.compile(r"\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+#(\d+)\b", re.IGNORECASE)
PR_REF_REGEX = re.compile(r"#(\d+)")


def _run(args: list[str], *, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        args,
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=check,
    )


def get_release_by_tag(tag: str) -> dict:
    proc = _run(
        ["gh", "release", "view", tag, "--repo", _repo(), "--json", "tagName,body,url"],
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or f"release {tag} not found")
    return json.loads(proc.stdout)


def latest_published_release() -> dict | None:
    data = _gh_json(["release", "list", "--repo", _repo(), "--limit", "10", "--json", "tagName,isDraft,isPrerelease,body,url"])
    if not isinstance(data, list):
        return None
    for item in data:
        if not isinstance(item, dict):
            continue
        if item.get("isDraft") or item.get("isPrerelease"):
            continue
        tag = str(item.get("tagName") or "").strip()
        if tag.startswith("v"):
            return item
    return None


def section_from_body(heading: str, body: str) -> str:
    pattern = re.compile(
        rf"## {re.escape(heading)}\s*\n+([\s\S]*?)(?=\n## |$)",
        re.IGNORECASE,
    )
    match = pattern.search(body or "")
    return match.group(1).strip() if match else ""


def issue_numbers_from_release(release_body: str) -> tuple[set[int], set[int]]:
    pr_numbers: set[int] = set()
    for match in PR_REF_REGEX.finditer(release_body or ""):
        pr_numbers.add(int(match.group(1)))

    issue_numbers: set[int] = set()
    for pr_number in pr_numbers:
        proc = _run(
            [
                "gh",
                "api",
                f"repos/{_repo()}/pulls/{pr_number}",
                "--jq",
                ".body",
            ],
            check=False,
        )
        if proc.returncode != 0:
            continue
        pr_body = proc.stdout
        for match in CLOSE_REGEX.finditer(pr_body):
            issue_numbers.add(int(match.group(1)))
    return issue_numbers, pr_numbers


def already_notified(issue_number: int) -> bool:
    proc = _run(
        [
            "gh",
            "api",
            f"repos/{_repo()}/issues/{issue_number}/comments",
            "--paginate",
            "--jq",
            ".[].body",
        ],
        check=False,
    )
    if proc.returncode != 0:
        return False
    return NOTIFY_MARKER in proc.stdout


def upgrade_block(tag: str) -> str:
    version = tag.lstrip("v")
    return "\n".join(
        [
            "### Upgrade",
            "",
            "In your Terraform configuration:",
            "",
            "```hcl",
            "terraform {",
            "  required_providers {",
            "    dockhand = {",
            '      source  = "kalebharrison/dockhand"',
            f'      version = ">= {version}"',
            "    }",
            "  }",
            "}",
            "```",
            "",
            "Then run `terraform init -upgrade`.",
        ]
    )


def reopen_block(tag: str) -> str:
    return "\n".join(
        [
            "### If the problem continues",
            "",
            f"Please **reopen this issue** or **comment here** if you still see the problem **after upgrading to `{tag}`**.",
            "",
            "Include provider version, Dockhand version, and a minimal repro or error message.",
        ]
    )


def pr_context(pr_numbers: set[int]) -> tuple[str, str]:
    what_fixed = ""
    user_impact = ""
    for pr_number in pr_numbers:
        proc = _run(
            ["gh", "api", f"repos/{_repo()}/pulls/{pr_number}", "--jq", ".body"],
            check=False,
        )
        if proc.returncode != 0:
            continue
        pr_body = proc.stdout
        if not what_fixed:
            what_fixed = section_from_body("What was fixed", pr_body)
        if not user_impact:
            user_impact = section_from_body("User impact", pr_body)
    return what_fixed, user_impact


def build_comment_body(
    *,
    tag: str,
    release_url: str,
    what_fixed: str = "",
    user_impact: str = "",
    generic: bool = False,
) -> str:
    parts = [
        f"## Released in `{tag}`",
        "",
        f"Release notes: {release_url}" if release_url else "",
        "",
    ]
    if generic:
        parts.extend(
            [
                "This fix is included in the published provider release above.",
                "",
            ]
        )
    if what_fixed:
        parts.extend(["### What was fixed", "", what_fixed, ""])
    if user_impact:
        parts.extend(["### User impact", "", user_impact, ""])
    parts.extend([upgrade_block(tag), "", reopen_block(tag), "", "---", f"_{NOTIFY_MARKER}_"])
    return "\n".join(line for line in parts if line != "")


def post_comment(issue_number: int, body: str) -> None:
    _run(
        [
            "gh",
            "issue",
            "comment",
            str(issue_number),
            "--repo",
            _repo(),
            "--body",
            body,
        ]
    )


def label_released(issue_number: int) -> None:
    _run(
        [
            "gh",
            "issue",
            "edit",
            str(issue_number),
            "--repo",
            _repo(),
            "--add-label",
            "released",
        ]
    )
    for label in ("in-progress", "awaiting-release"):
        _run(
            [
                "gh",
                "issue",
                "edit",
                str(issue_number),
                "--repo",
                _repo(),
                "--remove-label",
                label,
            ],
            check=False,
        )


def notify_issues_for_release(
    tag: str,
    *,
    issue_numbers: set[int] | None = None,
    generic_for_unlinked: bool = False,
) -> dict:
    release = get_release_by_tag(tag)
    release_url = str(release.get("url") or "")
    release_body = str(release.get("body") or "")
    linked_issues, pr_numbers = issue_numbers_from_release(release_body)
    what_fixed, user_impact = pr_context(pr_numbers)

    targets = set(issue_numbers or linked_issues)
    if generic_for_unlinked:
        targets.update(awaiting_release_without_released())

    notified: list[int] = []
    skipped: list[int] = []
    errors: list[str] = []

    for issue_number in sorted(targets):
        if already_notified(issue_number):
            skipped.append(issue_number)
            continue
        generic = issue_number not in linked_issues
        try:
            body = build_comment_body(
                tag=tag,
                release_url=release_url,
                what_fixed="" if generic else what_fixed,
                user_impact="" if generic else user_impact,
                generic=generic,
            )
            post_comment(issue_number, body)
            label_released(issue_number)
            notified.append(issue_number)
        except subprocess.CalledProcessError as err:
            errors.append(f"#{issue_number}: {err}")

    return {
        "tag": tag,
        "notified": notified,
        "skipped": skipped,
        "errors": errors,
        "linked_issues": sorted(linked_issues),
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tag", help="Release tag (for example v0.1.48)")
    parser.add_argument(
        "--backfill",
        action="store_true",
        help="Notify all awaiting-release issues using the latest published release",
    )
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    tag = (args.tag or "").strip()
    if args.backfill:
        if tag:
            parser.error("use either --tag or --backfill, not both")
        latest = latest_published_release()
        if latest is None:
            print("No published release found.", file=sys.stderr)
            return 1
        tag = str(latest.get("tagName") or "").strip()
        if not tag:
            print("Latest release has no tag.", file=sys.stderr)
            return 1
        result = notify_issues_for_release(tag, generic_for_unlinked=True)
    else:
        if not tag:
            parser.error("--tag is required unless --backfill is set")
        result = notify_issues_for_release(tag)

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 1 if result["errors"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
