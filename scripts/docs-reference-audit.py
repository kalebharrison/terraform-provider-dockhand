#!/usr/bin/env python3
"""
Audit Dockhand public docs API surface against provider/probe endpoint lists.

Outputs:
  - <out_dir>/docs-reference-api-endpoints.txt
  - <out_dir>/docs-reference-gap-audit.md

Environment:
  DOCKHAND_DOCS_URL    Optional. Default: https://dockhand.pro/manual/#api-reference
  DOCS_AUDIT_OUT_DIR   Optional. Default: docs/reports
"""

from __future__ import annotations

import datetime as dt
import html
import os
import re
import sys
import urllib.request
from pathlib import Path
from typing import Iterable


API_RE = re.compile(r"/api/[A-Za-z0-9_./{}:-]+")
TRAIL_RE = re.compile(r"""["'\)\],;]+$""")


def normalize(path: str) -> str:
    value = path.strip()
    if not value:
        return value
    value = TRAIL_RE.sub("", value)
    if value != "/api":
        value = value.rstrip("/")
    return value


def fetch_text(url: str) -> str:
    req = urllib.request.Request(
        url,
        headers={"User-Agent": "terraform-provider-dockhand-docs-audit/1.0"},
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return resp.read().decode("utf-8", errors="ignore")


def extract_api_routes(text: str) -> set[str]:
    decoded = html.unescape(text)
    return {normalize(m) for m in API_RE.findall(decoded)}


def iter_provider_files(repo_root: Path) -> Iterable[Path]:
    root = repo_root / "internal" / "provider"
    if not root.exists():
        return []
    return root.rglob("*.go")


def extract_provider_endpoints(repo_root: Path) -> set[str]:
    values: set[str] = set()
    for file_path in iter_provider_files(repo_root):
        text = file_path.read_text(encoding="utf-8", errors="ignore")
        values.update(extract_api_routes(text))
    return values


def extract_probe_endpoints(repo_root: Path) -> set[str]:
    probe_path = repo_root / "scripts" / "endpoint-probe.py"
    if not probe_path.exists():
        return set()
    text = probe_path.read_text(encoding="utf-8", errors="ignore")
    return extract_api_routes(text)


def write_list(path: Path, values: list[str]) -> None:
    body = "\n".join(values)
    if body:
        body += "\n"
    path.write_text(body, encoding="utf-8")


def write_report(
    report_path: Path,
    docs_url: str,
    docs_endpoints: set[str],
    provider_endpoints: set[str],
    probe_endpoints: set[str],
) -> None:
    missing_provider = sorted(x for x in docs_endpoints if x not in provider_endpoints)
    missing_probe = sorted(x for x in docs_endpoints if x not in probe_endpoints)

    lines: list[str] = []
    lines.append("# Docs Reference Gap Audit")
    lines.append("")
    lines.append(f"Date: {dt.date.today().isoformat()}")
    lines.append("")
    lines.append("## Source")
    lines.append("")
    lines.append(f"- URL: {docs_url}")
    lines.append("")
    lines.append("## Snapshot")
    lines.append("")
    lines.append(f"- Docs endpoint shapes: `{len(docs_endpoints)}`")
    lines.append(f"- Provider endpoint literals: `{len(provider_endpoints)}`")
    lines.append(f"- Probe endpoint literals: `{len(probe_endpoints)}`")
    lines.append(f"- Docs endpoints missing in provider literals: `{len(missing_provider)}`")
    lines.append(f"- Docs endpoints missing in probe list: `{len(missing_probe)}`")
    lines.append("")

    lines.append("## Missing In Provider Literals")
    lines.append("")
    if missing_provider:
        for value in missing_provider:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Missing In Probe List")
    lines.append("")
    if missing_probe:
        for value in missing_probe:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    report_path.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    repo_root = Path(__file__).resolve().parents[1]
    docs_url = os.getenv("DOCKHAND_DOCS_URL", "https://dockhand.pro/manual/#api-reference").strip()
    out_dir = Path(os.getenv("DOCS_AUDIT_OUT_DIR", "docs/reports"))
    out_dir.mkdir(parents=True, exist_ok=True)

    docs_html = fetch_text(docs_url)
    docs_endpoints = extract_api_routes(docs_html)
    provider_endpoints = extract_provider_endpoints(repo_root)
    probe_endpoints = extract_probe_endpoints(repo_root)

    endpoints_path = out_dir / "docs-reference-api-endpoints.txt"
    report_path = out_dir / "docs-reference-gap-audit.md"

    write_list(endpoints_path, sorted(docs_endpoints))
    write_report(
        report_path,
        docs_url=docs_url,
        docs_endpoints=docs_endpoints,
        provider_endpoints=provider_endpoints,
        probe_endpoints=probe_endpoints,
    )

    print(f"wrote {endpoints_path}")
    print(f"wrote {report_path}")
    print(
        "summary:",
        f"docs={len(docs_endpoints)}",
        f"provider={len(provider_endpoints)}",
        f"probe={len(probe_endpoints)}",
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # pragma: no cover
        sys.stderr.write(f"docs reference audit failed: {exc}\n")
        raise SystemExit(1)
