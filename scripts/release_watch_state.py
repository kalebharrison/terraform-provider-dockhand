#!/usr/bin/env python3
"""Helpers for Dockhand Release Watch skip/cache state."""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
from datetime import datetime, timezone
from pathlib import Path

STATE_DIR = Path(".ci-state/dockhand-release-watch")
STATE_FILE = STATE_DIR / "state.env"


def _parse_env_file(text: str) -> dict[str, str]:
    out: dict[str, str] = {}
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        out[key.strip()] = value.strip()
    return out


def load_state(path: Path = STATE_FILE) -> dict[str, str]:
    if not path.is_file():
        return {}
    return _parse_env_file(path.read_text(encoding="utf-8"))


def write_state(
    *,
    last_tag: str,
    last_digest: str,
    updated_at: str | None = None,
    source_run: str = "",
    last_provider_sha: str = "",
    path: Path = STATE_FILE,
) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    stamp = updated_at or datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    lines = [
        f"last_tag={last_tag}",
        f"last_digest={last_digest}",
        f"updated_at={stamp}",
        f"source_run={source_run}",
        f"last_provider_sha={last_provider_sha}",
    ]
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def parse_legacy_issue_body(body: str) -> dict[str, str]:
    out: dict[str, str] = {}
    for key in ("last_tag", "last_digest", "updated_at", "source_run", "last_provider_sha"):
        match = re.search(rf"^{re.escape(key)}=(.+)$", body, flags=re.MULTILINE)
        if match:
            out[key] = match.group(1).strip()
    return out


def decide_should_run(
    *,
    tag: str,
    digest: str,
    state: dict[str, str],
    event_name: str,
    force_validate: bool,
    manual_image_tag: bool,
    main_sha: str,
) -> tuple[bool, str]:
    last_tag = state.get("last_tag", "")
    last_digest = state.get("last_digest", "")
    last_provider_sha = state.get("last_provider_sha", "")

    if force_validate:
        return True, "gate_required"
    if manual_image_tag:
        return True, "manual_override"
    if not last_tag:
        return True, "state_unset"
    if tag != last_tag:
        return True, "new_tag"
    if digest and last_digest and digest != last_digest:
        return True, "changed_digest"
    if digest and not last_digest:
        return True, "state_missing_digest"
    if main_sha and last_provider_sha and main_sha != last_provider_sha:
        return True, "provider_main_changed"

    unchanged = (digest and last_digest and digest == last_digest) or (
        not digest or not last_digest
    )
    if not unchanged:
        return True, "new_tag_or_digest"

    return False, "unchanged_tag_digest"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="cmd", required=True)

    decide = sub.add_parser("decide", help="Return should_run decision as JSON")
    decide.add_argument("--json", action="store_true")
    decide.add_argument("--tag", required=True)
    decide.add_argument("--digest", default="")
    decide.add_argument("--event-name", default="workflow_dispatch")
    decide.add_argument("--force-validate", action="store_true")
    decide.add_argument("--manual-image-tag", action="store_true")
    decide.add_argument("--main-sha", default="")
    decide.add_argument("--state-file", type=Path, default=STATE_FILE)

    write = sub.add_parser("write", help="Write state.env")
    write.add_argument("--tag", required=True)
    write.add_argument("--digest", default="")
    write.add_argument("--updated-at", default="")
    write.add_argument("--source-run", default="")
    write.add_argument("--last-provider-sha", default="")
    write.add_argument("--state-file", type=Path, default=STATE_FILE)

    args = parser.parse_args(argv)

    if args.cmd == "decide":
        state = load_state(args.state_file)
        should_run, reason = decide_should_run(
            tag=args.tag,
            digest=args.digest,
            state=state,
            event_name=args.event_name,
            force_validate=args.force_validate,
            manual_image_tag=args.manual_image_tag,
            main_sha=args.main_sha,
        )
        payload = {
            "should_run": should_run,
            "reason": reason,
            "last_tag": state.get("last_tag", ""),
            "last_digest": state.get("last_digest", ""),
            "updated_at": state.get("updated_at", ""),
            "last_provider_sha": state.get("last_provider_sha", ""),
        }
        if args.json:
            print(json.dumps(payload))
        else:
            print(json.dumps(payload))
        return 0

    if args.cmd == "write":
        write_state(
            last_tag=args.tag,
            last_digest=args.digest,
            updated_at=args.updated_at or None,
            source_run=args.source_run,
            last_provider_sha=args.last_provider_sha,
            path=args.state_file,
        )
        return 0

    return 1


if __name__ == "__main__":
    raise SystemExit(main())
