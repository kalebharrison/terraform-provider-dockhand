#!/usr/bin/env python3
"""CI gate: agent branches must update agent-review-log.md when lenses apply."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from issue_agent_intake_lenses import REVIEW_LOG, branch_issue_number, select_lenses

REVIEW_LOG_PATH = REVIEW_LOG


def changed_files(base_ref: str, head_ref: str) -> set[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", f"{base_ref}...{head_ref}"],
        cwd=ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    return {line.strip() for line in result.stdout.splitlines() if line.strip()}


def gate_ok(*, labels: list[str], title: str, base_ref: str, head_ref: str) -> tuple[bool, str]:
    lenses = select_lenses(labels, title)
    if not lenses:
        return True, "no lenses required"

    changed = changed_files(base_ref, head_ref)
    if REVIEW_LOG_PATH in changed:
        return True, f"lens log updated ({', '.join(lenses)})"

    return (
        False,
        "missing required lens sweep log: "
        f"update {REVIEW_LOG_PATH} with sections for: {', '.join(lenses)}",
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--branch", required=True, help="Agent branch name")
    parser.add_argument("--labels", default="", help="Comma-separated issue labels")
    parser.add_argument("--title", default="")
    parser.add_argument("--base-ref", default="origin/main")
    parser.add_argument("--head-ref", default="HEAD")
    args = parser.parse_args(argv)

    issue_number = branch_issue_number(args.branch)
    if issue_number is None:
        print("skip: not an agent/issue-* branch")
        return 0

    labels = [part.strip() for part in args.labels.split(",") if part.strip()]
    ok, message = gate_ok(
        labels=labels,
        title=args.title,
        base_ref=args.base_ref,
        head_ref=args.head_ref,
    )
    if ok:
        print(message)
        return 0

    print(message, file=sys.stderr)
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
