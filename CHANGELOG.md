# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Track `GET /api/settings/navigation` in `scripts/endpoint-probe.py` so API drift gate passes on Dockhand `latest` (#252).
- Release loop no longer stalls after `GITHUB_TOKEN` agent merges: skip release-gate blockers for compatibility issues with merged fix PRs, dispatch **Agent Merge Cleanup** and **Release Drafter** via **Agent Merge Followup** after agent squash-merge, recover missed steps via **Agent Autonomy Watchdog**, and remove deprecated Release Drafter `exclude-labels` (#256).

## [0.1.89] - 2026-08-02

### Fixed

- `dockhand_git_stack`: when `webhook_enabled=true`, require `webhook_secret` or `webhook_secret_auto_generate=true`. Auto-generate now creates a provider-side secret and sends it (Dockhand no longer accepts omitting the secret).
- `dockhand_git_stack_webhook_action`: accept `webhook_secret` and sign webhook requests (`X-Hub-Signature-256` / `X-Gitlab-Token` / `?secret=`) so triggers succeed against current Dockhand.
- Allowlist `/api/registry/tag-info` in `docs/non-present-endpoints.md` so Acceptance Full API drift gate passes while provider coverage is tracked (#244).
- **Secret Smoke**: validate `CURSOR_API_KEY` via Cloud Agents `GET /v1/me` instead of the Bugbot admin API; Bugbot disable is best-effort so non-team accounts no longer open a false secret-health issue.

### Added

- _Nothing yet._

## [0.1.88] - 2026-07-21

### Fixed

- `dockhand_environment`: create with `public_ip` no longer produces an inconsistent apply when Dockhand omits or ignores `publicIp` on `POST /api/environments` (follow-up `PUT` persists the planned value).

## [0.1.87] - 2026-07-15

### Added

- Release automation tags when `main` has unreleased commits, not only `awaiting-release` issues.
- Weekly repo hygiene workflow and stale agent branch pruning.

### Changed

- **Compat Reports Sync** commits probe/drift baselines directly to `main` (zero-touch).
- **Dockhand Release Watch** skips full validation when Dockhand image digest and provider `main` SHA are unchanged.
- Agent loop routes CI failures and stuck PR approvals through automation; auto-merge polls for PR checks before merging.
- GitHub Actions dependency bumps: `actions/checkout` v7, `actions/cache` v6, `actions/download-artifact` v8, `actions/attest-build-provenance` v4.1.1.
- `golang.org/x/crypto` v0.52.0 (indirect).

### Fixed

- Release gate and compat-sync scripts stop passing `--repo` to `gh api` where unsupported.
- Release notify script compatibility with `gh release list` API change.
- Reduced `workflow_run` cascade noise and false failure emails.
- Compat Reports Sync PR branch and stall watchdog `gh api` usage.

## [0.1.86] - 2026-07-04

### Added

- `docs/AGENT_CODING_STANDARDS.md` — canonical engineering practices for agent and human contributors.
- `docs/AGENT_INTAKE.md` — how issues enter the autonomous agent loop.
- Split `internal/provider/client.go` into domain-focused `client_*.go` files for maintainability.
- `ImportState` on `dockhand_stack_env` with acceptance import/destroy coverage.
- `## Import` sections across `docs/resources/*` for import-capable resources and actions.
- Manifest acceptance tests verify declared operations in matching `TestAcc` sources.
- Expanded `scripts/endpoint-probe.py` coverage for environment subroutes, hawser tokens, git stack CRUD, git preview-env, and scanner settings.
- Three additional PR acceptance suites: registry/git credentials, container runtime surfaces, and git stack destroy.

### Changed

- Strengthened acceptance tests for `environment`, `git_stack`, `container`, `image`, and `stack_env` (import + `CheckDestroy` where applicable).
- Shrunk `manifestOperationExemptions` as coverage improved.
- `DELETE` probe calls for stacks, volumes, and images now send `force=true` to match the provider client.
- Agent docs (`AGENTS.md`, runbook, cursor rules, CONTRIBUTING) cross-link coding standards and intake guide.
- Acceptance harness git-stack bootstrap uses `postgresql-pgadmin/compose.yaml` so env-file fixtures include a committed `.env`.

### Fixed

- `force_redeploy` / `deploy_now` no longer persist as `true` in `dockhand_git_stack` state after apply.
- Inline `dockhand_batch_action` responses inspect `result` payloads for failures before reporting success.
- `dockhand_notifications` data source marks `config_json` sensitive; prune action job JSON aligned with batch sensitivity.
- `dockhand_stack.compose` and `smtp_username` marked sensitive.
- Acceptance harness bootstraps file-container, registry catalog, and git-stack fixtures; CI fails on manifest-mapped test skips.
- Endpoint probe uses POST for env/scanner mutations, DELETE scanner route, and `force=true` on container delete.
- `PushImage` uses 5-minute HTTP timeout; git deploy stream errors fail the apply.
- Action resources persist resolved `default_env` in state.
- Provider honors `DOCKHAND_INSECURE` env fallback; warns when unauthenticated mode is env-only.
- Agent workflows on `agent/**` branches cannot modify workflow files vs `main`; auto-merge requires `agent-auto-merge` label and filled issue resolution PR sections.
- Issue Resolution Notify, Release Issue Notify, and Issue Regression Intake workflows for substantive issue communication.
- GitOps example uses `dockhand_git_stack_deploy_action` with static triggers.
- Latest Dockhand compatibility: acceptance harness starts the file-container fixture and discovers git-stack env file paths after sync.
- `dockhand_git_repository` hydrates `environment_id` from the repositories list when single-repo GET omits it (import/read drift against Dockhand latest).
- Git stack deploy acceptance expects `deploy_completed` after deploy-stream finishes synchronously.

### Changed

- Nightly acceptance sets `API_DRIFT_FAIL_ON_NEW=true`.
- `verify.sh --quality` validates `acceptance-pr-ci-regex.py`.
- **Compat Reports Sync** workflow commits probe/drift baselines from CI (no maintainer machine).
- Docs realigned to CI-first model (`docs/AGENT_AUTONOMY.md`).
