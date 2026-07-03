#!/usr/bin/env python3
"""Select review lenses and build automated sweep instructions for agent intake."""

from __future__ import annotations

import re
from typing import Iterable

LENS_API = "API compatibility"
LENS_OPS = "Ops / SRE"
LENS_SECURITY = "Security engineer"
LENS_SCHEMA = "Terraform schema & state"
LENS_ACCEPTANCE = "Acceptance & regression"
LENS_SENIOR = "Senior developer"
LENS_DOCKHAND = "Dockhand domain / runtime"
LENS_ASYNC = "Async & long-running operations"
LENS_GITOPS = "GitOps / IaC practitioner"
LENS_RELEASE = "Release & upgrade"

RELEASE_CORE_5 = [
    LENS_API,
    LENS_SCHEMA,
    LENS_ACCEPTANCE,
    LENS_SECURITY,
    LENS_RELEASE,
]

RELEASE_ALL_11 = [
    LENS_API,
    LENS_SCHEMA,
    LENS_DOCKHAND,
    "Async & long-running operations",
    LENS_ACCEPTANCE,
    LENS_SECURITY,
    LENS_OPS,
    LENS_SENIOR,
    LENS_GITOPS,
    "Entry-level developer",
    LENS_RELEASE,
]

REVIEW_LOG = "docs/reports/agent-review-log.md"
LENSES_DOC = "docs/AGENT_REVIEW_LENSES.md"
BRANCH_ISSUE_RE = re.compile(r"^agent/issue-(\d+)-")


def branch_issue_number(branch: str) -> int | None:
    match = BRANCH_ISSUE_RE.match((branch or "").strip())
    if not match:
        return None
    return int(match.group(1))


def _dedupe(items: Iterable[str]) -> list[str]:
    seen: set[str] = set()
    ordered: list[str] = []
    for item in items:
        if item in seen:
            continue
        seen.add(item)
        ordered.append(item)
    return ordered


def release_lenses_for_tier(tier: str) -> list[str]:
    if tier in {"minor", "major"}:
        return list(RELEASE_ALL_11)
    return list(RELEASE_CORE_5)


def parse_release_tier(title: str, body: str) -> str:
    body_l = (body or "").lower()
    for line in body_l.splitlines():
        if line.strip().startswith("tier:"):
            value = line.split(":", 1)[1].strip()
            if value in {"patch", "minor", "major"}:
                return value
    return "patch"


def select_lenses(labels: list[str], title: str, body: str = "") -> list[str]:
    """Return ordered lens names required for this issue."""
    labels_l = {label.strip().lower() for label in labels if label.strip()}
    title_l = (title or "").strip().lower()

    if "release-candidate" in labels_l or title_l.startswith("release:"):
        return release_lenses_for_tier(parse_release_tier(title, body))

    lenses: list[str] = []

    compatibility = bool(
        labels_l & {"compatibility", "api-drift"}
        or "compatibility failure" in title_l
        or "api drift detected" in title_l
    )
    if compatibility:
        lenses.extend([LENS_API, LENS_OPS])

    if "regression" in labels_l:
        lenses.append(LENS_ACCEPTANCE)

    if "security" in labels_l or "security" in title_l:
        lenses.append(LENS_SECURITY)

    if "ci" in labels_l and "no-agent" not in labels_l:
        lenses.append(LENS_OPS)

    if "enhancement" in labels_l or title_l.startswith("[feature]"):
        lenses.extend([LENS_SCHEMA, LENS_ACCEPTANCE, LENS_GITOPS])

    if "bug" in labels_l or title_l.startswith("[bug]"):
        lenses.extend([LENS_ACCEPTANCE, LENS_SENIOR])

    if any(token in title_l for token in ("environment", "hawser", "stack", "container", "git stack")):
        lenses.append(LENS_DOCKHAND)

    if any(token in title_l for token in ("action", "job", "stream", "deploy", "batch")):
        lenses.append(LENS_ASYNC)

    if not lenses:
        lenses = [LENS_SENIOR, LENS_ACCEPTANCE]

    return _dedupe(lenses)


def build_lens_instructions(
    lenses: list[str],
    *,
    issue_number: int,
    title: str,
    is_release: bool = False,
    release_version: str = "",
) -> str:
    if not lenses:
        return ""

    lens_lines = "\n".join(f"- **{name}**" for name in lenses)
    release_close = ""
    if is_release and release_version:
        release_close = f"""
6. Close the release block with:

```markdown
### Release v{release_version} — verdict

- **Clear to tag:** yes | no
- **Blocking findings:** <none or list>
- **Deferred medium/low:** <issue links>
```

Set **Clear to tag: yes** only when all required lenses pass with no unresolved **high** findings.
"""

    return f"""## Required automated lens sweep (before finishing)

Every agent issue runs focused review lenses automatically. For issue #{issue_number} run:

{lens_lines}

### How to run (mandatory)
1. Open `{LENSES_DOC}` and complete the **Goal**, **Priority paths**, and **Checklist** for each lens above.
2. Follow tier order in `docs/testing/release-lens-review.md` when this is a release candidate issue.
3. Append one `### YYYY-MM-DD — <lens name>` section per lens to `{REVIEW_LOG}` using the finding table format in `{LENSES_DOC}`.
4. Start the log block with:
   `## {"Release v" + release_version if is_release else f"Issue #{issue_number} — {title.strip()}"}`
5. Fix **high** severity findings in this branch before push. File GitHub issues for deferred medium/low items and link them in the log.
6. Commit the review log update on this branch **before or with** your fix commits — **Agent Validate** fails if `{REVIEW_LOG}` is not updated on this branch.
{release_close}
Then implement the issue fix below.
"""


def lenses_required(labels: list[str], title: str, body: str = "") -> bool:
    return bool(select_lenses(labels, title, body))
