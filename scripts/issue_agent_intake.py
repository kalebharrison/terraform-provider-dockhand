#!/usr/bin/env python3
"""Build prompts and dispatch Cursor Cloud Agents for agent issue intake."""

from __future__ import annotations

import argparse
import base64
import json
import os
import re
import sys
import urllib.error
import urllib.request
from typing import Any

from issue_agent_intake_lenses import build_lens_instructions, select_lenses

DEFAULT_REPO_URL = "https://github.com/kalebharrison/terraform-provider-dockhand"
CURSOR_API = "https://api.cursor.com/v1/agents"
BRANCH_RE = re.compile(r"^agent/issue-(\d+)-([a-z0-9-]+)$")


def slugify(title: str, max_len: int = 48) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", title.strip().lower()).strip("-")
    if not slug:
        return "task"
    return slug[:max_len].strip("-")


def parse_labels(raw: str) -> list[str]:
    return [part.strip() for part in (raw or "").split(",") if part.strip()]


def branch_name(issue_number: int, title: str) -> str:
    return f"agent/issue-{issue_number}-{slugify(title)}"


def parse_release_version(title: str) -> str:
    match = re.search(r"release:\s*prepare\s+v(\d+\.\d+\.\d+)", (title or ""), re.IGNORECASE)
    return match.group(1) if match else ""


def build_prompt(
    issue_number: int,
    title: str,
    body: str,
    branch: str,
    labels: list[str] | None = None,
) -> str:
    issue_body = (body or "").strip()
    if not issue_body:
        issue_body = "(No description provided — infer scope from the title and codebase.)"

    label_list = labels or []
    labels_l = {label.lower() for label in label_list}
    is_release = "release-candidate" in labels_l or (title or "").lower().startswith("release:")
    release_version = parse_release_version(title)
    selected = select_lenses(label_list, title, issue_body)
    lens_block = build_lens_instructions(
        selected,
        issue_number=issue_number,
        title=title,
        is_release=is_release,
        release_version=release_version,
    )

    if is_release:
        done_when = f"""## Done when
1. Complete the release-tier lens set in `docs/reports/agent-review-log.md`
2. Append `### Release v{release_version or "X.Y.Z"} — verdict` with **Clear to tag: yes** when all required lenses pass with no unresolved **high** findings
3. Push branch `{branch}` — **Agent Validate** and **Agent Open PR** run automatically
4. **Agent Release Tag** publishes the GitHub release after the verdict merges to `main` — do not tag manually"""
    else:
        done_when = f"""## Done when
1. Required lens sweep sections are appended to `docs/reports/agent-review-log.md`
2. Fix is implemented with focused diffs and tests/docs as required by AGENT_CODING_STANDARDS.md
3. `./scripts/verify.sh --quality` passes locally before push
4. Branch `{branch}` is pushed to origin — GitHub **Agent Validate** and **Agent Open PR** run automatically
5. When the PR opens, ensure **What was fixed** and **User impact** in the PR body are filled (not placeholders)"""

    return f"""You are implementing a fix for GitHub issue #{issue_number} in terraform-provider-dockhand.

## Read first (in repo)
- docs/AGENT_RUNBOOK.md
- docs/AGENT_CODING_STANDARDS.md
- docs/AGENT_ISSUE_RESPONSE.md
- docs/AGENT_REVIEW_LENSES.md
- AGENTS.md

{lens_block}## Branch contract
Work on branch `{branch}` (already created from main). Do **not** push to `main` or use Cursor auto `cursor/*` branches.
Every commit on `agent/**` must include:
  Co-authored-by: Cursor Agent <noreply@cursor.com>
Use `./scripts/agent-commit-msg.sh` to format commit messages.

## Issue #{issue_number}: {title.strip()}

{issue_body}

{done_when}

Do not merge PRs, tag releases, or change branch protection. Push and let CI complete the loop.
"""


def dispatch_cursor_agent(
    api_key: str,
    *,
    branch: str,
    prompt: str,
    repo_url: str = DEFAULT_REPO_URL,
    model_id: str = "composer-2.5",
) -> dict[str, Any]:
    payload = {
        "prompt": {"text": prompt},
        "model": {"id": model_id},
        "repos": [{"url": repo_url, "startingRef": branch}],
        "workOnCurrentBranch": True,
        "skipReviewerRequest": True,
    }
    data = json.dumps(payload).encode("utf-8")
    request = urllib.request.Request(
        CURSOR_API,
        data=data,
        method="POST",
        headers={
            "Content-Type": "application/json",
            "Authorization": "Basic "
            + base64.b64encode(f"{api_key}:".encode()).decode(),
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as err:
        detail = err.read().decode(errors="replace")
        raise RuntimeError(f"Cursor API HTTP {err.code}: {detail}") from err


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--issue-number", type=int, required=True)
    parser.add_argument("--title", default="")
    parser.add_argument("--title-file", help="Read issue title from a file")
    parser.add_argument("--body", default="")
    parser.add_argument("--body-file", help="Read issue body from a file")
    parser.add_argument("--labels", default="", help="Comma-separated issue labels")
    parser.add_argument("--labels-file", help="Read comma-separated labels from a file")
    parser.add_argument("--branch", help="Override branch name")
    parser.add_argument("--repo-url", default=DEFAULT_REPO_URL)
    parser.add_argument("--model", default="composer-2.5")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args(argv)

    title = args.title
    if args.title_file:
        with open(args.title_file, encoding="utf-8") as handle:
            title = handle.read().rstrip("\n")
    if not title.strip():
        print("--title or --title-file is required", file=sys.stderr)
        return 2

    body = args.body
    if args.body_file:
        with open(args.body_file, encoding="utf-8") as handle:
            body = handle.read()

    labels_raw = args.labels
    if args.labels_file:
        with open(args.labels_file, encoding="utf-8") as handle:
            labels_raw = handle.read().strip()

    branch = args.branch or branch_name(args.issue_number, title)
    if not BRANCH_RE.match(branch):
        print(f"invalid agent branch: {branch}", file=sys.stderr)
        return 2

    labels = parse_labels(labels_raw)
    prompt = build_prompt(args.issue_number, title, body, branch, labels)
    if args.dry_run:
        print(
            json.dumps(
                {
                    "branch": branch,
                    "prompt_chars": len(prompt),
                    "lenses": select_lenses(labels, title, body),
                },
                indent=2,
            )
        )
        return 0

    api_key = (os.environ.get("CURSOR_API_KEY") or "").strip()
    if not api_key:
        print("CURSOR_API_KEY is not set", file=sys.stderr)
        return 2

    result = dispatch_cursor_agent(
        api_key,
        branch=branch,
        prompt=prompt,
        repo_url=args.repo_url,
        model_id=args.model,
    )
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
