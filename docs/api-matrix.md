# Dockhand API To Terraform Matrix

Source: [Dockhand Manual API Reference](https://dockhand.pro/manual/#api-reference)

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
| `dockhand_git_credential` | Create | `POST /api/git/credentials` | Observed payload supports name/authType/username/password. | partial |
| `dockhand_git_credential` | Read | `GET /api/git/credentials/{id}` | `404` removes from state. | implemented |
| `dockhand_git_credential` | Update | `PUT /api/git/credentials/{id}` | Password is write-only. | partial |
| `dockhand_git_credential` | Delete | `DELETE /api/git/credentials/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_git_repository` | Create | `POST /api/git/repositories` | Observed payload supports name/url/branch/composePath/credentialId/etc. | partial |
| `dockhand_git_repository` | Read | `GET /api/git/repositories/{id}` | `404` removes from state. | implemented |
| `dockhand_git_repository` | Update | `PUT /api/git/repositories/{id}` | Updates repo integration settings. | partial |
| `dockhand_git_repository` | Delete | `DELETE /api/git/repositories/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_git_stack_webhook_action` | Trigger webhook | `POST /api/git/stacks/{id}/webhook` | One-shot trigger for git stack deploy/sync webhook flow. | implemented |
| `dockhand_config_set` | Create | `POST /api/config-sets` | Supports name/description/envVars/labels/ports/volumes/networkMode/restartPolicy. | partial |
| `dockhand_config_set` | Read | `GET /api/config-sets/{id}` | `404` removes from state. | implemented |
| `dockhand_config_set` | Update | `PUT /api/config-sets/{id}` | Updates config set settings. | partial |
| `dockhand_config_set` | Delete | `DELETE /api/config-sets/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_notification` | Create | `POST /api/notifications` | Known types observed: `apprise`, `smtp`. | partial |
| `dockhand_notification` | Read | `GET /api/notifications/{id}` | `404` removes from state. | implemented |
| `dockhand_notification` | Update | `PUT /api/notifications/{id}` | Updates config and event types. | partial |
| `dockhand_notification` | Delete | `DELETE /api/notifications/{id}` | `404` treated as already deleted. | implemented |
| `dockhand_environment` | Create | `POST /api/environments` | Supports Docker environment connection + collection settings. | partial |
| `dockhand_environment` | Read | `GET /api/environments/{id}` | `404` removes from state. | implemented |
| `dockhand_environment` | Update | `PUT /api/environments/{id}` | Updates environment settings. | partial |
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
| `dockhand_stack_action` | Execute action | `POST /api/stacks/{name}/start`, `POST /api/stacks/{name}/stop` | One-shot runtime action resource; `restart` runs stop then start. | implemented |
| `dockhand_schedule` | Read | `GET /api/schedules` | Resolves existing schedule by `type` + `schedule_id`. | partial |
| `dockhand_schedule` | Update state | `POST /api/schedules/system/{id}/toggle` or `POST /api/schedules/{type}/{id}/toggle` | Manages pause/resume (`enabled`) for existing schedules. | partial |

## Data Sources

| Terraform Data Source | API Endpoint | Notes | Status |
| --- | --- | --- | --- |
| `dockhand_health` | `GET /api/dashboard/stats?env={env_id}` | Successful request is treated as API health (`status = ok`). | partial |
| `dockhand_activity` | `GET /api/activity` | Returns recent event stream/history for observability. | implemented |
| `dockhand_hawser_status` | `GET /api/hawser/connect` | Reads Hawser websocket endpoint readiness and active connection count. | implemented |
| `dockhand_auth_providers` | `GET /api/auth/providers` | Exposes configured auth providers and default provider (local/free providers in current scope). | implemented |
| `dockhand_schedules` | `GET /api/schedules` | Exposes schedule inventory (system cleanup + generated schedules). | implemented |
| `dockhand_stacks` | `GET /api/stacks?env={env_id}` | Exposes stack list with runtime status and container count. | implemented |
| `dockhand_container_logs` | `GET /api/containers/{id}/logs?env={env_id}&tail={n}` | Reads container logs for debugging/verification workflows. | implemented |
| `dockhand_container_inspect` | `GET /api/containers/{id}?env={env_id}` | Exposes full inspect payload as raw JSON for advanced automation. | implemented |

## Additional Endpoints Not Yet Mapped

| API Endpoint Group | Candidate Terraform Surface | Status |
| --- | --- | --- |
| `/api/environments` | additional environment data sources | partial |
| `/api/schedules` | schedule details/advanced actions (`run`, executions history/settings) | partial |
| `/api/images` | image actions (`scan`, `push`) | partial |
| `/api/containers` | container logs, restart action, and advanced create options coverage | partial |
| `/api/volumes` | advanced volume operations (`clone`, `browse`, import/export) | partial |
| `/api/networks` | advanced network operations (`connect`, inspect details as separate surface) | partial |
| `/api/configs` | config management resource/data source | planned |
| `/api/backups` | backup resource/data source | planned |
| license-tier auth endpoints (LDAP/AD/roles) | auth enterprise resources/data sources | planned |

## Open Contract Questions

1. Exact behavior of `DELETE /api/stacks/{name}?force=true` for server error handling (observed non-2xx even when delete appears to succeed).
2. Whether create/update semantics support true in-place compose updates.
3. Whether auth should always be session-cookie based for provider use.
4. Which endpoints are stable enough for Terraform-managed desired state vs read-only telemetry.
