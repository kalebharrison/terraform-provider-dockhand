#!/usr/bin/env python3
"""Semver helpers for automated release orchestration."""

from __future__ import annotations

import re
from dataclasses import dataclass

SEMVER_RE = re.compile(r"^v?(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$")


@dataclass(frozen=True)
class Version:
    major: int
    minor: int
    patch: int

    @classmethod
    def parse(cls, raw: str) -> Version | None:
        match = SEMVER_RE.match((raw or "").strip())
        if not match:
            return None
        return cls(
            major=int(match.group("major")),
            minor=int(match.group("minor")),
            patch=int(match.group("patch")),
        )

    def tag(self) -> str:
        return f"v{self.major}.{self.minor}.{self.patch}"

    def __str__(self) -> str:
        return self.tag().lstrip("v")


def release_tier(previous: Version | None, target: Version) -> str:
    """Return patch, minor, or major tier for lens selection."""
    if previous is None:
        return "minor"
    if target.major > previous.major:
        return "major"
    if target.minor > previous.minor:
        return "minor"
    return "patch"
