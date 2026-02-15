# Non-Present Endpoints Backlog

This file tracks API endpoints that are not currently available on the tested Dockhand instance and should be reconsidered for future provider expansion.

## Last Verified

- Date: February 15, 2026
- Probe: `scripts/endpoint-probe.py` (safe mode)
- Report: `docs/reports/endpoint-probe.md`

## Not Present (404)

- `GET /api/configs`
- `GET /api/backups`

## Notes

- These are documented as backlog candidates only; no provider resources/data sources should depend on them until the API is present and stable.
- Re-run the probe after Dockhand upgrades and update this file when status changes.
