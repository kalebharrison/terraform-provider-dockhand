#!/usr/bin/env python3
"""
Probe one authenticated/private Dockhand endpoint and write a report.

Outputs:
  - <out_dir>/private-endpoint-probe.json
  - <out_dir>/private-endpoint-probe.md

Environment:
  DOCKHAND_ENDPOINT        Required
  DOCKHAND_USERNAME        Required
  DOCKHAND_PASSWORD        Required
  DOCKHAND_AUTH_PROVIDER   Optional. Default: local
  DOCKHAND_INSECURE        Optional. true/false
  PRIVATE_ENDPOINT_METHOD  Optional. Default: GET
  PRIVATE_ENDPOINT_PATH    Optional. Default: /api/health
  PRIVATE_ENDPOINT_QUERY   Optional. URL query string (without '?')
  PRIVATE_ENDPOINT_BODY    Optional. JSON object string for POST/PUT
  PRIVATE_PROBE_OUT_DIR    Optional. Default: docs/reports
"""

from __future__ import annotations

import datetime as dt
import json
import os
import re
import ssl
import sys
import urllib.error
import urllib.parse
import urllib.request
from http.cookiejar import CookieJar
from pathlib import Path


def getenv_required(name: str) -> str:
    value = os.getenv(name, "").strip()
    if not value:
        raise RuntimeError(f"missing required environment variable: {name}")
    return value


class Session:
    def __init__(self, endpoint: str, insecure: bool) -> None:
        self.endpoint = endpoint.rstrip("/")
        cookies = CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookies))
        if self.endpoint.startswith("https://") and insecure:
            ctx = ssl.create_default_context()
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE
            self.opener.add_handler(urllib.request.HTTPSHandler(context=ctx))

    def request(
        self,
        method: str,
        path: str,
        body: dict | None = None,
        query: dict | None = None,
    ) -> tuple[int, str]:
        url = self.endpoint + path
        if query:
            url += "?" + urllib.parse.urlencode(query)
        payload = None
        if body is not None:
            payload = json.dumps(body).encode("utf-8")

        req = urllib.request.Request(url, data=payload, method=method)
        req.add_header("Content-Type", "application/json")
        try:
            with self.opener.open(req, timeout=20) as resp:
                return resp.status, resp.read().decode("utf-8", errors="replace")
        except urllib.error.HTTPError as exc:
            return exc.code, exc.read().decode("utf-8", errors="replace")


def summarize(status: int) -> str:
    if status == 404:
        return "not_present_or_not_found"
    if 200 <= status < 300:
        return "present"
    if status in (400, 401, 403, 405):
        return "present_or_rejected"
    return "unknown"


def parse_optional_json(text: str) -> dict | None:
    raw = text.strip()
    if not raw:
        return None
    parsed = json.loads(raw)
    if isinstance(parsed, dict):
        return parsed
    raise RuntimeError("PRIVATE_ENDPOINT_BODY must be a JSON object")


def parse_query(raw: str) -> dict[str, str]:
    if not raw.strip():
        return {}
    parsed = urllib.parse.parse_qs(raw, keep_blank_values=True)
    return {k: v[0] if v else "" for k, v in parsed.items()}


def scrub_response_preview(text: str) -> str:
    if not text:
        return text
    patterns = (
        r'("(?:password|token|agent_token|agentToken|hawserToken|secret|apiKey|api_token)")\s*:\s*"[^"]*"',
        r'("(?:cert|key|privateKey|certificate)")\s*:\s*"[^"]*"',
    )
    scrubbed = text
    for pattern in patterns:
        scrubbed = re.sub(pattern, r'\1:"[redacted]"', scrubbed, flags=re.IGNORECASE)
    return scrubbed


def main() -> int:
    endpoint = getenv_required("DOCKHAND_ENDPOINT")
    username = getenv_required("DOCKHAND_USERNAME")
    password = getenv_required("DOCKHAND_PASSWORD")
    auth_provider = os.getenv("DOCKHAND_AUTH_PROVIDER", "local").strip() or "local"
    insecure = os.getenv("DOCKHAND_INSECURE", "false").lower() in ("1", "true", "yes")

    method = os.getenv("PRIVATE_ENDPOINT_METHOD", "GET").strip().upper()
    path = os.getenv("PRIVATE_ENDPOINT_PATH", "/api/health").strip() or "/api/health"
    query_raw = os.getenv("PRIVATE_ENDPOINT_QUERY", "")
    body_raw = os.getenv("PRIVATE_ENDPOINT_BODY", "")

    out_dir = Path(os.getenv("PRIVATE_PROBE_OUT_DIR", "docs/reports"))
    out_dir.mkdir(parents=True, exist_ok=True)

    session = Session(endpoint, insecure)
    login_status, login_body = session.request(
        "POST",
        "/api/auth/login",
        body={"username": username, "password": password, "authProvider": auth_provider},
    )
    if login_status < 200 or login_status >= 300:
        raise RuntimeError(f"login failed with status {login_status}: {login_body[:200]}")

    status, response = session.request(
        method,
        path,
        body=parse_optional_json(body_raw),
        query=parse_query(query_raw),
    )

    now = dt.datetime.utcnow().replace(microsecond=0).isoformat() + "Z"
    result = summarize(status)
    preview = scrub_response_preview(response[:1000])

    payload = {
        "timestamp_utc": now,
        "endpoint": endpoint,
        "method": method,
        "path": path,
        "query": parse_query(query_raw),
        "status": status,
        "result": result,
        "response_preview": preview,
    }

    json_file = out_dir / "private-endpoint-probe.json"
    md_file = out_dir / "private-endpoint-probe.md"
    json_file.write_text(json.dumps(payload, indent=2), encoding="utf-8")

    lines: list[str] = []
    lines.append("# Private Endpoint Probe")
    lines.append("")
    lines.append(f"- Timestamp (UTC): `{now}`")
    lines.append(f"- Endpoint base: `{endpoint}`")
    lines.append(f"- Request: `{method} {path}`")
    lines.append(f"- Query: `{query_raw or '(none)'}`")
    lines.append(f"- HTTP status: `{status}`")
    lines.append(f"- Result: `{result}`")
    lines.append("")
    lines.append("## Response Preview")
    lines.append("")
    lines.append("```text")
    lines.append(preview if preview else "(empty)")
    lines.append("```")
    lines.append("")
    md_file.write_text("\n".join(lines), encoding="utf-8")

    print(f"wrote {json_file}")
    print(f"wrote {md_file}")
    print(f"summary: method={method} path={path} status={status} result={result}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover
        sys.stderr.write(f"private endpoint probe failed: {exc}\n")
        raise SystemExit(1)
