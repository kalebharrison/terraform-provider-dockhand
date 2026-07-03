#!/usr/bin/env python3
"""Eligibility rules for automated Issue Agent Intake dispatch."""

from __future__ import annotations

import argparse
import json
import re
import sys

SKIP_LABELS = frozenset({"released", "awaiting-release", "no-agent"})
BOT_AUTO_LABELS = frozenset({"compatibility", "api-drift"})
RETRIGGER_LABELS = frozenset({"agent", "compatibility", "api-drift", "regression", "release-candidate"})
AUTOMATION_TRACKER_PREFIXES = ("[Automation] Workflow failing:", "[Automation] Release gate blocked")

ACCEPTANCE_SECTION_RE = re.compile(
    r"(?m)^#{2,3}\s*("
    r"done when|acceptance criteria|expected behavior|actual behavior|"
    r"problem|problem statement|proposed solution|what was fixed|scope|context|endpoints"
    r")\b",
    re.IGNORECASE,
)


def has_acceptance_section(body: str) -> bool:
    return bool(ACCEPTANCE_SECTION_RE.search(body or ""))


def is_bot(author_type: str, author_login: str) -> bool:
    login = (author_login or "").lower()
    return (author_type or "").lower() == "bot" or login.endswith("[bot]")


def intake_eligible(
    *,
    title: str,
    labels: list[str],
    author_type: str,
    author_login: str,
    trigger: str,
    body: str,
    has_dispatched: bool,
    has_regression: bool,
    labeled_name: str = "",
    comment_body: str = "",
) -> tuple[bool, str | None]:
    labels_l = {label.strip().lower() for label in labels if label.strip()}
    title_s = (title or "").strip()

    if labels_l & SKIP_LABELS:
        blocked = ", ".join(sorted(labels_l & SKIP_LABELS))
        return False, f"issue has skip label(s): {blocked}"

    if any(title_s.startswith(prefix) for prefix in AUTOMATION_TRACKER_PREFIXES):
        return False, "automation tracker (not agent work)"

    if "release-candidate" in labels_l or title_s.lower().startswith("release:"):
        if has_dispatched and not has_regression and trigger != "manual":
            return False, "already has agent-dispatched label"
        body_trim = (body or "").strip()
        if len(body_trim) < 80 and not has_acceptance_section(body_trim):
            return False, "release issue body too vague for dispatch"
        return True, None

    if has_dispatched and not has_regression and trigger != "manual":
        return False, "already has agent-dispatched label"

    if trigger == "labeled":
        label = (labeled_name or "").strip().lower()
        if label not in RETRIGGER_LABELS and label != "release-candidate":
            return False, f"ignored label event: {label or 'unknown'}"

    if trigger == "comment":
        text = (comment_body or "").strip()
        if is_bot(author_type, author_login):
            return False, "bot comment"
        if not has_regression and not re.search(r"(^|\s)/agent(\s|$)", text, re.IGNORECASE):
            return False, "comment did not request agent (/agent) and issue is not regression"

    if trigger in {"opened", "labeled", "comment", "manual"}:
        if is_bot(author_type, author_login):
            if not (labels_l & BOT_AUTO_LABELS):
                return False, "bot-opened issue without compatibility/api-drift labels"
        body_trim = (body or "").strip()
        if is_bot(author_type, author_login):
            if len(body_trim) < 80 and not has_acceptance_section(body_trim):
                return False, "bot issue body too vague for dispatch"
        elif len(body_trim) < 20 and not has_acceptance_section(body_trim):
            return False, "issue body too short for dispatch"
        return True, None

    return False, f"unsupported trigger: {trigger}"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--title", required=True)
    parser.add_argument("--labels", default="", help="Comma-separated label names")
    parser.add_argument("--author-type", default="")
    parser.add_argument("--author-login", default="")
    parser.add_argument("--trigger", required=True)
    parser.add_argument("--body-file", help="Issue body file")
    parser.add_argument("--body", default="")
    parser.add_argument("--has-dispatched", action="store_true")
    parser.add_argument("--has-regression", action="store_true")
    parser.add_argument("--labeled-name", default="")
    parser.add_argument("--comment-body", default="")
    parser.add_argument("--json", action="store_true", help="Print JSON result")
    args = parser.parse_args(argv)

    body = args.body
    if args.body_file:
        with open(args.body_file, encoding="utf-8") as handle:
            body = handle.read()

    labels = [part.strip() for part in args.labels.split(",") if part.strip()]
    eligible, reason = intake_eligible(
        title=args.title,
        labels=labels,
        author_type=args.author_type,
        author_login=args.author_login,
        trigger=args.trigger,
        body=body,
        has_dispatched=args.has_dispatched,
        has_regression=args.has_regression,
        labeled_name=args.labeled_name,
        comment_body=args.comment_body,
    )

    result = {"eligible": eligible, "reason": reason}
    if args.json:
        print(json.dumps(result))
        return 0

    print("eligible" if eligible else "skip")
    if reason:
        print(reason, file=sys.stderr)
    return 0 if eligible else 1


if __name__ == "__main__":
    raise SystemExit(main())
