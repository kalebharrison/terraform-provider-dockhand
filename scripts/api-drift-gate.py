#!/usr/bin/env python3
"""
Gate compatibility runs on newly discovered API routes.

The gate compares discovered API endpoints (from WebUI/docs audits) against:
1) endpoint-probe.py tracked routes (current provider-aware API contract), and
2) docs/non-present-endpoints.md allowlist for known backlog gaps.

If new relevant routes are discovered that are not tracked and not allowlisted,
the script exits non-zero.
"""

from __future__ import annotations

import os
import re
import sys
import json
from pathlib import Path


API_PATH_RE = re.compile(r"/api/[A-Za-z0-9_./{}:-]+")
ALLOWLIST_RE = re.compile(r"`(?:[A-Z]+\s+)?(/api/[A-Za-z0-9_./{}:-]+)`")
PROBE_PATH_RE = re.compile(r'"path"\s*:\s*"(/api/[A-Za-z0-9_./{}:-]+)"')

RELEVANT_PREFIXES = (
    "/api/auth",
    "/api/users",
    "/api/environments",
    "/api/registries",
    "/api/registry",
    "/api/git",
    "/api/config-sets",
    "/api/notifications",
    "/api/stacks",
    "/api/networks",
    "/api/volumes",
    "/api/images",
    "/api/containers",
    "/api/schedules",
    "/api/system",
    "/api/hawser",
    "/api/batch",
    "/api/jobs",
    "/api/prune",
    "/api/dashboard/stats",
)


def normalize(path: str) -> str:
    value = path.strip()
    if not value:
        return value
    if value != "/api":
        value = value.rstrip("/")
    return value


def read_api_lines(path: Path) -> set[str]:
    if not path.exists():
        return set()
    values: set[str] = set()
    for line in path.read_text(encoding="utf-8", errors="ignore").splitlines():
        line = line.strip()
        if not line or not line.startswith("/api/"):
            continue
        values.add(normalize(line))
    return values


def read_probe_paths(path: Path) -> set[str]:
    if not path.exists():
        raise RuntimeError(f"missing probe source file: {path}")
    text = path.read_text(encoding="utf-8", errors="ignore")
    return {normalize(match) for match in PROBE_PATH_RE.findall(text)}


def read_allowlist(path: Path) -> set[str]:
    if not path.exists():
        return set()
    text = path.read_text(encoding="utf-8", errors="ignore")
    return {normalize(match) for match in ALLOWLIST_RE.findall(text)}


def read_paths_from_report(path: Path) -> set[str]:
    if not path.exists():
        return set()
    text = path.read_text(encoding="utf-8", errors="ignore")
    return {normalize(match) for match in API_PATH_RE.findall(text)}


def is_relevant(path: str) -> bool:
    return any(path.startswith(prefix) for prefix in RELEVANT_PREFIXES)


