# Dockhand API To Terraform Matrix

Source: [Dockhand Manual API Reference](https://dockhand.pro/manual/#api-reference)

Live verification artifacts:

- Endpoint probe script: `scripts/endpoint-probe.py`
- Latest safe probe report: `docs/reports/endpoint-probe.md` (March 8, 2026)
- Latest WebUI endpoint inventory: `docs/reports/webui-api-endpoints.txt` (March 7, 2026)
- Latest WebUI/provider gap audit: `docs/reports/webui-endpoint-gap-audit.md` (March 8, 2026)
- Non-present backlog: `docs/non-present-endpoints.md`

## Status Legend

- `implemented`: Provider code exists.
- `partial`: Implemented with assumptions that still need confirmation.
- `planned`: Not implemented yet.

## Provider Configuration

| Terraform Surface | API Input | Notes | Status |
| --- | --- | --- | --- |
| `provider.dockhand.endpoint` | Base URL | Supports `DOCKHAND_ENDPOINT`. | implemented |
| `provider.dockhand.username` | Username | Supports `DOCKHAND_USERNAME`. | implemented |
| `provider.dockhand.password` | Password | Supports `DOCKHAND_PASSWORD`. | implemented |
| `provider.dockhand.mfa_token` | MFA token | Supports `DOCKHAND_MFA_TOKEN`. | implemented |
| `provider.dockhand.auth_provider` | Auth provider | Supports `DOCKHAND_AUTH_PROVIDER`; defaults to `local`. | implemented |
| `provider.dockhand.default_env` | `env` query default | Supports `DOCKHAND_DEFAULT_ENV`. | implemented |
| `provider.dockhand.insecure` | TLS behavior | Disables TLS verification for development. | implemented |
| `provider.dockhand.allow_unauthenticated` | Bootstrap mode | Supports `DOCKHAND_ALLOW_UNAUTHENTICATED`; allows initialization without login credentials for first-install bootstrap flows. | implemented |

## Resources

| Terraform Resource | CRUD Step | API Endpoint | Notes | Status |
| --- | --- | --- | --- | --- |
| `dockhand_stack` | Create | `POST /api/stacks?env={env_id}` | Payload uses `name` and `compose`. | implemented |
| `dockhand_stack` | Read | `GET /api/stacks?env={env_id}` | Reads full list and filters by `name`. | partial |
| `dockhand_stack` | Update runtime | `POST /api/stacks/{name}/start` or `POST /api/stacks/{name}/stop` | `enabled` toggles running state. | implemented |
| `dockhand_stack` | Replace | `DELETE /api/stacks/{name}?force=true` + create | `name`, `env`, `compose` are `ForceNew`. | implemented |
| `dockhand_stack` | Import | `GET /api/stacks` | Import formats: `<name>` or `<env>:<name>`. | implemented |
| `dockhand_user` | Create | `POST /api/users` | Requires `username` + `password`. | implemented |
| `dockhand_user` | Read | `GET /api/users/{id}` | `404` removes from state. | implemented |
| `dockhand_user` | Update | `PUT /api/users/{id}` | Supports email/displayName/isAdmin/isActive/password. | implemented |
| `dockhand_user` | Delete | `DELETE /api/users/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_user` | Import | `GET /api/users/{id}` | Import by ID. | implemented |
| `dockhand_settings_general` | Read | `GET /api/settings/general` | Singleton settings document. | implemented |
| `dockhand_settings_general` | Update | `POST /api/settings/general` | Writes merged settings payload. | implemented |
| `dockhand_auth_settings` | Read | `GET /api/auth/settings` | Singleton authentication settings document. | implemented |
| `dockhand_auth_settings` | Update | `PUT /api/auth/settings` | Writes merged auth settings payload (local/free scope). | implemented |
| `dockhand_license` | Read | `GET /api/license` | Singleton license status document. | implemented |
| `dockhand_license` | Update/Apply | `POST /api/license` | Sets/updates license with name + key. | partial |
| `dockhand_license` | Delete | `DELETE /api/license` | Revokes current license. | partial |
| `dockhand_registry` | Create | `POST /api/registries` | Payload supports name/url/isDefault/username/password. | implemented |
| `dockhand_registry` | Read | `GET /api/registries/{id}` | `404` removes from state. | implemented |
| `dockhand_registry` | Update | `PUT /api/registries/{id}` | Omitting username/password preserves credentials. | implemented |
| `dockhand_registry` | Delete | `DELETE /api/registries/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_registry_image_delete_action` | Delete remote image tag | `DELETE /api/registry/image` | One-shot remote registry delete action for (`registry`,`image`,`tag`) tuples. | implemented |
| `dockhand_git_credential` | Create | `POST /api/git/credentials` | Observed payload supports name/authType/username/password. | partial |
| `dockhand_git_credential` | Read | `GET /api/git/credentials/{id}` | `404` removes from state. | implemented |
| `dockhand_git_credential` | Update | `PUT /api/git/credentials/{id}` | Password is write-only. | partial |
| `dockhand_git_credential` | Delete | `DELETE /api/git/credentials/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_git_repository` | Create | `POST /api/git/repositories` | Observed payload supports name/url/branch/composePath/credentialId/etc. | partial |
| `dockhand_git_repository` | Read | `GET /api/git/repositories/{id}` | `404` removes from state. | implemented |
| `dockhand_git_repository` | Update | `PUT /api/git/repositories/{id}` | Updates repo integration settings. | partial |
| `dockhand_git_repository` | Delete | `DELETE /api/git/repositories/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_git_repository_test_action` | Test connectivity | `POST /api/git/repositories/test` | One-shot repository connectivity test. Supports inline URL payloads or existing `repository_id` resolution. | implemented |
| `dockhand_git_stack` | Create/Read/Update/Delete | `GET/POST/PUT/DELETE /api/git/stacks?env={env_id}` | Manages deployed Git-backed stacks (stack name + repo + compose path) in a target environment. | implemented |
| `dockhand_git_stack_webhook_action` | Trigger webhook | `POST /api/git/stacks/{id}/webhook` | One-shot trigger for git stack deploy/sync webhook flow. | implemented |
| `dockhand_git_stack_deploy_action` | Trigger deploy | `POST /api/git/stacks/{id}/deploy-stream` | One-shot deploy request for git-managed stacks. | implemented |
| `dockhand_git_stack_env_file` | Read available env-file paths | `GET /api/git/stacks/{id}/env-files` | Reads env-file path inventory for a git-managed stack. | implemented |
| `dockhand_git_stack_env_file` | Read selected env-file variables | `POST /api/git/stacks/{id}/env-files` | Reads key/value variables for a selected env file path. | implemented |
| `dockhand_config_set` | Create | `POST /api/config-sets` | Supports name/description/envVars/labels/ports/volumes/networkMode/restartPolicy. | partial |
| `dockhand_config_set` | Read | `GET /api/config-sets/{id}` | `404` removes from state. | implemented |
| `dockhand_config_set` | Update | `PUT /api/config-sets/{id}` | Updates config set settings. | partial |
| `dockhand_config_set` | Delete | `DELETE /api/config-sets/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_notification` | Create | `POST /api/notifications` | Known types observed: `apprise`, `smtp`. | partial |
| `dockhand_notification` | Read | `GET /api/notifications/{id}` | `404` removes from state. | implemented |
| `dockhand_notification` | Update | `PUT /api/notifications/{id}` | Updates config and event types. | partial |
| `dockhand_notification` | Delete | `DELETE /api/notifications/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_notification_test_action` | Send test notification | `POST /api/notifications/test` | One-shot notification test action using existing `notification_id` or inline `type` + `config_json` payload. | implemented |
| `dockhand_environment` | Create | `POST /api/environments`, `POST /api/hawser/tokens` | Supports Docker environment connection + collection settings, including mTLS cert/key fields (`ca_cert`, `client_cert`, `client_key`). `connection_type = "agent"` maps to `hawser-edge` and provisions `agent_token` through Hawser token API. | partial |
| `dockhand_environment` | Read | `GET /api/environments/{id}` | `404` removes from state. | implemented |
| `dockhand_environment` | Update | `PUT /api/environments/{id}`, `POST /api/hawser/tokens` | Updates environment settings, including mTLS cert/key fields. For `connection_type = "agent"`, token changes are applied through Hawser token API. | partial |
| `dockhand_environment` | Update-check settings | `GET/POST /api/environments/{id}/update-check` | Manages `update_check_enabled`, `update_check_auto_update`, `update_check_cron`, and `update_check_vulnerability_criteria`. | implemented |
| `dockhand_environment` | Image-prune settings | `GET/POST /api/environments/{id}/image-prune` | Manages `image_prune_enabled`, `image_prune_cron`, and `image_prune_mode`. | implemented |
| `dockhand_environment` | Timezone settings | `GET/POST /api/environments/{id}/timezone` | Manages environment timezone (`timezone`). | implemented |
| `dockhand_environment` | Vulnerability scanner settings | `GET/POST /api/settings/scanner?env={env_id}` | Manages scanner enable/selection per environment and exposes scanner availability/version status. Optional install enforcement pulls scanner images when missing. | implemented |
| `dockhand_environment_test_action` | Test connectivity | `POST /api/environments/test` | One-shot connectivity validation action for direct/socket/agent payloads with optional apply failure on unsuccessful test. | implemented |
| `dockhand_environment_scanner_action` | Scanner install/remove/update-check actions | `POST /api/images/pull?env={env_id}`, `DELETE /api/settings/scanner?removeImages=true&scanner={name}&env={env_id}`, `GET /api/settings/scanner?checkUpdates=true&env={env_id}` | One-shot scanner operations for install/remove/update-check workflows. | implemented |
| `dockhand_environment` | Delete | `DELETE /api/environments/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_network` | Create | `POST /api/networks?env={env_id}` | Minimal create payload: name + driver (replace-only resource). | partial |
| `dockhand_network` | Read | `GET /api/networks?env={env_id}` | Reads network list and matches by `id`. | partial |
| `dockhand_network` | Delete | `DELETE /api/networks/{id}?env={env_id}` | `404` treated as already deleted. | partial |
| `dockhand_volume` | Create | `POST /api/volumes?env={env_id}` | Minimal create payload: name + driver (replace-only resource). | partial |
| `dockhand_volume` | Read | `GET /api/volumes/{name}/inspect?env={env_id}` | `404` removes from state. | partial |
| `dockhand_volume` | Delete | `DELETE /api/volumes/{name}?force=true&env={env_id}` | `404` treated as already deleted. | partial |
| `dockhand_image` | Create | `POST /api/images/pull?env={env_id}` | Pulls image by reference; then resolves image by tags from list. | partial |
| `dockhand_image` | Read | `GET /api/images?env={env_id}` | Matches by `id`, then by tags if needed. | partial |
| `dockhand_image` | Delete | `DELETE /api/images/{id}?env={env_id}` | `404` treated as already deleted. | partial |
| `dockhand_image_scan_action` | Execute scan | `POST /api/images/scan?env={env_id}` | One-shot image scan action; payload uses `imageName`. | implemented |
| `dockhand_container` | Create | `POST /api/containers?env={env_id}` | Supports create payload for name/image, runtime options, memory/cpu, and capability adds. | partial |
| `dockhand_container` | Read | `GET /api/containers?env={env_id}` | Reads full list and matches by container `id`. | partial |
| `dockhand_container` | Update runtime | `POST /api/containers/{id}/start` or `POST /api/containers/{id}/stop` | `enabled` toggles runtime state. | implemented |
| `dockhand_container` | Delete | `DELETE /api/containers/{id}?env={env_id}` | `404` treated as already deleted. | implemented |
| `dockhand_container` | Import | `GET /api/containers?env={env_id}` | Import formats: `<id>` or `<env>:<id>`. | implemented |
| `dockhand_container_action` | Execute action | `POST /api/containers/{id}/start`, `POST /api/containers/{id}/stop`, `POST /api/containers/{id}/restart` | One-shot runtime action resource with replace-by-trigger behavior. | implemented |
| `dockhand_container_file` | Manage file/directory | `POST /api/containers/{id}/files/create`, `GET/PUT /api/containers/{id}/files/content`, `DELETE /api/containers/{id}/files/delete` | Supports creating `file` or `directory`; content read/write applies to `file` type. | implemented |
| `dockhand_stack_action` | Execute action | `POST /api/stacks/{name}/start`, `POST /api/stacks/{name}/stop`, `POST /api/stacks/{name}/restart`, `POST /api/stacks/{name}/down` | One-shot runtime action resource for stack lifecycle operations. | implemented |
| `dockhand_stack_env` | Read raw env | `GET /api/stacks/{name}/env/raw?env={env_id}` | Reads stack raw `.env` document. | implemented |
| `dockhand_stack_env` | Read secret env variables | `GET /api/stacks/{name}/env?env={env_id}` | Reads stack secret variable objects. | implemented |
| `dockhand_stack_env` | Update raw env | `PUT /api/stacks/{name}/env/raw?env={env_id}` | Writes stack raw `.env` document. | implemented |
| `dockhand_stack_env` | Update secret env variables | `PUT /api/stacks/{name}/env?env={env_id}` | Writes secret variable list (`isSecret=true`). | implemented |
| `dockhand_schedule` | Read | `GET /api/schedules` | Resolves existing schedule by `type` + `schedule_id`. | partial |
| `dockhand_schedule` | Update state | `POST /api/schedules/system/{id}/toggle` or `POST /api/schedules/{type}/{id}/toggle` | Manages pause/resume (`enabled`) for existing schedules. | partial |
| `dockhand_schedule_settings` | Read | `GET /api/schedules/settings` | Reads global schedule settings document. | implemented |
| `dockhand_schedule_settings` | Update | `PUT /api/schedules/settings` | Manages `hide_system_jobs` schedule view setting. | implemented |
| `dockhand_schedule_run_action` | Execute run-now action | `POST /api/schedules/{type}/{id}/run` | One-shot run trigger resource with replace-by-trigger behavior. | implemented |
| `dockhand_prune_action` | Execute cleanup action | `POST /api/prune/all`, `POST /api/prune/containers`, `POST /api/prune/images`, `POST /api/prune/networks`, `POST /api/prune/volumes` | One-shot prune action with optional async job polling when Dockhand returns `jobId`. | implemented |
| `dockhand_batch_action` | Execute async batch operation | `POST /api/batch?env={env_id}` + optional poll `GET /api/jobs/{jobId}` | One-shot async action for batch operations (entity + operation + ids) with optional wait-for-completion and captured job output JSON. | implemented |

## Data Sources

| Terraform Data Source | API Endpoint | Notes | Status |
| --- | --- | --- | --- |
| `dockhand_health` | `GET /api/dashboard/stats?env={env_id}` | Successful request is treated as API health (`status = ok`). | partial |
| `dockhand_activity` | `GET /api/activity` | Returns recent event stream/history for observability. | implemented |
| `dockhand_hawser_status` | `GET /api/hawser/connect` | Reads Hawser websocket endpoint readiness and active connection count. | implemented |
| `dockhand_auth_providers` | `GET /api/auth/providers` | Exposes configured auth providers and default provider (local/free providers in current scope). | implemented |
| `dockhand_environment_detect_socket` | `GET /api/environments/detect-socket` | Reads Dockhand local socket discovery payload (`home_dir`, `socket_paths`, raw `sockets_json`). | implemented |
| `dockhand_registry_search` | `GET /api/registry/search` | Searches remote registry repositories by `term` and optional `registry`. | implemented |
| `dockhand_registry_tags` | `GET /api/registry/tags` | Lists tags for an image repository with optional paging and registry selector. | implemented |
| `dockhand_registry_catalog` | `GET /api/registry/catalog` | Reads raw catalog payload and extracted repository names. | implemented |
| `dockhand_git_preview_env` | `POST /api/git/preview-env` | Previews compose/environment variable requirements by repository URL or saved repository ID. | implemented |
| `dockhand_schedules` | `GET /api/schedules` | Exposes schedule inventory (system cleanup + generated schedules). | implemented |
| `dockhand_schedule_settings` | `GET /api/schedules/settings` | Reads singleton schedule settings payload. | implemented |
| `dockhand_schedule_stream` | `GET /api/schedules/stream` | Captures bounded stream snapshot events for connectivity/observability workflows. | implemented |
| `dockhand_system` | `GET /api/system` | Exposes full system summary payload plus extracted runtime/database/stats sections as JSON. | implemented |
| `dockhand_system_disk` | `GET /api/system/disk?env={env_id}` | Exposes environment-scoped Docker disk usage payload (`diskUsage`). | implemented |
| `dockhand_system_files` | `GET /api/system/files?path=<dir>` | Reads directory entries for host/container file-browser paths. | implemented |
| `dockhand_system_file_content` | `GET /api/system/files/content?path=<file>` | Reads file content and metadata (`size`,`mtime`) for a selected path. | implemented |
| `dockhand_stack_base_path` | `GET /api/stacks/base-path` | Reads global base path used for Dockhand-managed stack directories. | implemented |
| `dockhand_stack_default_path` | `GET /api/stacks/default-path?name=<stack>` | Resolves default directory/compose/env paths for a stack name. | implemented |
| `dockhand_stacks` | `GET /api/stacks?env={env_id}` | Exposes stack list with runtime status and container count. | implemented |
| `dockhand_container_logs` | `GET /api/containers/{id}/logs?env={env_id}&tail={n}` | Reads container logs for debugging/verification workflows. | implemented |
| `dockhand_container_inspect` | `GET /api/containers/{id}?env={env_id}` | Exposes full inspect payload as raw JSON for advanced automation. | implemented |
| `dockhand_container_processes` | `GET /api/containers/{id}/top?env={env_id}` | Reads process table (`Titles` + `Processes`) for running containers. | implemented |
| `dockhand_job` | `GET /api/jobs/{jobId}` | Reads async job status/result/lines from Dockhand job queue. | implemented |

## Additional Endpoints Not Yet Mapped

| API Endpoint Group | Candidate Terraform Surface | Status |
| --- | --- | --- |
| `/api/environments` | additional environment data sources | partial |
| `/api/schedules` | schedule details/advanced actions (`run`, executions history/stream/settings) | partial |
| `/api/images` | image actions (`scan`, `push`) | partial |
| `/api/containers` | exec websocket, upload/download streams, and advanced create/update options coverage | partial |
| `/api/git/repositories/test` | git repository connectivity test action | implemented |
| `/api/git/preview-env` | git preview env data source | implemented |
| `/api/stacks/{name}/env` | broader non-secret env var editing semantics | partial |
| `/api/stacks/base-path` + `/api/stacks/default-path` | stack path helper data sources | implemented |
| `/api/volumes` | advanced volume operations (`clone`, `browse`, import/export) | partial |
| `/api/networks` | advanced network operations (`connect`, inspect details as separate surface) | partial |
| `/api/batch` + `/api/jobs/{jobId}` | generic async job action + job status data source (`run-and-poll`) | implemented |
| `/api/environments/test` + `/api/environments/detect-socket` | environment connectivity validation action + socket discovery data source | implemented |
| `/api/notifications/test` | notification test-send one-shot action | implemented |
| `/api/registry/*` (`search`,`tags`,`catalog`,`image`) | registry catalog/search data sources + remote image delete action | implemented |
| `/api/prune/*` | explicit cleanup action resources | implemented |
| `/api/system*` | system introspection/file-browser data sources | implemented |
| `/api/configs` | config management resource/data source | planned (verified not present on tested instance; `404`) |
| `/api/backups` | backup resource/data source | planned (verified not present on tested instance; `404`) |
| license-tier auth endpoints (LDAP/AD/roles) | auth enterprise resources/data sources | planned |

## Probe Follow-Up (March 8, 2026)

- Not present on tested instance:
  - `GET /api/configs`
  - `GET /api/backups`
- Unverified or unexpected on parameterized routes (fixture-dependent):
  - `GET /api/config-sets/{id}` (`unverified_no_fixture`)
  - `GET /api/notifications/{id}` (`unverified_no_fixture`)
  - `GET /api/git/stacks/{id}/env-files` (`unverified_no_fixture`)
  - `PUT /api/users/{id}` (`unexpected_404` in safe placeholder probe)
  - `DELETE /api/users/{id}` (`unexpected_404` in safe placeholder probe)
  - `PUT /api/environments/{id}` (`unexpected_404` in safe placeholder probe)
  - `DELETE /api/environments/{id}` (`unexpected_404` in safe placeholder probe)
  - `POST /api/git/stacks/{id}/deploy-stream` (`unexpected_404` in safe placeholder probe)
  - `POST /api/git/stacks/{id}/env-files` (`unexpected_404` in safe placeholder probe)

## WebUI Gap Follow-Up (March 8, 2026)

- Recursive WebUI bundle crawl (`/login` entrypoint) discovered `112` raw route strings (`93` normalized unique endpoint shapes).
- Compared against provider API client surface, `59` normalized endpoint shapes are currently not mapped by provider resources/data sources/actions.
- Highest value additions for Terraform workflows:
  - environment connectivity validation action via `/api/environments/test`
  - notification smoke-test action via `/api/notifications/test`
- Known likely-non-Terraform/UI-only endpoints are documented in `docs/reports/webui-endpoint-gap-audit.md` and should not block provider feature parity work.

## Open Contract Questions

1. Exact behavior of `DELETE /api/stacks/{name}?force=true` for server error handling (observed non-2xx even when delete appears to succeed).
2. Whether create/update semantics support true in-place compose updates.
3. Whether auth should always be session-cookie based for provider use.
4. Which endpoints are stable enough for Terraform-managed desired state vs read-only telemetry.
