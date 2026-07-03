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

DEFAULT_REPO_URL = "https://github.com/kalebharrison/terraform-provider-dockhand"
CURSOR_API = "https://api.cursor.com/v1/agents"
BRANCH_RE = re.compile(r"^agent/issue-(\d+)-([a-z0-9-]+)$")


def slugify(title: str, max_len: int = 48) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", title.strip().lower()).strip("-")
    if not slug:
        return "task"
    return slug[:max_len].strip("-")


def branch_name(issue_number: int, title: str) -> str:
    return f"agent/issue-{issue_number}-{slugify(title)}"


def build_prompt(issue_number: int, title: str, body: str, branch: str) -> str:
    issue_body = (body or "").strip()
    if not issue_body:
        issue_body = "(No description provided — infer scope from the title and codebase.)"

    return f"""You are implementing a fix for GitHub issue #{issue_number} in terraform-provider-dockhand.

## Read first (in repo)
- docs/AGENT_RUNBOOK.md
- docs/AGENT_CODING_STANDARDS.md
- docs/AGENT_ISSUE_RESPONSE.md
- AGENTS.md

## Branch contract
Work on branch `{branch}` (already created from main). Do **not** push to `main` or use Cursor auto `cursor/*` branches.
Every commit on `agent/**` must include:
  Co-authored-by: Cursor Agent <noreply@cursor.com>
Use `./scripts/agent-commit-msg.sh` to format commit messages.

## Issue #{issue_number}: {title.strip()}

{issue_body}

## Done when
1. Fix is implemented with focused diffs and tests/docs as required by AGENT_CODING_STANDARDS.md
2. `./scripts/verify.sh --quality` passes locally before push
3. Branch `{branch}` is pushed to origin — GitHub **Agent Validate** and **Agent Open PR** run automatically
4. When the PR opens, ensure **What was fixed** and **User impact** in the PR body are filled (not placeholders)

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
    parser.add_argument("--title", required=True)
    parser.add_argument("--body", default="")
    parser.add_argument("--body-file", help="Read issue body from a file")
    parser.add_argument("--branch", help="Override branch name")
    parser.add_argument("--repo-url", default=DEFAULT_REPO_URL)
    parser.add_argument("--model", default="composer-2.5")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args(argv)

    body = args.body
    if args.body_file:
        with open(args.body_file, encoding="utf-8") as handle:
            body = handle.read()

    branch = args.branch or branch_name(args.issue_number, args.title)
    if not BRANCH_RE.match(branch):
        print(f"invalid agent branch: {branch}", file=sys.stderr)
        return 2

    prompt = build_prompt(args.issue_number, args.title, body, branch)
    if args.dry_run:
        print(json.dumps({"branch": branch, "prompt_chars": len(prompt)}, indent=2))
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
