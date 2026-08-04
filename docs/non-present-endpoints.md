# Non-Present Endpoints Backlog

This file tracks API endpoints that are not currently available on the tested Dockhand instance and should be reconsidered for future provider expansion.

## Last Verified

- Date: August 4, 2026 (CI Acceptance Full / Release Watch)
- Probe: `scripts/endpoint-probe.py` (safe mode)
- Report: `docs/reports/endpoint-probe.md`

## Not Present (404)

- `GET /api/configs`
- `GET /api/backups`

## Present but not yet in provider/probe surface

- `/api/registry/tag-info` — discovered on Dockhand `latest` (2026-08-02); track for future registry tag metadata coverage (issue #244).

## Notes

- These are documented as backlog candidates only; no provider resources/data sources should depend on them until the API is present and stable.
- Compatibility reports refresh via **Compat Reports Sync** on green Acceptance Full / Release Watch runs.
- Paths listed above also serve as the API drift allowlist consumed by `scripts/api-drift-gate.py`.
