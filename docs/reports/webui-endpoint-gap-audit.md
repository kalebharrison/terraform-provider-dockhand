# WebUI Endpoint Gap Audit

Date: March 8, 2026

## Scope

Audit current Dockhand API surface from three angles:

1. Public docs reference (`/manual/#api-reference`)
2. Live safe endpoint probe (`scripts/endpoint-probe.py`)
3. Live WebUI bundle crawl (recursive JS import walk from `/login`)

Goal: identify provider coverage gaps and prioritize additions that are useful for Terraform workflows.

## Inputs

- Public reference: [Dockhand Manual API Reference](https://dockhand.pro/manual/#api-reference)
- Probe outputs:
  - `docs/reports/endpoint-probe.csv`
  - `docs/reports/endpoint-probe.md`
- WebUI crawl output:
  - `docs/reports/webui-api-endpoints.txt`

## Current Snapshot

- Probe summary:
  - Total checked routes: `131`
  - Present (non-404): `120`
  - Not present: `2` (`/api/configs`, `/api/backups`)
  - Unverified (fixture missing): `3`
  - Unexpected 404 (placeholder parameter probe): `6`
- WebUI summary:
  - JS bundles crawled: `173`
  - Raw `/api/...` route strings: `112`
  - Normalized unique endpoint shapes: `93`
- WebUI vs provider client literal-path diff:
  - Not currently represented in provider endpoint literals: `59` normalized endpoint shapes

## Priority Gaps To Add Next

These are practical additions for Terraform users and align with current Dockhand behavior.

1. Generic async job support
   - Endpoints: `/api/batch`, `/api/jobs/{jobId}`
   - Why: many UI operations are async; provider currently has endpoint-specific polling logic.
   - Candidate surfaces:
     - `dockhand_job` data source (read status/log lines/result)
     - shared job polling helper used by action resources

2. Environment connectivity validation action
   - Endpoints: `/api/environments/test`, `/api/environments/detect-socket`
   - Why: immediate post-create/post-update validation for environment resources.
   - Candidate surface:
     - `dockhand_environment_test_action`

3. Notification smoke-test action
   - Endpoint: `/api/notifications/test`
   - Why: verify notification configs during CI without destructive changes.
   - Candidate surface:
     - `dockhand_notification_test_action`

## Completed Since This Snapshot

- Registry browser/search + delete coverage:
  - `dockhand_registry_search` data source
  - `dockhand_registry_tags` data source
  - `dockhand_registry_catalog` data source
  - `dockhand_registry_image_delete_action` resource
- Cleanup action coverage:
  - `dockhand_prune_action` resource (`/api/prune/*`)
- Schedule settings/stream coverage:
  - `dockhand_schedule_settings` resource
  - `dockhand_schedule_settings` data source
  - `dockhand_schedule_stream` data source
- Stack path helper coverage:
  - `dockhand_stack_base_path` data source
  - `dockhand_stack_default_path` data source
- System coverage:
  - `dockhand_system` data source
  - `dockhand_system_disk` data source
  - `dockhand_system_files` data source
  - `dockhand_system_file_content` data source

## Lower Priority / Likely UI-Oriented

These appear mainly UX/session preferences and are less useful for Terraform desired state:

- `/api/dashboard/preferences`
- `/api/preferences/*`
- `/api/profile*`
- `/api/settings/theme`
- `/api/legal/*`
- `/api/changelog`
- `/api/host`
- `/api/events`
- `/api/logs/merged`
- `/api/self-update*`
- `/api/auto-update*`

## Enterprise / License-Tier Candidates

These should be tracked but gated by edition/availability:

- `/api/auth/ldap*`
- `/api/auth/oidc*`
- `/api/roles*`
- `/api/audit*`

## Not Present On Tested Instance

Confirmed absent in latest probe:

- `GET /api/configs`
- `GET /api/backups`

Keep these in backlog only until API presence is confirmed on supported versions.
