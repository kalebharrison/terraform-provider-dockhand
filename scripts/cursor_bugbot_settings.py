#!/usr/bin/env python3
"""Manage Cursor Bugbot repository settings via the Cursor API."""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DEFAULT_REPO_URL = "https://github.com/kalebharrison/terraform-provider-dockhand"
API_URL = "https://api.cursor.com/bugbot/repo/update"

# Bugbot admin API is Team/Enterprise-oriented. User API keys used for Cloud
# Agents commonly receive "Invalid Team API Key" here; treat that as optional.
_UNAVAILABLE_MARKERS = (
    "invalid team api key",
    "team api key",
    "not a team",
    "enterprise",
    "http 401",
    "http 403",
)


def _repo_url(explicit: str | None) -> str:
    if explicit:
        return explicit
    env = os.environ.get("GITHUB_REPOSITORY", "").strip()
    if env:
        return f"https://github.com/{env}"
    return DEFAULT_REPO_URL


def update_bugbot(
    *,
    enabled: bool,
    repo_url: str,
    api_key: str,
    manual_trigger_only: bool | None = None,
) -> dict:
    payload: dict[str, object] = {
        "repoUrl": repo_url,
        "enabled": enabled,
    }
    if manual_trigger_only is not None:
        payload["manualTriggerOnly"] = manual_trigger_only

    request = urllib.request.Request(
        API_URL,
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            body = response.read().decode("utf-8")
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Bugbot API HTTP {err.code}: {detail}") from err
    except urllib.error.URLError as err:
        raise RuntimeError(f"Bugbot API request failed: {err}") from err

    if not body.strip():
        return {"repoUrl": repo_url, "enabled": enabled}
    return json.loads(body)


def is_unavailable_error(err: BaseException) -> bool:
    """Return True when Bugbot API rejects a non-team / non-admin key."""
    text = str(err).lower()
    return any(marker in text for marker in _UNAVAILABLE_MARKERS)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--enable", action="store_true")
    parser.add_argument("--disable", action="store_true")
    parser.add_argument("--manual-trigger-only", action="store_true")
    parser.add_argument("--repo-url", default="")
    parser.add_argument("--json", action="store_true")
    parser.add_argument(
        "--best-effort",
        action="store_true",
        help=(
            "Treat Bugbot API failures as skipped (exit 0). Use in Secret Smoke: "
            "Bugbot admin API is Team/Enterprise-only; Cloud Agents user keys are enough."
        ),
    )
    args = parser.parse_args(argv)

    if args.enable == args.disable:
        parser.error("specify exactly one of --enable or --disable")

    api_key = os.environ.get("CURSOR_API_KEY", "").strip()
    if not api_key:
        print("CURSOR_API_KEY is not set", file=sys.stderr)
        return 2

    try:
        result = update_bugbot(
            enabled=args.enable,
            repo_url=_repo_url(args.repo_url or None),
            api_key=api_key,
            manual_trigger_only=True if args.manual_trigger_only else None,
        )
    except RuntimeError as err:
        if args.best_effort:
            payload = {
                "skipped": True,
                "reason": str(err),
                "unavailable": is_unavailable_error(err),
            }
            print(
                "Bugbot API unavailable or failed; skipping "
                f"(best-effort): {err}",
                file=sys.stderr,
            )
            if args.json:
                print(json.dumps(payload, indent=2))
            return 0
        print(str(err), file=sys.stderr)
        return 1

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
