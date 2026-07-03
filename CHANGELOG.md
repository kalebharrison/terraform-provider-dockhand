# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
