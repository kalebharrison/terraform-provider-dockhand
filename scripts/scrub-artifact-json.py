#!/usr/bin/env python3
"""Redact sensitive fields from /tmp/dh-*.json before CI artifact upload."""

from __future__ import annotations

import glob
import re
import sys
from pathlib import Path

PATTERNS = (
    r'("(?:password|token|agent_token|agentToken|hawserToken|secret|apiKey|api_token)")\s*:\s*"[^"]*"',
    r'("(?:cert|key|privateKey|certificate)")\s*:\s*"[^"]*"',
    r'("(?:smtp_username|config_json|compose)")\s*:\s*"[^"]*"',
)


def scrub(text: str) -> str:
    scrubbed = text
    for pattern in PATTERNS:
        scrubbed = re.sub(pattern, r'\1:"[redacted]"', scrubbed, flags=re.IGNORECASE)
    return scrubbed


def main() -> int:
    paths = sorted(glob.glob("/tmp/dh-*.json"))
    if not paths:
        return 0
    for path_str in paths:
        path = Path(path_str)
        original = path.read_text(encoding="utf-8", errors="replace")
        cleaned = scrub(original)
        if cleaned != original:
            path.write_text(cleaned, encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