def write_report(
    out_file: Path,
    discovered: set[str],
    tracked: set[str],
    allowlisted: set[str],
    unknown: set[str],
    unknown_relevant: set[str],
    unknown_relevant_unallowlisted: set[str],
) -> None:
    lines: list[str] = []
    lines.append("# API Drift Gate Report")
    lines.append("")
    lines.append("## Snapshot")
    lines.append("")
    lines.append(f"- Discovered endpoints (docs + webui): `{len(discovered)}`")
    lines.append(f"- Tracked probe endpoints: `{len(tracked)}`")
    lines.append(f"- Allowlisted non-present endpoints: `{len(allowlisted)}`")
    lines.append(f"- Unknown endpoints: `{len(unknown)}`")
    lines.append(f"- Unknown relevant endpoints: `{len(unknown_relevant)}`")
    lines.append(
        f"- New relevant endpoints not allowlisted: `{len(unknown_relevant_unallowlisted)}`"
    )
    lines.append("")

    lines.append("## New Relevant Endpoints (Failing)")
    lines.append("")
    if unknown_relevant_unallowlisted:
        for value in sorted(unknown_relevant_unallowlisted):
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Unknown But Allowlisted")
    lines.append("")
    known_backlog = sorted(x for x in unknown_relevant if x in allowlisted)
    if known_backlog:
        for value in known_backlog:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    lines.append("## Unknown Non-Relevant")
    lines.append("")
    unknown_non_relevant = sorted(x for x in unknown if not is_relevant(x))
    if unknown_non_relevant:
        for value in unknown_non_relevant:
            lines.append(f"- `{value}`")
    else:
        lines.append("- None")
    lines.append("")

    out_file.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    repo_root = Path(__file__).resolve().parents[1]
    out_dir = Path(os.getenv("API_DRIFT_OUT_DIR", "docs/reports"))
    out_dir.mkdir(parents=True, exist_ok=True)

    webui_file = Path(
        os.getenv("WEBUI_ENDPOINTS_FILE", str(out_dir / "webui-api-endpoints.txt"))
    )
    docs_file = Path(
        os.getenv("DOCS_ENDPOINTS_FILE", str(out_dir / "docs-reference-api-endpoints.txt"))
    )
    docs_report_file = Path(
        os.getenv("DOCS_REPORT_FILE", str(out_dir / "docs-reference-gap-audit.md"))
    )
    baseline_webui_file = Path(
        os.getenv(
            "BASELINE_WEBUI_ENDPOINTS_FILE",
            str(repo_root / "docs/reports/webui-api-endpoints.txt"),
        )
    )
    baseline_docs_file = Path(
        os.getenv(
            "BASELINE_DOCS_ENDPOINTS_FILE",
            str(repo_root / "docs/reports/docs-reference-api-endpoints.txt"),
        )
    )
    allowlist_file = Path(
        os.getenv("KNOWN_MISSING_ENDPOINTS_FILE", str(repo_root / "docs/non-present-endpoints.md"))
    )
    probe_source = Path(
        os.getenv("PROBE_SOURCE_FILE", str(repo_root / "scripts/endpoint-probe.py"))
    )
    report_file = Path(os.getenv("API_DRIFT_REPORT_FILE", str(out_dir / "api-drift-gate.md")))
    json_file = Path(os.getenv("API_DRIFT_JSON_FILE", str(out_dir / "api-drift-gate.json")))
    fail_on_new = os.getenv("API_DRIFT_FAIL_ON_NEW", "true").lower() in ("1", "true", "yes")

    current_webui = read_api_lines(webui_file)
    current_docs = read_api_lines(docs_file)
    if not current_docs:
        current_docs = read_paths_from_report(docs_report_file)
    discovered = set()
    discovered.update(current_webui)
    discovered.update(current_docs)

    baseline_webui = read_api_lines(baseline_webui_file)
    baseline_docs = read_api_lines(baseline_docs_file)
    baseline_discovered = set()
    baseline_discovered.update(baseline_webui)
    baseline_discovered.update(baseline_docs)

    new_webui = current_webui - baseline_webui
    new_docs = set()
    if baseline_docs_file.exists():
        new_docs = current_docs - baseline_docs
    newly_discovered = new_webui | new_docs

    tracked = read_probe_paths(probe_source)
    allowlisted = read_allowlist(allowlist_file)

    unknown = {x for x in discovered if x not in tracked}
    unknown_relevant = {x for x in unknown if is_relevant(x)}
    unknown_relevant_unallowlisted = {
        x for x in unknown_relevant if x in newly_discovered and x not in allowlisted
    }

    write_report(
        report_file,
        discovered=discovered,
        tracked=tracked,
        allowlisted=allowlisted,
        unknown=unknown,
        unknown_relevant=unknown_relevant,
        unknown_relevant_unallowlisted=unknown_relevant_unallowlisted,
    )

    payload = {
        "discovered": sorted(discovered),
        "tracked": sorted(tracked),
        "allowlisted": sorted(allowlisted),
        "unknown": sorted(unknown),
        "unknown_relevant": sorted(unknown_relevant),
        "unknown_relevant_unallowlisted": sorted(unknown_relevant_unallowlisted),
        "counts": {
            "discovered": len(discovered),
            "baseline": len(baseline_discovered),
            "newly_discovered": len(newly_discovered),
            "tracked": len(tracked),
            "unknown": len(unknown),
            "unknown_relevant": len(unknown_relevant),
            "new_relevant_unallowlisted": len(unknown_relevant_unallowlisted),
        },
    }
    json_file.write_text(json.dumps(payload, indent=2), encoding="utf-8")

    print(f"wrote {report_file}")
    print(f"wrote {json_file}")
    print(
        "summary:",
        f"discovered={len(discovered)}",
        f"baseline={len(baseline_discovered)}",
        f"newly_discovered={len(newly_discovered)}",
        f"tracked={len(tracked)}",
        f"unknown={len(unknown)}",
        f"unknown_relevant={len(unknown_relevant)}",
        f"new_relevant_unallowlisted={len(unknown_relevant_unallowlisted)}",
    )

    if fail_on_new and unknown_relevant_unallowlisted:
        sys.stderr.write(
            "api drift gate failed: new relevant endpoints discovered and not allowlisted in "
            f"{allowlist_file}\n"
        )
        return 2

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
