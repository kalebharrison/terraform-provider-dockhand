#!/usr/bin/env python3
"""Ensure Dockhand Compat validated on the current main SHA."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import (
    RELEASE_WATCH_ENSURE_TIMEOUT_SEC,
    ensure_release_watch_green,
)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--retries", type=int, default=1)
    parser.add_argument("--timeout-sec", type=int, default=RELEASE_WATCH_ENSURE_TIMEOUT_SEC)
    args = parser.parse_args(argv)

    try:
        result = ensure_release_watch_green(
            retries=max(0, args.retries),
            timeout_sec=args.timeout_sec,
        )
    except (RuntimeError, TimeoutError) as err:
        print(str(err), file=sys.stderr)
        return 1

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
