#!/usr/bin/env python3
"""
Crawl Dockhand WebUI bundles and audit API route coverage.

Outputs:
  - <out_dir>/webui-api-endpoints.txt
  - <out_dir>/webui-endpoint-gap-audit.md

Environment:
  DOCKHAND_ENDPOINT        Required. Example: http://127.0.0.1:13001
  WEBUI_AUDIT_OUT_DIR      Optional. Default: docs/reports
  WEBUI_AUDIT_MAX_BUNDLES  Optional. Default: 4000
"""

from __future__ import annotations

import datetime as dt
import os
import re
import sys
import urllib.parse
import urllib.request
from collections import deque
from pathlib import Path
from typing import Iterable


JS_REF_RE = re.compile(r"""['"]([^'"\s]+\.js[^'"\s]*)['"]""")
API_RE = re.compile(r"""/api/[A-Za-z0-9_./{}:-]+""")
API_TRAIL_RE = re.compile(r"""["'\)\],;]+$""")
GO_ENDPOINT_RE = re.compile(r"""/api/[A-Za-z0-9_./{}:-]*""")


def getenv_required(name: str) -> str:
    value = os.getenv(name, "").strip()
    if not value:
        raise RuntimeError(f"missing required environment variable: {name}")
    return value


def normalize_endpoint(path: str) -> str:
    value = path.strip()
    if not value:
        return value
    if value != "/api":
        value = value.rstrip("/")
    return value


def fetch_text(url: str) -> str:
    req = urllib.request.Request(
        url,
        headers={"User-Agent": "terraform-provider-dockhand-webui-audit/1.0"},
    )
    with urllib.request.urlopen(req, timeout=20) as resp:
        return resp.read().decode("utf-8", errors="ignore")


def extract_api_routes(text: str) -> set[str]:
    values: set[str] = set()
    for match in API_RE.findall(text):
        values.add(API_TRAIL_RE.sub("", match))
    return values


def iter_go_files(repo_root: Path) -> Iterable[Path]:
    provider_dir = repo_root / "internal" / "provider"
    if not provider_dir.exists():
        return []
    return provider_dir.rglob("*.go")


def extract_provider_endpoints(repo_root: Path) -> set[str]:
    found: set[str] = set()
    for file_path in iter_go_files(repo_root):
        text = file_path.read_text(encoding="utf-8", errors="ignore")
        for endpoint in GO_ENDPOINT_RE.findall(text):
            found.add(normalize_endpoint(endpoint))
    return found


def extract_probe_endpoints(repo_root: Path) -> set[str]:
    probe_file = repo_root / "scripts" / "endpoint-probe.py"
    if not probe_file.exists():
        return set()
    text = probe_file.read_text(encoding="utf-8", errors="ignore")
    return {normalize_endpoint(x) for x in GO_ENDPOINT_RE.findall(text)}


def crawl_webui(base_url: str, max_bundles: int) -> tuple[int, set[str]]:
    login_url = base_url + "/login"
    base_host = urllib.parse.urlparse(base_url).netloc

    html = fetch_text(login_url)
    queue: deque[str] = deque()
    visited: set[str] = set()
    endpoints: set[str] = set()

    for ref in JS_REF_RE.findall(html):
        resolved = urllib.parse.urljoin(login_url, ref.split("?", 1)[0])
        parsed = urllib.parse.urlparse(resolved)
        if parsed.netloc == base_host:
            queue.append(resolved)

    while queue and len(visited) < max_bundles:
        url = queue.popleft()
        if url in visited:
            continue
        visited.add(url)

        parsed = urllib.parse.urlparse(url)
        if parsed.netloc != base_host:
            continue

        try:
            js_text = fetch_text(url)
        except Exception:
            continue

        endpoints.update(extract_api_routes(js_text))

        for ref in JS_REF_RE.findall(js_text):
            ref = ref.split("?", 1)[0]
            resolved = urllib.parse.urljoin(url, ref)
            parsed_ref = urllib.parse.urlparse(resolved)
            if parsed_ref.netloc != base_host:
                continue
            if not parsed_ref.path.endswith(".js"):
                continue
            if "/_app/" not in parsed_ref.path:
                continue
            if resolved not in visited:
                queue.append(resolved)

    return len(visited), endpoints


def by_prefix(values: Iterable[str], prefixes: list[str]) -> list[str]:
    picked: list[str] = []
    for value in sorted(set(values)):
        if any(value.startswith(prefix) for prefix in prefixes):
            picked.append(value)
    return picked


