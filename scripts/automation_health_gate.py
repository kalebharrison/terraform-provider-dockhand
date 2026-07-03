#!/usr/bin/env python3
"""Track release-gate blockers and 24h persistence for automation health alerts."""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

from release_gate_check import evaluate_gate

ISSUE_TITLE = "[Automation] Release gate blocked"
ALERT_HOURS = 24


def blockers_fingerprint(blockers: list[str]) -> str:
    normalized = "\n".join(sorted(blocker.strip() for blocker in blockers if blocker.strip()))
    return hashlib.sha256(normalized.encode("utf-8")).hexdigest()[:16]


def effective_blockers(gate: dict) -> list[str]:
    blockers = list(gate.get("blockers") or [])
    if gate.get("ready_for_lens_dispatch") and gate.get("open_release_issue") is None:
        blockers.append("release orchestrator has not opened release lens issue")
    return blockers


def should_track(gate: dict) -> bool:
    return bool(effective_blockers(gate))


def parse_state(state_path: Path) -> tuple[datetime | None, str]:
    if not state_path.is_file():
        return None, ""
    first_seen: datetime | None = None
    fingerprint = ""
    for line in state_path.read_text(encoding="utf-8").splitlines():
        if line.startswith("FIRST_SEEN="):
            raw = line.split("=", 1)[1].strip()
            if raw:
                first_seen = datetime.fromisoformat(raw.replace("Z", "+00:00"))
        elif line.startswith("BLOCKERS_FP="):
            fingerprint = line.split("=", 1)[1].strip()
    return first_seen, fingerprint


def write_state(state_path: Path, first_seen: datetime, fingerprint: str) -> None:
    state_path.parent.mkdir(parents=True, exist_ok=True)
    iso = first_seen.astimezone(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    state_path.write_text(
        f"FIRST_SEEN={iso}\nBLOCKERS_FP={fingerprint}\n",
        encoding="utf-8",
    )


def clear_state(state_path: Path) -> None:
    if state_path.is_file():
        state_path.unlink()


def hours_blocked(first_seen: datetime, now: datetime) -> float:
    delta = now - first_seen.astimezone(timezone.utc)
    return max(delta.total_seconds() / 3600.0, 0.0)


def format_issue_body(gate: dict, blockers: list[str], first_seen: datetime, hours: float) -> str:
    awaiting = gate.get("awaiting_release_issues") or []
    lines = [
        "The automated release pipeline has been blocked for at least 24 hours.",
        "",
        "This issue is managed automatically. It updates while blockers persist and closes when the gate clears.",
        "",
        "## Gate status",
        "",
        f"- CI gates pass: `{gate.get('ci_gates_pass')}`",
        f"- Ready for release lens dispatch: `{gate.get('ready_for_lens_dispatch')}`",
        f"- Ready to tag: `{gate.get('ready_to_tag')}`",
        f"- Draft version: `{gate.get('version') or 'unknown'}`",
        f"- Draft tag: `{gate.get('tag') or 'unknown'}`",
        f"- Lens verdict clear: `{gate.get('lens_verdict_clear')}`",
        f"- Open release issue: `{gate.get('open_release_issue') or 'none'}`",
        f"- Awaiting-release issues: `{', '.join(str(n) for n in awaiting) or 'none'}`",
        f"- Blockers first seen (UTC): `{first_seen.astimezone(timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z')}`",
        f"- Hours blocked: `{hours:.1f}`",
        "",
        "## Blockers",
        "",
    ]
    if blockers:
        lines.extend(f"- {blocker}" for blocker in blockers)
    else:
        lines.append("- none listed")
    lines.extend(
        [
            "",
            "## What happens next",
            "",
            "No maintainer action is required for the normal path. Automation should:",
            "",
            "1. Resolve open compatibility or CI failures via **Issue Agent Intake**",
            "2. Merge agent fixes to `main`",
            "3. Run **Agent Release Orchestrate** → release lens issue → **Agent Release Tag**",
            "",
            "Use `scripts/release_gate_check.py --json` locally only when debugging CI.",
            "",
            "## Raw gate JSON",
            "",
            "```json",
            json.dumps(gate, indent=2),
            "```",
        ]
    )
    return "\n".join(lines)


def evaluate_health(*, state_path: Path, now: datetime | None = None) -> dict:
    now = now or datetime.now(timezone.utc)
    gate = evaluate_gate().as_json()
    blockers = effective_blockers(gate)
    fingerprint = blockers_fingerprint(blockers)

    if not should_track(gate):
        clear_state(state_path)
        return {
            "issue_title": ISSUE_TITLE,
            "should_track": False,
            "should_alert": False,
            "should_close_issue": True,
            "hours_blocked": 0.0,
            "blockers": blockers,
            "gate": gate,
            "issue_body": "",
        }

    first_seen, previous_fp = parse_state(state_path)
    if first_seen is None or previous_fp != fingerprint:
        first_seen = now
        write_state(state_path, first_seen, fingerprint)
    else:
        write_state(state_path, first_seen, fingerprint)

    hours = hours_blocked(first_seen, now)
    should_alert = hours >= ALERT_HOURS
    return {
        "issue_title": ISSUE_TITLE,
        "should_track": True,
        "should_alert": should_alert,
        "should_close_issue": False,
        "hours_blocked": round(hours, 2),
        "blockers": blockers,
        "gate": gate,
        "issue_body": format_issue_body(gate, blockers, first_seen, hours) if should_alert else "",
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true", help="Print JSON result")
    parser.add_argument(
        "--state-file",
        default=str(ROOT / ".ci-state/automation-health/state.env"),
        help="Path to persisted blocker timing state",
    )
    args = parser.parse_args(argv)

    try:
        result = evaluate_health(state_path=Path(args.state_file))
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
