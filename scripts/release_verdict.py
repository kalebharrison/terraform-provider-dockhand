#!/usr/bin/env python3
"""Parse release lens verdicts from agent-review-log.md."""

from __future__ import annotations

import re

VERDICT_BLOCK_RE = re.compile(
    r"(?ms)^###\s+Release\s+v(?P<version>\d+\.\d+\.\d+)\s+—\s+verdict\s*$"
    r".*?"
    r"^\-\s+\*\*Clear to tag:\*\*\s*(?P<clear>yes|no)\s*$"
)


def clear_to_tag_versions(text: str) -> list[str]:
    """Return semver strings (no v prefix) cleared for tagging."""
    cleared: list[str] = []
    for match in VERDICT_BLOCK_RE.finditer(text or ""):
        if match.group("clear").lower() != "yes":
            continue
        cleared.append(match.group("version"))
    return cleared


def is_cleared_for_version(text: str, version: str) -> bool:
    normalized = version.lstrip("v")
    return normalized in clear_to_tag_versions(text)
