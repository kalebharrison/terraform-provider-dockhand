#!/usr/bin/env python3
"""Verify CURSOR_API_KEY against the Cursor Cloud Agents API (GET /v1/me)."""

from __future__ import annotations

import argparse
import base64
import json
import os
import sys
import urllib.error
import urllib.request

ME_URL = "https://api.cursor.com/v1/me"


def verify_api_key(api_key: str) -> dict:
    """Return API key metadata from GET /v1/me.

    Uses Basic auth with the API key as the username (empty password), matching
    Issue Agent Intake and the Cloud Agents API docs.
    """
    request = urllib.request.Request(
        ME_URL,
        method="GET",
        headers={
            "Authorization": "Basic "
            + base64.b64encode(f"{api_key}:".encode()).decode(),
            "Accept": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            body = response.read().decode("utf-8")
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Cloud Agents API HTTP {err.code}: {detail}") from err
    except urllib.error.URLError as err:
        raise RuntimeError(f"Cloud Agents API request failed: {err}") from err

    if not body.strip():
        return {"ok": True}
    payload = json.loads(body)
    if not isinstance(payload, dict):
        raise RuntimeError("Cloud Agents API returned a non-object JSON body")
    return payload


def _public_summary(payload: dict) -> dict:
    """Strip PII fields from /v1/me for safer CI logs."""
    return {
        "ok": True,
        "apiKeyName": payload.get("apiKeyName"),
        "createdAt": payload.get("createdAt"),
        "userId": payload.get("userId"),
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    api_key = os.environ.get("CURSOR_API_KEY", "").strip()
    if not api_key:
        print("CURSOR_API_KEY is not set", file=sys.stderr)
        return 2

    try:
        result = verify_api_key(api_key)
    except RuntimeError as err:
        print(str(err), file=sys.stderr)
        return 1

    summary = _public_summary(result)
    if args.json:
        print(json.dumps(summary, indent=2))
    else:
        name = summary.get("apiKeyName") or "(unnamed)"
        print(f"CURSOR_API_KEY ok for Cloud Agents API (key={name})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