def write_text(path: Path, values: list[str]) -> None:
    body = "\n".join(values)
    if body:
        body += "\n"
    path.write_text(body, encoding="utf-8")


def write_markdown(
    out_file: Path,
    bundles_fetched: int,
    webui_endpoints: set[str],
    provider_endpoints: set[str],
    probe_endpoints: set[str],
) -> None:
    normalized_webui = {normalize_endpoint(x) for x in webui_endpoints}
    missing_provider = sorted(x for x in normalized_webui if x not in provider_endpoints)
    missing_probe = sorted(x for x in normalized_webui if x not in probe_endpoints)

    priority = by_prefix(
        missing_provider,
        [
            "/api/batch",
            "/api/jobs",
            "/api/environments/test",
            "/api/environments/detect-socket",
            "/api/notifications/test",
            "/api/registry/",
            "/api/prune/",
        ],
    )
    enterprise = by_prefix(
        missing_provider,
        [
            "/api/auth/ldap",
            "/api/auth/oidc",
            "/api/roles",
            "/api/audit",
        ],
    )
    likely_ui_only = by_prefix(
        missing_provider,
        [
            "/api/dashboard/preferences",
            "/api/preferences/",
            "/api/profile",
            "/api/settings/theme",
            "/api/legal/",
            "/api/changelog",
            "/api/host",
            "/api/events",
            "/api/logs/merged",
            "/api/self-update",
            "/api/auto-update",
        ],
    )

    lines: list[str] = []
    lines.append("# WebUI Endpoint Gap Audit")
    lines.append("")
    lines.append(f"Date: {dt.date.today().isoformat()}")
    lines.append("")
    lines.append("## Snapshot")
    lines.append("")
    lines.append(f"- JS bundles crawled: `{bundles_fetched}`")
    lines.append(f"- WebUI raw endpoint strings: `{len(webui_endpoints)}`")
    lines.append(f"- WebUI normalized endpoint shapes: `{len(normalized_webui)}`")
    lines.append(f"- Provider endpoint literals: `{len(provider_endpoints)}`")
    lines.append(f"- Probe endpoint literals: `{len(probe_endpoints)}`")
    lines.append(f"- WebUI not currently in provider literals: `{len(missing_provider)}`")
    lines.append(f"- WebUI not currently in probe list: `{len(missing_probe)}`")
    lines.append("")

    lines.append("## Priority Candidates")
    lines.append("")
    if priority:
        for value in priority:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Enterprise Candidates")
    lines.append("")
    if enterprise:
        for value in enterprise:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Likely UI-Oriented")
    lines.append("")
    if likely_ui_only:
        for value in likely_ui_only:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Missing From Provider Literals (Full)")
    lines.append("")
    if missing_provider:
        for value in missing_provider:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Missing From Probe List (Full)")
    lines.append("")
    if missing_probe:
        for value in missing_probe:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    out_file.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    repo_root = Path(__file__).resolve().parents[1]
    endpoint = getenv_required("DOCKHAND_ENDPOINT").rstrip("/")
    out_dir = Path(os.getenv("WEBUI_AUDIT_OUT_DIR", "docs/reports"))
    max_bundles = int(os.getenv("WEBUI_AUDIT_MAX_BUNDLES", "4000"))
    out_dir.mkdir(parents=True, exist_ok=True)

    bundles, webui_endpoints = crawl_webui(endpoint, max_bundles)
    provider_endpoints = extract_provider_endpoints(repo_root)
    probe_endpoints = extract_probe_endpoints(repo_root)

    endpoints_file = out_dir / "webui-api-endpoints.txt"
    report_file = out_dir / "webui-endpoint-gap-audit.md"

    write_text(endpoints_file, sorted(webui_endpoints))
    write_markdown(
        report_file,
        bundles_fetched=bundles,
        webui_endpoints=webui_endpoints,
        provider_endpoints=provider_endpoints,
        probe_endpoints=probe_endpoints,
    )

    print(f"wrote {endpoints_file}")
    print(f"wrote {report_file}")
    print(
        "summary:",
        f"bundles={bundles}",
        f"webui_routes={len(webui_endpoints)}",
        f"provider_literals={len(provider_endpoints)}",
        f"probe_literals={len(probe_endpoints)}",
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover
        sys.stderr.write(f"webui endpoint audit failed: {exc}\n")
        raise SystemExit(1)
