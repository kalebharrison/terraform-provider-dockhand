#!/usr/bin/env python3
"""Keep the automated release loop moving after GITHUB_TOKEN merges.

Recovers missing Release Drafter drafts and dispatches lens/tag workflows
when ``release_gate_check.py`` reports ready. Scheduled watchdog runs are a
trusted trigger, so these dispatches do fire.
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(Path(__file__).resolve().parent))

import release_gate_check as gate  # noqa: E402


def dispatch_workflow(workflow: str) -> None:
    proc = subprocess.run(
        [
            "gh",
            "workflow",
            "run",
            workflow,
            "--repo",
            gate._repo(),
            "--ref",
            "main",
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or proc.stdout.strip() or f"failed to dispatch {workflow}")


def decide_actions(result: gate.GateResult) -> list[str]:
    actions: list[str] = []
    if result.unreleased_commits_on_main > 0 and not result.version:
        actions.append("release-drafter.yml")
    if result.ready_for_lens_dispatch:
        actions.append("agent-release-orchestrate.yml")
    if result.ready_to_tag:
        actions.append("agent-release-tag.yml")
    return actions


def run_watchdog(*, dispatch: bool = True) -> dict:
    result = gate.evaluate_gate(release_watch_mode="strict")
    actions = decide_actions(result)
    dispatched: list[dict] = []
    for workflow in actions:
        entry: dict = {"workflow": workflow}
        if not dispatch:
            entry["status"] = "skipped"
            dispatched.append(entry)
            continue
        try:
            dispatch_workflow(workflow)
            entry["status"] = "ok"
        except RuntimeError as err:
            entry["status"] = str(err)
        dispatched.append(entry)
    return {
        "gate": result.as_json(),
        "actions": dispatched,
        "count": len(dispatched),
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--no-dispatch", action="store_true")
    args = parser.parse_args(argv)

    try:
        payload = run_watchdog(dispatch=not args.no_dispatch)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(json.dumps(payload))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
