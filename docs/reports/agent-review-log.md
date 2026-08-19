# Agent Review Log

Append-only record of lens sweeps. See `docs/AGENT_REVIEW_LENSES.md`.

**Next lens:** none — last release pass 2026-08-10 (v0.1.91). Re-run before next `v*` release.

---

## Issue #115 regression — environment `public_ip` on create

- **Branch:** `agent/issue-115-environment-public-ip-create`
- **Lenses:** Terraform schema & state; Acceptance & regression; GitOps / IaC practitioner; Dockhand domain / runtime
- **Started:** 2026-07-20

### 2026-07-20 — Terraform schema & state

**Scope:** `internal/provider/resource_environment.go` Create/`buildEnvironmentPayload`/`modelFromEnvironmentResponse`, `client_types.go` `environmentPayload`/`environmentResponse`.

**Summary:** Create sent `publicIp` on POST but state used the create response, which can be empty/`null` while Update (PUT) persists the value. That produced Plugin Framework inconsistent-result-after-apply for planned non-empty `public_ip`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `resource_environment.go` Create | POST response can leave `public_ip` as `""` despite plan | Follow-up PUT when planned `public_ip` ≠ create response (implemented) |
| — | `environmentPublicIPNeedsFollowUp` | Unit-tested mismatch/empty/match/null/unknown cases | Covered |

### 2026-07-20 — Acceptance & regression

**Scope:** `resource_environment_test.go`, `resource_environment_tf_acc_test.go`, `acceptance_pr_ci.json` (`TestAccEnvironmentResourceDirectDinDTerraform`).

**Summary:** Unit coverage for follow-up detection; DinD acceptance create+update now asserts `public_ip`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | prior DinD acc config | Did not set `public_ip` on create | Assert create+update values (implemented) |

### 2026-07-20 — GitOps / IaC practitioner

**Scope:** `docs/resources/environment.md`, `examples/resources/dockhand_environment/resource.tf`, `docs/api-matrix.md`, `CHANGELOG.md`.

**Summary:** Documented create-time PUT follow-up; example and docs show `public_ip`; changelog Unreleased note added.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | docs/examples | Operators need create-path caveat | Documented |

### 2026-07-20 — Dockhand domain / runtime

**Scope:** `client_environment.go` Create/Update, reporter notes (update works, create inconsistent).

**Summary:** Matches Dockhand behavior where PUT accepts `publicIp` and POST may ignore/omit it; provider create path now mirrors a successful update for that field.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | Create follow-up | Same payload as create reused for PUT | Implemented |

---

## Baseline full lens review (pre-autonomy deployment)

- **Tier:** all 11 (baseline audit)
- **Started:** 2026-07-03
- **Base commit:** `96c06b00130446f7a48f3cdd64980f8e0c09c630` (+ 37 uncommitted local files: agent CI bundle)
- **CI gates:** not re-run for this sweep (local review only)
- **Status:** complete

---

### 2026-07-03 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`, `internal/provider/client.go` (route usage vs probe list).

**Summary:** Documented 404 backlog (`GET /api/configs`, `GET /api/backups`) is consistent. Probe/drift tooling lags the client (~15 implemented routes untracked) and the March 2026 probe report has fixture false positives. No evidence that shipped provider features call absent APIs beyond the known backlog.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `scripts/endpoint-probe.py:13-148` | ~15 client-used routes not probed (env test/detect-socket, hawser tokens, git stacks CRUD, git preview-env, scanner settings, etc.) | Extend `ENDPOINTS`; re-run probe; refresh `docs/reports/endpoint-probe.md` |
| high | `scripts/api-drift-gate.py:26-47` | `RELEVANT_PREFIXES` omits `/api/settings`, `/api/license`, `/api/activity` | Extend prefixes or derive from client + api-matrix |
| high | `scripts/endpoint-probe.py:316-326` | `git_stack_id` fixture taken from `/api/stacks` not `/api/git/stacks` | Fix fixture discovery; re-probe git subroutes |
| med | `scripts/endpoint-probe.py:466-489` | Safe-mode placeholder 404s classified as `unexpected_404` | Treat as `unverified_placeholder` or skip |
| med | `scripts/endpoint-probe.py:92` | `DELETE /api/stacks/{name}` probe omits `force=true` (client always sends it) | Align probe query params with client |
| med | `docs/api-matrix.md:193-200` | Stale WebUI gap lists routes already implemented | Update gap section |
| med | `docs/reports/endpoint-probe.md` | Report dated March 2026 (~4 months old) | Re-run `./scripts/verify.sh --endpoint-probe` on current Dockhand |
| med | `internal/provider/client.go:1960-1991` | `GetJob` uses loose `map[string]any` + multi-key fallbacks | Add contract tests or typed struct when stable |
| low | `docs/non-present-endpoints.md:11-14` | Allowlist matches probe for configs/backups only | Re-verify after Dockhand upgrades |

---

### 2026-07-03 — Terraform schema & state

**Scope:** `resource_environment.go`, `resource_stack.go`, `resource_container.go`, `resource_git_stack.go`, `resource_registry.go`, `resource_auth_settings.go`, related action resources.

**Summary:** Plugin Framework patterns are generally sound (write-only registry password, webhook secret merge, action `trigger`+RequiresReplace). Blocking issues: `enabled` on stack/container does not reconcile runtime drift; `deploy_now`/`force_redeploy` on git_stack can re-fire every apply; several secret-bearing attrs lack `Sensitive`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `internal/provider/resource_stack.go:204-210` | `enabled` only applied when plan diff; external stop/start not reconciled on read | Reconcile from `status` on read or document intentional drift |
| high | `internal/provider/resource_container.go:364-391,506-514` | Same `enabled` vs runtime gap as stack | Same as stack |
| high | `internal/provider/resource_git_stack.go:168-173,410-436` | `deploy_now=true` persists in state and is sent on every update — perpetual redeploy risk | Reset to `false` after apply or use `git_stack_deploy_action` only |
| high | `internal/provider/resource_git_stack.go:186-190` | `force_redeploy` same one-shot-on-persistent-resource problem | Same as `deploy_now` |
| med | `internal/provider/resource_git_stack.go:192-197` | `env_vars_json` can hold secrets; not `Sensitive` | Mark sensitive; consider structured block |
| med | `internal/provider/resource_container.go:165-172,201-205` | `env_vars`, `update_payload_json` not sensitive | Mark `Sensitive: true` |
| med | `internal/provider/resource_container.go:481-504` | Import/read does not backfill `name`/`image` from API | Populate from `GetContainerByID` on read |
| med | `internal/provider/resource_git_stack.go:373-375` | Import ID-only; requires provider `default_env` | Support `<env>:<id>` import format |
| low | `internal/provider/resource_stack.go:52-55` | `id` Computed without `UseStateForUnknown` | Add plan modifier for stability |
| — | `internal/provider/resource_registry.go:75-79,265-272` | Write-only password + preserve pattern | Good reference for other secrets |

---

### 2026-07-03 — Dockhand domain / runtime

**Scope:** environment connection types, hawser tokens, stack vs git_stack, `default_env`, container lifecycle, registry/git credentials.

**Summary:** Agent→`hawser-edge` mapping and git_stack destroy ordering are correct. Main operator gaps: inconsistent `default_env` enforcement/persistence across resources, environment resource lacks direct/agent validation present on test action.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `internal/provider/resource_git_stack.go:235-238` vs `resource_stack.go:129` | `git_stack` errors on missing env; stack/container pass empty env silently | Shared `resolveEnvOrError()` for all env-scoped resources |
| med | `internal/provider/resource_git_stack.go:257-258` vs `resource_stack.go:129-149` | Resolved `default_env` persisted in git_stack state but not stack/container | Always persist resolved env in state |
| med | `internal/provider/resource_environment.go:377-456` | `direct`/`agent` field validation only on test action, not managed resource | Port validation into `buildEnvironmentPayload` |
| med | `internal/provider/resource_environment.go:615-631` | `agent-standard`/`hawser-standard` alias asymmetry vs `agent`↔`hawser-edge` | Document or normalize aliases |
| low | `internal/provider/resource_environment.go:269-282` | Dual hawser token paths on create; rotation may orphan tokens | Document; prefer token API path |
| low | `internal/provider/resource_git_credential.go:260-274` | Unknown `auth_type` accepted | Validate allowed types at plan time |
| — | `internal/provider/resource_git_stack.go:349-371` | Destroy deletes runtime stack then git record | Correct; keep tested |

---

### 2026-07-03 — Async & long-running operations

**Scope:** `client.go` HTTP/timeouts, `*_action.go` resources, job polling, stream readers.

**Summary:** `WaitForJob` + prune action pattern is good reference. Dominant risk: 30s global HTTP client timeout on pull/deploy/push streams; `batch_action` does not fail apply on terminal job failure (unlike prune).

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `internal/provider/client.go:961` | Global 30s `http.Client` timeout on most operations | Per-operation long-timeout clients (see `ScanImage` pattern) |
| high | `internal/provider/client.go:1824-1881` | `PullImage` NDJSON stream bounded by 30s client | Dedicated long-timeout client for pulls |
| high | `internal/provider/client.go:2761-2780` | `DeployGitStack` reads stream once, no completion wait | Stream-until-terminal or poll job; optional `wait_for_completion` |
| high | `internal/provider/resource_batch_action.go:205-218` | After `WaitForJob`, no `isFailureStatus` check (prune has it) | Fail apply on failed job status; set `success` computed |
| med | `internal/provider/client.go:1973-1995` | Default job wait 2m / batch timeout 120s may be short for large jobs | Raise defaults; surface last status on timeout |
| med | `internal/provider/resource_image_push_action.go:116-127` | Fire-and-forget after HTTP 2xx | Optional completion wait or poll |
| med | `internal/provider/resource_network_connection_action.go:195-201` | Hardcoded 20s poll; ignores `ctx` during sleep | Configurable timeout; context-aware polling |
| med | `internal/provider/resource_environment_scanner_action.go:121-137` | Chained `PullImage` calls hit 30s cap | Long-timeout pull client for scanner bootstrap |
| low | `internal/provider/resource_prune_action.go:193-215` | Correct wait + `success` from job status | Extract shared async helper for batch/deploy |

---

### 2026-07-03 — Acceptance & regression

**Scope:** `acceptance_manifest.json`, `acceptance_pr_ci.json`, manifest tests, sample `*_tf_acc_test.go`, harness env wiring.

**Summary:** Manifest↔provider parity guard (`TestAcceptanceManifestCoverage`) is strong. Gap: manifest claims coverage for git-stack/git-repo/container-file surfaces whose suites **skip** when harness does not export fixture env vars — nightly full can pass with silent skips.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `scripts/run-acceptance-harness.sh:206-229` + manifest git/container-file entries | Harness never exports `DOCKHAND_TEST_GIT_STACK_*`, `DOCKHAND_TEST_FILE_CONTAINER_ID`, `DOCKHAND_TEST_GIT_REPO_ENV_ID` | Bootstrap fixtures in harness or fail on manifest-mapped skips |
| high | `internal/provider/acceptance_manifest_test.go:169-176` | Manifest `operations` (import/delete) not enforced against test suites | Extend test or secondary operations map |
| med | `internal/provider/testdata/acceptance_pr_ci.json` | PR subset omits registry CRUD, git-stack lifecycle, container runtime bundle | Add 1–2 high-signal suites or document nightly-only deferral |
| med | `internal/provider/testdata/acceptance_manifest.json:4,42` | `dockhand_stack` shares action suite; weak create/import/update/destroy proof | Add dedicated `TestAccStackResourceTerraform` or narrow manifest ops |
| med | `resource_user_acc_test.go` vs `resource_user_tf_acc_test.go` | Manifest maps to weaker client-only test; PR CI runs TF lifecycle test | Point manifest regex at `TestAccUserResourceTerraform` |
| low | `acceptance_manifest_test.go` / PR CI workflows | `TestAcceptancePRCISuites` only in `go-ci`, not acceptance workflows | Run manifest/PR CI parity tests in acceptance jobs too |

---

### 2026-07-03 — Security engineer

**Scope:** provider auth, sensitive schema attrs, workflows, `.gitignore`, action job output fields.

**Summary:** Provider-level auth (TLS 1.2+, sensitive provider config, no `pull_request_target`) is solid. Main gap: multiple resources store credentials/secrets/log output in state without `Sensitive` marking.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `internal/provider/resource_stack_env.go:90` | `secret_variables.value` not `Sensitive` | Mark nested value sensitive |
| high | `internal/provider/resource_notification.go:94-98` | `apprise_urls` contain tokens; not sensitive | Mark sensitive |
| high | `internal/provider/resource_container_file.go:74-77` | `content` not sensitive | Mark sensitive |
| high | `internal/provider/data_source_system_file_content.go:50-51` | Host file `content` not sensitive | Mark sensitive; warn in docs |
| high | `internal/provider/resource_batch_action.go:111-112` | `result_json`/`lines_json` may contain secrets from job logs | Mark sensitive; consider redaction |
| med | `internal/provider/resource_stack_env.go:80-83` | `raw_content` can hold `.env` secrets | Mark sensitive |
| med | `internal/provider/resource_git_stack_deploy_action.go:59-60` | Deploy stream `output` not sensitive | Mark sensitive; truncate stored output |
| med | `internal/provider/provider.go:87-94` | `allow_unauthenticated` bootstrap escape hatch | Stronger doc warning |
| med | `.github/workflows/acceptance-ci.yml:26` | Throwaway CI password in workflow env (public logs) | Acceptable ephemeral; optional move to generated secret |
| med | `.github/workflows/agent-auto-merge.yml` | Auto-merge with `contents: write` | Ensure `head.repo == github.repository` guard |
| low | `internal/provider/auth.go:132` | Error bodies may echo server JSON | Truncate/sanitize diagnostics |
| — | `.github/workflows/*.yml` | No `pull_request_target` | Maintain |
| — | `.gitignore` | `.env`, `secrets/`, `*.tfvars` ignored | Maintain |

---

### 2026-07-03 — Ops / SRE

**Scope:** `agent-*.yml`, acceptance workflows, `run-acceptance-harness.sh`, `verify.sh`, `go-ci.yml`.

**Summary:** Agent loop (validate → open PR → auto-merge) is wired correctly. Ops gaps: failure artifacts omit container logs; `go-ci` duplicates `verify.sh` with drift risk; uncommitted agent CI not yet on `main`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `acceptance-ci.yml` / `agent-validate.yml` artifacts | Only `/tmp/dh-*.json` uploaded; Docker logs stdout-only | Persist `dump_logs` output to artifact directory |
| med | `.github/workflows/go-ci.yml` vs `scripts/verify.sh` | CI reimplements verify steps (shellcheck scope differs) | Call `./scripts/verify.sh --quality` as single source |
| med | `scripts/run-acceptance-harness.sh:121-125` | Dockhand mounts host docker.sock while tests target DinD env | Document in playbook why |
| med | Local tree | 37 uncommitted files (agent CI bundle) not on `main` | Commit/push; sync branch protection per `AGENT_DEPLOYMENT.md` |
| low | `agent-auto-merge.yml:38-52` | Auto-merge enable failure is warning-only | Comment on PR or fail job when enable fails |
| low | `pr-issue-link.yml:25-27` | Agent PRs skip issue lifecycle labels entirely | Replicate `in-progress`/`awaiting-release` in Agent Open PR |
| low | `agent-validate.yml` + `acceptance-ci.yml` | Acceptance runs twice per agent PR (by design) | Monitor duration; acceptable tradeoff |
| — | `agent-validate.yml` → `agent-open-pr.yml` → `agent-auto-merge.yml` | Pipeline complete | Smoke test after merge |

---

### 2026-07-03 — Senior developer

**Scope:** `client.go` structure, `request_retry.go`, `provider.go`, error handling patterns.

**Summary:** Retry policy is well isolated. `client.go` remains a ~3.3k-line god file; acceptable short-term but split by API domain when touching major areas.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| low | `internal/provider/client.go` | Monolithic client (types + all API methods) | Split into domain files when next major client work |
| low | `internal/provider/request_retry.go` | Centralized retry | Keep new retry logic here |
| — | `internal/provider/client.go` | `doRequest` body limits (64KiB err / 10MiB success) | Good defensive pattern |

---

### 2026-07-03 — GitOps / IaC practitioner

**Scope:** `docs/index.md`, `docs/resources/`, `examples/scenarios/`, per-resource examples.

**Summary:** Per-resource doc↔example parity (42/42) is strong. Two scenario examples use invalid attribute names; import documented for only ~7 resources despite ~34 import-capable; no central action-vs-resource guide.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `examples/scenarios/gitops-stack/main.tf:32-48` | Invalid attrs: `name`, `environment_id`, `auto_deploy`; env file treated as writable | Rewrite to match real schema (`stack_name`, `env`, `deploy_now`; env file is read/sync) |
| high | `examples/scenarios/registry-and-image/main.tf` | Invalid `pull`, `registry_id`, `image_id` attrs | Align with `dockhand_image` + scan action schema |
| med | `docs/resources/*.md` | Import sections for ~7 resources; many more support import in Go | Import doc sweep + central guide in `docs/index.md` |
| med | `docs/index.md` | No resource vs action vs data source decision table | Add “Choosing the right resource” section |
| med | `docs/index.md` | `default_env` in schema but weak multi-env narrative | Add worked multi-env example |
| med | `docs/resources/git_stack_env_file.md` | Read/sync resource; scenario implies write | Clarify; point writes to `env_vars_json` |
| low | `examples/scenarios/` | No per-scenario `terraform.tfvars.example` | Add for copy-paste onboarding |

---

### 2026-07-03 — Entry-level developer

**Scope:** `README.md`, `CONTRIBUTING.md`, `AGENTS.md`, `docs/LOCAL_DEV.md`, agent docs.

**Summary:** Material exists but is fragmented across 6+ files. No single first-time path; jargon (Hawser, DinD, manifest, harness) undefined for contributors.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CONTRIBUTING.md` | No sequenced clone → Go → `./scripts/verify.sh --quality` | Add “First-time setup” section |
| med | Onboarding docs | Hawser, DinD, acceptance harness, manifest undefined | Glossary in CONTRIBUTING or README |
| med | `docs/LOCAL_DEV.md` | Thin; no Docker/acceptance prerequisites | Expand with harness pointer |
| med | Agent docs (6 files) | Human vs `agent/**` workflow repeated | Single “Contributor paths” summary with links |
| med | `CONTRIBUTING.md` vs playbook | Human branch `codex/*` only in playbook | Document both conventions in CONTRIBUTING |
| med | `AGENTS.md` | Long resource inventory overwhelms onboarding | Trim; link to `docs/index.md` |
| low | `CONTRIBUTING.md` | Thin CI failure guidance vs `AGENT_RUNBOOK` | Add artifact/`gh pr checks` hints |
| low | `docs/AGENT_DEPLOYMENT.md` | Maintainer-only steps not bannered | Add audience label at top |
| — | `CONTRIBUTING.md` | SECURITY.md linked | Good |

---

### 2026-07-03 — Release & upgrade

**Scope:** `README.md`, `docs/testing/release-gate.md`, release workflow, version pins, local uncommitted state.

**Summary:** Release-first workflow is documented. No `CHANGELOG` file. README pins `>= 0.1.63`. Large uncommitted agent CI bundle must land before first post-agent release. Endpoint probe report stale.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | Repo root | No `CHANGELOG.md` | Add changelog discipline before next tag |
| med | `docs/reports/endpoint-probe.md` | Stale (March 2026) | Re-run probe before release |
| med | Working tree | Agent CI/autonomy work uncommitted | Merge agent bundle; smoke test; sync branch protection |
| low | `README.md:23` | `version = ">= 0.1.63"` | Bump constraint when tagging |
| low | `docs/testing/release-gate.md` | Lens review step present | Run tiered lenses before tag per tier |
| — | `.github/workflows/release-artifacts.yml` | Signed tag → artifacts pipeline exists | Use release-first validation |

---

### Baseline full sweep — verdict

- **Clear to tag a release:** **no** — unresolved **high** product findings (git_stack redeploy loop, batch job success semantics, enabled drift, sensitive schema gaps, broken scenario examples, acceptance silent skips).
- **Clear to proceed with autonomy deployment (5-step plan):** **yes, with parallel fix backlog** — agent CI wiring is sound; product highs should be filed as issues and fixed via `agent/issue-*` branches, not block landing the automation layer.
- **Blocking findings (product):**
  - `deploy_now` / `force_redeploy` perpetual apply on `dockhand_git_stack`
  - `dockhand_batch_action` succeeds when job failed
  - Stack/container `enabled` does not reconcile runtime
  - Manifest-covered acceptance suites skip without harness fixtures
  - Broken `examples/scenarios/gitops-stack` and `registry-and-image`
- **Blocking findings (tooling/docs):**
  - Endpoint probe coverage drift vs client
  - Import docs / action-vs-resource guide gaps
- **Deferred medium/low:** track via GitHub issues; see per-lens tables above.

**Recommended issue themes for agent backlog (priority order):**

1. Git stack deploy flags one-shot semantics — **fixed locally**
2. Batch action job failure detection + shared async helper — **fixed locally**
3. Harness git-stack/container-file fixtures (stop silent skips) — **fixed locally**
4. Sensitive schema sweep (`stack_env`, notifications, container file, job JSON) — **fixed locally**
5. Fix scenario examples + add `docs/index.md` resource chooser — **fixed locally**
6. Endpoint probe + drift gate alignment with client — **addressed in deferred pass (probe routes + drift prefixes)**
7. `default_env` persistence consistency across resources — **partially fixed** (`requireResolvedEnv` / `persistEnvAttr`)

---

### 2026-07-03 — Deferred lens backlog (post-baseline fixes)

**Scope:** client split, endpoint probe, import docs, manifest operation enforcement, PR CI suites, CHANGELOG.

**Status:** complete (local; not yet released)

| Item | Result |
|------|--------|
| Split `client.go` by API domain | `client_types.go` + `client_{settings,registry,git,config,environment,network,volume,image,schedule,stack,container,system}.go` |
| Endpoint probe expansion | +15 routes; `force=true` on destructive deletes; `query` merge support |
| Import doc sweep | `## Import` on all import-capable resource docs (43 files) |
| Manifest operations enforcement | `validateManifestOperationsInTests` with documented exemptions for known gaps |
| PR CI expansion | +`TestAccRegistryAndGitCredentialResourcesTerraform`, `TestAccContainerRuntimeSurfacesTerraform`, `TestAccGitStackResourceDestroyRemovesRuntimeTerraform` |
| CHANGELOG | Added `CHANGELOG.md` with Unreleased section |

**Remaining medium/low (track as issues):**

- Re-run `./scripts/verify.sh --endpoint-probe` against live Dockhand and refresh `docs/reports/endpoint-probe.md` counts
- Shrink remaining `manifestOperationExemptions` (action delete checks, schedule delete, etc.)
- Agent autonomy deployment (commit CI bundle, branch protection, smoke test)

### 2026-07-03 — Agent readiness pass (coding standards + test hardening)

**Scope:** `AGENT_CODING_STANDARDS.md`, `AGENT_INTAKE.md`, acceptance import/destroy coverage, `stack_env` ImportState, cursor rules, doc cross-links.

**Status:** complete (local)

| Item | Result |
|------|--------|
| Agent coding standards | `docs/AGENT_CODING_STANDARDS.md` + `.cursor/rules/agent-coding-standards.mdc` |
| Issue intake guide | `docs/AGENT_INTAKE.md` |
| `dockhand_stack_env` import | `ImportState` + acceptance import/`CheckDestroy` |
| Acceptance hardening | import/destroy on environment, git_stack, container, image suites |
| Manifest exemptions | Removed git_stack, environment, container, image, stack_env gaps |
| Docs | AGENTS, RUNBOOK, CONTRIBUTING, SWEEP, DEPLOYMENT, CHANGELOG updated |

---

### 2026-07-03 — Senior developer (historical baseline)

**Scope:** `client.go` (structure, `doRequest`/retry), `request_retry.go`, `provider.go` (sample), manifest test patterns.

**Summary:** Core HTTP/retry design is solid and centralized. Main maintainability risk is `client.go` size (~3.3k lines); acceptable for now but worth splitting by API domain if it keeps growing.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| low | `internal/provider/client.go` | Single large client file holds all API types + methods | Track; split when touching major areas |
| low | `internal/provider/request_retry.go` | Retry policy well isolated from resources | Keep new retry behavior here, not in resources |
| — | `internal/provider/client.go` | `doRequest` limits error body to 64KiB, success to 10MiB | Good pattern; no action |

---

## Post-deferred full lens review (all 11)

- **Tier:** all 11 (minor/major depth — large uncommitted tree)
- **Started:** 2026-07-03
- **Base commit:** `96c06b00130446f7a48f3cdd64980f8e0c09c630` (+ ~115 uncommitted local files)
- **CI gates:** `./scripts/verify.sh --quality` pass (local); GitHub Actions not re-run
- **Status:** complete

**Delta since baseline:** Client split, import doc sweep, manifest operation enforcement, `AGENT_CODING_STANDARDS.md`, acceptance import/destroy hardening, endpoint probe expansion. Several baseline highs are **fixed**; findings below reflect current tree.

---

### 2026-07-03 — API compatibility (re-run)

**Scope:** `scripts/endpoint-probe.py` (~154 entries), `client_*.go`, `api-drift-gate.py`, `docs/api-matrix.md`, `docs/reports/endpoint-probe.md`.

**Summary:** Client routes are represented in probe paths. Remaining gaps are method/query accuracy and a stale live report.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `endpoint-probe.py:59-66` vs `client_environment.go:122,135,148` | Probe uses PUT; client uses POST for env timezone/update-check/image-prune and scanner settings | Change probe methods to POST |
| med | `client_environment.go:197` vs `endpoint-probe.py:65-66` | Client DELETE `/api/settings/scanner`; probe only GET+PUT | Add DELETE scanner probe |
| med | `endpoint-probe.py:145` vs `client_container.go:274-279` | Container DELETE omits `force=true` | Add query `force=true` |
| med | `docs/reports/endpoint-probe.md` | Report stale vs current ENDPOINTS | Re-run endpoint probe; refresh report |
| low | `api-drift-gate.py` | Path-only matching; missing configs/backups prefixes | Extend gate |
| — | stack/volume/image DELETE `force=true` | Aligned with client | Fixed since baseline |

---

### 2026-07-03 — Terraform schema & state (re-run)

**Summary:** `deploy_now` fixed; `force_redeploy` can still drift; inline batch can false-success.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `resource_git_stack.go:592-696` | `force_redeploy` from API not reset like `deploy_now`; HCL `false` does not win over API `true` | Reset `force_redeploy` to false on read; prefer HCL in merge |
| high | `resource_batch_action.go:190-225` | Inline batch (no job_id) sets success from status only | Parse `submitted.Result` for failures |
| med | `client_schedule_job.go` | Terminal status variants may cause WaitForJob timeout | Align with extractBatchStatus |
| med | `resource_stack.go`, `resource_container.go` | enabled unchanged for transitional statuses | Map more statuses or skip with diagnostic |
| med | `resource_git_stack.go` | env_vars_json write-only | Document or masked read |
| med | `resource_stack.go` | compose not Sensitive | Mark sensitive or document |
| — | stack_env ImportState, deploy_now one-shot, enabled reconciliation | Improved | Fixed since baseline |

---

### 2026-07-03 — Dockhand domain / runtime (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | action resources | Resolved default_env not persisted in state | Persist env on create like resource_image |
| med | `resource_git_repository.go` | environment_id vs default_env confusion | Document; consider defaulting |
| — | runtime_helpers.go usage | Partial env validation/persistence | Improved since baseline |

---

### 2026-07-03 — Async & long-running operations (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `client_image.go` PushImage | 30s HTTP client vs 5m for pull/deploy | Use httpClientWithTimeout(5m) |
| med | `client_git.go` deploy stream | Stream not parsed for failures | Parse stream; fail apply on error |
| med | image_scan_action | Apply completes on submit not completion | Document async semantics |
| — | pull/deploy 5m timeouts | Present | Fixed since baseline |

---

### 2026-07-03 — Acceptance & regression (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `run-acceptance-harness.sh:202-206` | FILE_CONTAINER_ID often empty before tests | Bootstrap container in harness |
| high | harness | GIT_STACK_ID / ENV_PATH / REGISTRY_CATALOG_ID never exported | Bootstrap fixtures; stop silent skips |
| high | CI workflows | No t.Skip detection for manifest-mapped tests | Parse go test -json; fail on skips |
| med | manifestOperationExemptions | 22 entries remain | Shrink over time |
| med | acceptance_pr_ci.json | 12 suites vs 85 manifest entries | Document; rotate PR suites |
| — | import/CheckDestroy hardening, exemption shrink | Improved | Fixed since baseline |

---

### 2026-07-03 — Security engineer (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `data_source_notifications.go` | config_json not Sensitive | Mark Sensitive; redact secrets |
| high | probe/release-watch artifacts | response_preview may leak tokens | Scrub artifacts |
| med | acceptance failure artifacts | dh-*.json may include session cookies | Scrub on upload |
| med | agent-validate.yml | Workflow from agent branch commit | Pin to main workflow ref |
| med | agent-auto-merge.yml | Auto-merge without human review | Optional required review |
| — | resource secret schema sweep | Mostly complete | Fixed since baseline |

---

### 2026-07-03 — Ops / SRE (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | workflows | :latest image pins | Pin digests/tags |
| med | acceptance-full | API_DRIFT_FAIL_ON_NEW=false | Enable when stable |
| med | PR CI | No endpoint probe | Probe subset on PR or gate nightly |
| low | verify.sh | Missing acceptance-pr-ci-regex check | Add to core gate |
| — | Agent CI workflow chain | Ready to land | Documented in AGENT_SWEEP |

---

### 2026-07-03 — Senior developer (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | runtime_helpers.go | persistEnvAttr under-used | Standardize across env-scoped resources |
| low | client_types.go | Monolithic types file | Split when touching |
| — | client.go domain split | Done | Fixed since baseline |

---

### 2026-07-03 — GitOps / IaC practitioner (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | gitops-stack example | deploy_now vs deploy_action guidance | Rewrite to deploy_action |
| med | gitops-stack example | timestamp() trigger | Use static triggers |
| low | docs/index.md | Incomplete resource chooser | Extend table |
| — | Import docs, registry-and-image scenario | Complete | Fixed since baseline |

---

### 2026-07-03 — Entry-level developer (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| low | CONTRIBUTING.md | No CI troubleshooting section | Add failure playbook |
| low | AGENTS.md | Missing DEPLOYMENT/SWEEP links | Cross-link |
| — | AGENT_CODING_STANDARDS, AGENT_INTAKE | Added | Fixed since baseline |

---

### 2026-07-03 — Release & upgrade (re-run)

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | CHANGELOG.md | Unreleased only | Backfill or cut release |
| med | AGENT_SWEEP.md | Agent CI not merged to main | Complete deployment checklist |
| low | version pins in README/examples | No migration notes | Align with CHANGELOG |

---

### Post-deferred full sweep — verdict

- **Clear to tag a release:** **no** — highs: force_redeploy drift, inline batch false success, harness skips, notifications config_json sensitivity, probe method gaps.
- **Clear to proceed with agent autonomy deployment:** **yes, with parallel fix backlog** — land agent CI/docs; file product issues on agent branches.
- **Fixed since baseline:** deploy_now one-shot, batch failure when job_id present, enabled reconciliation, sensitive schema (most), examples, client split, import docs, manifest enforcement, drift prefixes (partial).

**Recommended agent backlog (priority):**

1. ~~force_redeploy one-shot semantics~~ — fixed
2. ~~Inline batch result failure parsing~~ — fixed
3. ~~Harness fixtures + CI skip detection~~ — fixed
4. ~~notifications config_json Sensitive~~ — fixed
5. ~~Endpoint probe method/query refresh~~ — fixed (re-run live probe before release)
6. ~~GitOps scenario rewrite~~ — fixed
7. ~~PushImage timeout + deploy stream parsing~~ — fixed
8. ~~Persist env on action resources~~ — fixed
9. ~~Agent CI trust boundary hardening~~ — fixed (`agent-auto-merge` label + workflow integrity check)

### 2026-07-03 — Lens fix pass — verdict

- **Clear to tag a release:** **yes** after agent CI merge to `main` and green Acceptance Full / Release Watch (compat baselines via **Compat Reports Sync**)
- **Clear to proceed with agent autonomy deployment:** **yes** — land CI bundle, sync branch protection, add `agent-auto-merge` label in GitHub, smoke test
- **`./scripts/verify.sh --quality`:** passing

---

## Release v0.1.85 — lens review

- **Tier:** patch
- **Started:** 2026-07-04
- **Base commit:** `7ff227f7bced3ef4470aedde7a300f0b045deb3b`
- **CI gates:** pass (`scripts/release_gate_check.py`)
- **Awaiting-release issues:** #135, #132, #121, #99, #92, #66, #60, #50, #36, #22
- **Status:** clear to tag

---

### 2026-07-04 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `internal/provider/client_*.go`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`.

**Summary:** Probe list (131 routes) aligns with the split client modules. Only documented backlog routes (`GET /api/configs`, `GET /api/backups`) are absent on the tested Dockhand instance. Six “unexpected 404” probe rows are fixture/safe-mode artifacts (e.g. deploy-stream, env-files) — acceptance and Release Watch exercise these paths successfully on current Dockhand.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `docs/reports/endpoint-probe.md:30-36` | Parameterized routes (deploy-stream, env-files, env/user DELETE) show unexpected 404 in static report | Refresh via **Compat Reports Sync** after tag; tune probe fixtures if 404 persists on live Dockhand |
| med | `docs/non-present-endpoints.md:7-9` | Last-verified date still March 2026 | **Compat Reports Sync** updates after green Release Watch |
| low | `docs/api-matrix.md` | Residual WebUI gap notes may lag newest resources | Periodic doc sweep on minor releases |
| — | `scripts/endpoint-probe.py` | POST env/scanner mutations, DELETE scanner, `force=true` on destructive deletes | Fixed since v0.1.84 backlog |
| — | `scripts/api-drift-gate.py` | Prefixes cover settings, license, activity | Fixed since baseline |

---

### 2026-07-04 — Terraform schema & state

**Scope:** `resource_git_stack.go`, `resource_stack.go`, `resource_container.go`, `resource_batch_action.go`, `resource_git_repository.go`, `runtime_helpers.go` (env resolution/persistence).

**Summary:** Prior release-blocking schema issues are resolved on `main`. One-shot deploy flags reset after apply; runtime `enabled` reconciles from stack/container status; inline batch failures surface via `jobPayloadIndicatesFailure`; action resources persist resolved `default_env`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `internal/provider/resource_git_stack.go:192-197` | `env_vars_json` can hold secrets; marked Sensitive but write-only on read | Document write-only semantics in resource doc (already partially noted) |
| med | `internal/provider/resource_stack.go`, `resource_container.go` | `enabled` reconciliation skips transitional statuses | Extend status map or emit diagnostic on unknown status |
| low | `internal/provider/resource_stack.go:52-55` | Stack `id` Computed without `UseStateForUnknown` | Add plan modifier when touching stack resource |
| — | `resource_git_stack.go:592,646,706-711` | `deploy_now` / `force_redeploy` reset to `false` on read; HCL wins in merge | Fixed (#121 area) |
| — | `resource_batch_action.go:199-237` | Inline + polled batch inspect result payload and fail on terminal failure | Fixed |
| — | `resource_stack.go:197-198`, `resource_container.go` | `enabled` reconciles from runtime status on read | Fixed |

---

### 2026-07-04 — Acceptance & regression

**Scope:** `scripts/run-acceptance-harness.sh`, `scripts/check-acceptance-skips.py`, `acceptance_manifest.json`, `acceptance_pr_ci.json`, `acceptance_manifest_test.go`, sample `*_tf_acc_test.go` (git stack, container runtime, registry, image import).

**Summary:** Harness bootstraps file-container, git-stack, and git-repo fixtures; CI fails manifest-mapped skips via `check-acceptance-skips.py`. PR CI runs 13 targeted suites including registry/git credentials, container runtime, git stack destroy, and image import. Latest Dockhand compatibility stabilized in #136.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `acceptance_manifest_test.go:34-60` | 22 `manifestOperationExemptions` remain (mostly action delete checks) | Shrink incrementally on future agent branches |
| med | `acceptance_pr_ci.json` | 13 suites vs ~85 manifest entries — nightly-only coverage by design | Document in CONTRIBUTING; rotate PR suites on minor releases |
| low | `examples/scenarios/registry-and-image/main.tf:34` | `timestamp()` trigger causes perpetual diff | Replace with static trigger string in follow-up |
| — | `run-acceptance-harness.sh:231-309` | Exports `DOCKHAND_TEST_FILE_CONTAINER_ID`, git-stack/repo fixtures | Fixed |
| — | `check-acceptance-skips.py` | Fails CI when manifest-mapped tests skip | Fixed |
| — | `resource_git_stack_tf_acc_test.go`, `resource_git_repository_tf_acc_test.go` | Latest Dockhand drift/import coverage (#136) | Fixed |

---

### 2026-07-04 — Security engineer

**Scope:** `provider.go`, `auth.go`, `client.go`, secret-bearing resources (`stack_env`, `notification`, `container_file`, `batch_action`, `git_*`), data sources (`notifications`, `system_file_content`), `.github/workflows/*`, `.gitignore`.

**Summary:** Provider auth defaults are safe (TLS 1.2+, sensitive provider config, `allow_unauthenticated` warning). Secret-bearing schema attributes are marked `Sensitive` across resources and relevant data sources. No `pull_request_target` workflows; agent branches cannot modify workflow files vs `main`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `.github/workflows/acceptance-ci.yml` | Throwaway CI password in workflow env (public repo) | Acceptable ephemeral creds; optional generated secret |
| med | `internal/provider/auth.go:132` | Error bodies may echo server JSON | Truncate/sanitize diagnostics in future hardening |
| low | CI failure artifacts | Session cookies possible in `dh-*.json` uploads | Scrub artifacts in follow-up |
| — | `data_source_notifications.go:65`, `resource_batch_action.go:116-117`, `resource_stack_env.go:83,93` | Sensitive marking on config/job/env secrets | Fixed since baseline |
| — | `.github/workflows/*.yml` | No `pull_request_target`; agent workflow integrity checks | Maintain |

---

### 2026-07-04 — Release & upgrade

**Scope:** `CHANGELOG.md`, `README.md`, `docs/testing/release-gate.md`, `scripts/release_gate_check.py`, awaiting-release queue (#135, #132, #121, #99, #92, #66, #60, #50, #36, #22), `.github/workflows/agent-release-tag.yml`.

**Summary:** `CHANGELOG.md` Unreleased section documents all fixes shipping in v0.1.85. Release gate script reports `ci_gates_pass: true` with patch tier and ten awaiting-release issues. Examples scenarios (`gitops-stack`, `registry-and-image`) align with current schema. `./scripts/verify.sh --quality` passes locally.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CHANGELOG.md:8-51` | Unreleased section not yet cut to `[0.1.85]` heading | **Release Artifacts** / post-tag housekeeping via Compat Reports Sync |
| low | `README.md:23` | `version = ">= 0.1.63"` constraint | Bump to `>= 0.1.85` after tag (optional doc follow-up) |
| low | `examples/scenarios/registry-and-image/main.tf:34` | `timestamp()` trigger | Static trigger in follow-up example PR |
| — | `scripts/release_gate_check.py` | Gates green; `tier: patch`; `lens_verdict_clear: false` until this log | Satisfied by this review |
| — | Agent automation (#132, release orchestrate/tag) | Hands-off release loop wired on `main` | Ship with v0.1.85 |

---

### Release v0.1.85 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** probe fixture 404 tuning and report refresh (**Compat Reports Sync** post-tag); shrink `manifestOperationExemptions`; registry-and-image `timestamp()` example; README version pin bump; optional artifact scrubbing — track on future `agent/**` branches (no new high-severity regressions)

---

## Release v0.1.86 — lens review

- **Tier:** patch
- **Started:** 2026-07-04
- **Base commit:** `cd3f854`
- **CI gates:** pass (`scripts/release_gate_check.py`)
- **Awaiting-release issues:** #121, #99, #92, #66, #60, #50, #36, #22
- **Status:** clear to tag

---

### 2026-07-04 — Acceptance & regression

**Scope:** Release Watch cache/skip (`scripts/release_watch_state.py`, `.github/workflows/dockhand-release-watch.yml`), acceptance harness bootstrap, `dockhand_git_repository` import hydration, optional `dockhand_job` coverage, API drift gate.

**Summary:** Acceptance Full and Dockhand Release Watch are green on `main`. Release Watch persists digest/tag/provider-SHA in the `dockhand-release-watch-state` Actions cache and skips revalidation when the same Dockhand image and provider SHA were recently tested. Harness fixes stabilize git-stack/env-file and git-repository import paths on latest Dockhand.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| low | `resource_new_surfaces_tf_acc_test.go` | `dockhand_job` skipped when batch completes inline (no async `job_id`) | Optional coverage only; revisit if Dockhand exposes stable async job ids for small batches |
| — | `release_watch_state.py`, Release Watch workflow | Cache seed/migrate, stale-only schedule, provider SHA tracking, skip path verified | Fixed |
| — | `run-acceptance-harness.sh` | DinD bootstrap + git deploy-stream/env-files polling | Fixed |
| — | `resource_git_repository.go` | `environment_id` import/read hydration | Fixed (#136 area) |
| — | `endpoint-probe.py` | Track `/api/settings/scanner/cache` for WebUI drift | Fixed |

---

### 2026-07-04 — Release & upgrade

**Scope:** `CHANGELOG.md` Unreleased, release automation, awaiting-release queue, Compat Reports Sync workflow.

**Summary:** Unreleased changelog captures harness, import, cache, and drift fixes shipping in v0.1.86. Release gate reports green Acceptance Full / Release Watch with patch tier and eight awaiting-release issues. Compat Reports Sync checkout order fixed so probe baselines can commit post-run.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CHANGELOG.md` | Unreleased section not yet cut to `[0.1.86]` heading | Post-tag housekeeping via release automation |
| — | `compat-reports-sync.yml` | Artifact download after checkout | Fixed |
| — | Release Watch `persist_state` | Checkout before writing cache | Fixed |

---

### Release v0.1.86 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** optional async `dockhand_job` acceptance when API stable; CHANGELOG cut heading post-tag; Compat Reports Sync PR merge after next green run

---

## Release v0.1.87 — lens review

- **Tier:** patch
- **Started:** 2026-07-13
- **Base commit:** `800bd5b`
- **CI gates:** pass (`scripts/release_gate_check.py`)
- **Awaiting-release issues:** none
- **Unreleased commits:** 33 (CI/automation, compat baselines, dependency bumps)
- **Status:** clear to tag

---

### 2026-07-13 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`, `docs/reports/api-drift-gate.md`, `docs/reports/dockhand-last-tested.json`, `internal/provider/client_*.go` (grep for changes since `v0.1.86`).

**Summary:** No provider client or probe-list changes since `v0.1.86`. Compat Reports Sync refreshed baselines on 2026-07-13; Release Watch validated Dockhand `latest` (`sha256:871700eb…`). Probe tracks 156 routes with only the documented backlog absent (`GET /api/configs`, `GET /api/backups`). API drift gate reports zero new relevant endpoints.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `docs/reports/endpoint-probe.md:23-37` | 13 parameterized routes show unexpected 404 in static probe (users, environments, git stacks, containers) — fixture/safe-mode artifacts; acceptance and Release Watch exercise these paths on live Dockhand | Continue tuning probe fixtures on future agent branches; no release block |
| low | `docs/api-matrix.md` | Residual WebUI gap notes may lag newest resources | Periodic doc sweep on minor releases |
| — | `docs/non-present-endpoints.md:7-14` | Last-verified July 13, 2026; backlog matches probe | Current |
| — | `scripts/api-drift-gate.md:10` | `New relevant endpoints not allowlisted: 0` | Current |
| — | `internal/provider/client_*.go` | No edits since `v0.1.86` | Skim confirmed |

---

### 2026-07-13 — Terraform schema & state

**Scope:** `internal/provider/provider.go`, `resource_git_stack.go`, `resource_stack.go`, `resource_container.go`, `resource_batch_action.go`, `runtime_helpers.go` (grep for changes since `v0.1.86`).

**Summary:** No provider schema or state logic changed in the v0.1.87 window. Prior release fixes remain intact: one-shot `deploy_now`/`force_redeploy` reset, `enabled` reconciliation, inline batch failure detection, resolved `default_env` on action resources.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `internal/provider/resource_git_stack.go:192-197` | `env_vars_json` can hold secrets; marked Sensitive but write-only on read | Document write-only semantics in resource doc (deferred from v0.1.86) |
| med | `internal/provider/resource_stack.go`, `resource_container.go` | `enabled` reconciliation skips transitional statuses | Extend status map or emit diagnostic on unknown status (deferred) |
| low | `internal/provider/resource_stack.go:52-55` | Stack `id` Computed without `UseStateForUnknown` | Add plan modifier when touching stack resource (deferred) |
| — | `internal/provider/*.go` | No edits since `v0.1.86` | Skim confirmed; no new regressions |

---

### 2026-07-13 — Acceptance & regression

**Scope:** `acceptance_manifest.json`, `acceptance_pr_ci.json`, `acceptance_manifest_test.go`, `scripts/run-acceptance-harness.sh`, `scripts/check-acceptance-skips.py`, `.github/workflows/acceptance-ci.yml`, `.github/workflows/dockhand-release-watch.yml`, `.github/workflows/compat-reports-sync.yml`, CI automation commits since `v0.1.86`.

**Summary:** Acceptance Full and Dockhand Release Watch are green on `main` per `release_gate_check.py`. Manifest and PR CI subset (13 suites) unchanged. This release window improves CI reliability: Release Watch skips revalidation when Dockhand digest and provider SHA are unchanged; Compat Reports Sync commits baselines directly to `main`; auto-merge polls for checks; agent loop routes CI failures and stuck PRs.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `acceptance_manifest_test.go:34-60` | 22 `manifestOperationExemptions` remain (mostly action delete checks) | Shrink incrementally on future agent branches (deferred) |
| med | `acceptance_pr_ci.json` | 13 suites vs ~85 manifest entries — nightly-only coverage by design | Document in CONTRIBUTING; rotate PR suites on minor releases (deferred) |
| low | `resource_new_surfaces_tf_acc_test.go` | `dockhand_job` skipped when batch completes inline | Optional coverage when API stable (deferred from v0.1.86) |
| low | `examples/scenarios/registry-and-image/main.tf:34` | `timestamp()` trigger causes perpetual diff | Replace with static trigger in follow-up (deferred) |
| — | `dockhand-release-watch.yml`, `release_watch_state.py` | Skip full validation when image digest + provider SHA unchanged | Fixed (#167 area) |
| — | `compat-reports-sync.yml` | Direct commit to `main`; zero-touch baseline refresh | Fixed |
| — | `agent-pr-approve-ci.yml` | Poll for PR checks before auto-merge | Fixed (#162) |

---

### 2026-07-13 — Security engineer

**Scope:** `internal/provider/provider.go`, `internal/provider/auth.go`, `internal/provider/client.go`, secret-bearing resources (sample grep), `.github/workflows/*`, `go.mod` (`golang.org/x/crypto` bump), `.gitignore`.

**Summary:** No provider auth or secret-handling code changed since `v0.1.86`. `golang.org/x/crypto` bumped to v0.52.0 (#165); GitHub Actions bumped (checkout v7, cache v6, download-artifact v8, attest-build-provenance v4.1.1). No `pull_request_target` workflows. Agent branches still cannot modify workflow files vs `main`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `.github/workflows/acceptance-ci.yml` | Throwaway CI password in workflow env (public repo) | Acceptable ephemeral creds; optional generated secret (deferred) |
| med | `internal/provider/auth.go:132` | Error bodies may echo server JSON | Truncate/sanitize diagnostics in future hardening (deferred) |
| low | CI failure artifacts | Session cookies possible in `dh-*.json` uploads | Scrub artifacts in follow-up (deferred) |
| — | `go.mod` | `golang.org/x/crypto v0.52.0` indirect dep bump | Merged via Dependabot #165 |
| — | `.github/workflows/*.yml` | No `pull_request_target`; workflow integrity checks maintained | Current |

---

### 2026-07-13 — Release & upgrade

**Scope:** `CHANGELOG.md`, `README.md`, `docs/testing/release-gate.md`, `scripts/release_gate_check.py`, `scripts/release_housekeeping.py`, `.github/workflows/agent-release-tag.yml`, unreleased commit log since `v0.1.86`.

**Summary:** Patch-tier maintenance release: 33 commits since `v0.1.86` with no provider schema or client changes. Release automation now tags when `main` has unreleased commits (not only `awaiting-release` issues). `release_gate_check.py` reports `ci_gates_pass: true`, `tier: patch`, zero blockers. `./scripts/verify.sh --quality` passes locally. `CHANGELOG.md` Unreleased documents CI/automation improvements shipping in v0.1.87.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CHANGELOG.md` | Unreleased section not yet cut to `[0.1.87]` heading | Post-tag housekeeping via **Agent Release Tag** / `release_housekeeping.py` |
| low | `README.md:23` | `version = ">= 0.1.63"` constraint | Bump to `>= 0.1.87` after tag (optional doc follow-up) |
| — | `scripts/release_gate_check.py` | Gates green; `ready_for_lens_dispatch: true` | Satisfied |
| — | `agent-release-tag.yml`, #167 | Release when main has unreleased commits | Fixed |
| — | `scripts/release_notify.py` (861906b) | `gh release list` API compatibility | Fixed |

---

### Release v0.1.87 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** probe fixture 404 tuning; shrink `manifestOperationExemptions`; registry-and-image `timestamp()` example; README version pin bump; optional artifact scrubbing and auth error sanitization — carry forward from v0.1.85/v0.1.86 (no new high-severity regressions)

---

## Release v0.1.88 — lens review

- **Tier:** patch
- **Started:** 2026-07-20
- **Base commit:** `49fe8fc`
- **CI gates:** pass (`scripts/release_gate_check.py`)
- **Awaiting-release issues:** none
- **Unreleased commits:** 24 (CI/automation, compat baselines, dependency bumps)
- **Status:** clear to tag

---

### 2026-07-20 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`, `docs/reports/api-drift-gate.md`, `docs/reports/dockhand-last-tested.json`, `docs/reports/docs-reference-api-endpoints.txt`, `internal/provider/client_*.go` (grep for changes since `v0.1.87`).

**Summary:** No provider client or probe-list changes since `v0.1.87`. Compat Reports Sync refreshed baselines on 2026-07-20; Release Watch validated Dockhand `latest` (`sha256:871700eb…`). Probe tracks 156 routes with only the documented backlog absent (`GET /api/configs`, `GET /api/backups`). API drift gate reports zero new relevant endpoints. Docs-reference drift audit discovered backup API routes in Dockhand manual (not yet in probe list) — informational only.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `docs/reports/endpoint-probe.md:23-37` | 13 parameterized routes show unexpected 404 in static probe (users, environments, git stacks, containers) — fixture/safe-mode artifacts; acceptance and Release Watch exercise these paths on live Dockhand | Continue tuning probe fixtures on future agent branches; no release block |
| med | `docs/reports/docs-reference-api-endpoints.txt` | Docs drift audit lists `/api/backup/*` routes not yet in probe or provider | Track for future provider expansion; no release block |
| low | `docs/api-matrix.md` | Residual WebUI gap notes may lag newest resources | Periodic doc sweep on minor releases |
| — | `docs/non-present-endpoints.md:7-14` | Last-verified July 20, 2026; backlog matches probe | Current |
| — | `docs/reports/api-drift-gate.md:10` | `New relevant endpoints not allowlisted: 0` | Current |
| — | `internal/provider/client_*.go` | No edits since `v0.1.87` | Skim confirmed |

---

### 2026-07-20 — Terraform schema & state

**Scope:** `internal/provider/provider.go`, `resource_git_stack.go`, `resource_stack.go`, `resource_container.go`, `resource_batch_action.go`, `runtime_helpers.go` (grep for changes since `v0.1.87`).

**Summary:** No provider schema or state logic changed in the v0.1.88 window. Prior release fixes remain intact: one-shot `deploy_now`/`force_redeploy` reset, `enabled` reconciliation, inline batch failure detection, resolved `default_env` on action resources.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `internal/provider/resource_git_stack.go:192-197` | `env_vars_json` can hold secrets; marked Sensitive but write-only on read | Document write-only semantics in resource doc (deferred from v0.1.86) |
| med | `internal/provider/resource_stack.go`, `resource_container.go` | `enabled` reconciliation skips transitional statuses | Extend status map or emit diagnostic on unknown status (deferred) |
| low | `internal/provider/resource_stack.go:52-55` | Stack `id` Computed without `UseStateForUnknown` | Add plan modifier when touching stack resource (deferred) |
| — | `internal/provider/*.go` | No edits since `v0.1.87` | Skim confirmed; no new regressions |

---

### 2026-07-20 — Acceptance & regression

**Scope:** `acceptance_manifest.json`, `acceptance_pr_ci.json`, `acceptance_manifest_test.go`, `scripts/run-acceptance-harness.sh`, `.github/workflows/acceptance-ci.yml`, `.github/workflows/dockhand-release-watch.yml`, `.github/workflows/compat-reports-sync.yml`, CI automation commits since `v0.1.87`.

**Summary:** Acceptance Full and Dockhand Release Watch are green on `main` per `release_gate_check.py`. Manifest and PR CI subset (13 suites) unchanged. This release window is CI-only: `actions/setup-go` v6→v7 (#210), Secret Smoke invalid-key reporting (#212), and compat baseline refreshes.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `acceptance_manifest_test.go:34-60` | 22 `manifestOperationExemptions` remain (mostly action delete checks) | Shrink incrementally on future agent branches (deferred) |
| med | `acceptance_pr_ci.json` | 13 suites vs ~85 manifest entries — nightly-only coverage by design | Document in CONTRIBUTING; rotate PR suites on minor releases (deferred) |
| low | `resource_new_surfaces_tf_acc_test.go` | `dockhand_job` skipped when batch completes inline | Optional coverage when API stable (deferred from v0.1.86) |
| low | `examples/scenarios/registry-and-image/main.tf:34` | `timestamp()` trigger causes perpetual diff | Replace with static trigger in follow-up (deferred) |
| — | `.github/workflows/*.yml` | `actions/setup-go` v6→v7 across CI workflows | Merged via Dependabot #210 |
| — | `secret-smoke.yml` | Reports invalid `CURSOR_API_KEY` via Bugbot API check | Fixed (#212) |

---

### 2026-07-20 — Security engineer

**Scope:** `internal/provider/provider.go`, `internal/provider/auth.go`, `internal/provider/client.go`, secret-bearing resources (sample grep), `.github/workflows/*`, `go.mod`, `.gitignore`.

**Summary:** No provider auth or secret-handling code changed since `v0.1.87`. Secret Smoke workflow now validates `CURSOR_API_KEY` against the Cursor Bugbot API and reports invalid keys in the automation health issue (#212). `actions/setup-go` bumped to v7 (#210). No `pull_request_target` workflows. Agent branches still cannot modify workflow files vs `main`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `.github/workflows/acceptance-ci.yml` | Throwaway CI password in workflow env (public repo) | Acceptable ephemeral creds; optional generated secret (deferred) |
| med | `internal/provider/auth.go:132` | Error bodies may echo server JSON | Truncate/sanitize diagnostics in future hardening (deferred) |
| low | CI failure artifacts | Session cookies possible in `dh-*.json` uploads | Scrub artifacts in follow-up (deferred) |
| — | `secret-smoke.yml:59-79` | Invalid Cursor key now surfaced in health issue body | Fixed (#212) |
| — | `.github/workflows/*.yml` | No `pull_request_target`; workflow integrity checks maintained | Current |

---

### 2026-07-20 — Release & upgrade

**Scope:** `CHANGELOG.md`, `README.md`, `docs/testing/release-gate.md`, `scripts/release_gate_check.py`, `scripts/release_housekeeping.py`, `.github/workflows/agent-release-tag.yml`, unreleased commit log since `v0.1.87`.

**Summary:** Patch-tier maintenance release: 24 commits since `v0.1.87` with no provider schema or client changes. CI improvements: Secret Smoke invalid-key reporting, `actions/setup-go` v7, pr-issue-link skip for hygiene automation branches. `release_gate_check.py` reports `ci_gates_pass: true`, `tier: patch`, zero blockers. `./scripts/verify.sh --quality` passes locally.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CHANGELOG.md` | Unreleased section empty; v0.1.88 notes not yet written | Post-tag housekeeping via **Agent Release Tag** / `release_housekeeping.py` |
| low | `README.md:23` | `version = ">= 0.1.63"` constraint | Bump to `>= 0.1.88` after tag (optional doc follow-up) |
| — | `scripts/release_gate_check.py` | Gates green; `ready_for_lens_dispatch: true` | Satisfied |
| — | `secret-smoke.yml` (#212) | Invalid Cursor key reported in automation health issue | Fixed |
| — | `pr-issue-link.yml` | Skip `cursor/repository-hygiene-automation-*` branches | Fixed |

---

### Release v0.1.88 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** probe fixture 404 tuning; backup API routes for future provider expansion; shrink `manifestOperationExemptions`; registry-and-image `timestamp()` example; README version pin bump; optional artifact scrubbing and auth error sanitization — carry forward from v0.1.85/v0.1.86/v0.1.87 (no new high-severity regressions)

---

## Release v0.1.89 — lens review

- **Tier:** patch
- **Started:** 2026-07-21
- **Base commit:** `f0d910f`
- **CI gates:** pass (`scripts/release_gate_check.py`)
- **Awaiting-release issues:** none
- **Unreleased commits:** 3 (CI/automation hardening, compat baseline refreshes)
- **Status:** clear to tag

---

### 2026-07-21 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`, `docs/reports/api-drift-gate.md`, `docs/reports/dockhand-last-tested.json`, `internal/provider/client_*.go` (grep for changes since `v0.1.88`).

**Summary:** No provider client or probe-list changes since `v0.1.88`. Compat Reports Sync refreshed baselines on 2026-07-21; Release Watch validated Dockhand `latest` (`sha256:871700eb…`). Probe tracks 156 routes with only the documented backlog absent (`GET /api/configs`, `GET /api/backups`). API drift gate reports zero new relevant endpoints.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `docs/reports/endpoint-probe.md:23-37` | 13 parameterized routes show unexpected 404 in static probe — fixture/safe-mode artifacts; acceptance and Release Watch exercise these paths on live Dockhand | Continue tuning probe fixtures on future agent branches; no release block |
| med | `docs/reports/docs-reference-api-endpoints.txt` | Docs drift audit lists `/api/backup/*` routes not yet in probe or provider | Track for future provider expansion; no release block |
| low | `docs/api-matrix.md` | Residual WebUI gap notes may lag newest resources | Periodic doc sweep on minor releases |
| — | `docs/non-present-endpoints.md:7-14` | Last-verified July 21, 2026; backlog matches probe | Current |
| — | `docs/reports/api-drift-gate.md:10` | `New relevant endpoints not allowlisted: 0` | Current |
| — | `internal/provider/client_*.go` | No edits since `v0.1.88` | Skim confirmed |

---

### 2026-07-21 — Terraform schema & state

**Scope:** `internal/provider/provider.go`, `resource_environment.go`, `resource_git_stack.go`, `resource_stack.go`, `resource_container.go` (grep for changes since `v0.1.88`).

**Summary:** No provider schema or state logic changed in the v0.1.89 window. v0.1.88 fix for `dockhand_environment` `public_ip` on create remains the latest provider change. Prior release fixes intact: one-shot `deploy_now`/`force_redeploy` reset, `enabled` reconciliation, inline batch failure detection, resolved `default_env` on action resources.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `internal/provider/resource_git_stack.go:192-197` | `env_vars_json` can hold secrets; marked Sensitive but write-only on read | Document write-only semantics in resource doc (deferred from v0.1.86) |
| med | `internal/provider/resource_stack.go`, `resource_container.go` | `enabled` reconciliation skips transitional statuses | Extend status map or emit diagnostic on unknown status (deferred) |
| low | `internal/provider/resource_stack.go:52-55` | Stack `id` Computed without `UseStateForUnknown` | Add plan modifier when touching stack resource (deferred) |
| — | `internal/provider/*.go` | No edits since `v0.1.88` | Skim confirmed; no new regressions |

---

### 2026-07-21 — Acceptance & regression

**Scope:** `acceptance_manifest.json`, `acceptance_pr_ci.json`, `acceptance_manifest_test.go`, `.github/workflows/acceptance-ci.yml`, `.github/workflows/dockhand-release-watch.yml`, `.github/workflows/compat-reports-sync.yml`, CI automation commits since `v0.1.88`.

**Summary:** Acceptance Full and Dockhand Release Watch are green on `main` per `release_gate_check.py`. Manifest and PR CI subset (13 suites) unchanged. This release window is CI-only: agent intake hardening (#222), release-drafter compat-sync exclusion, compat baseline refreshes (#220, #221).

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `acceptance_manifest_test.go:34-60` | 22 `manifestOperationExemptions` remain (mostly action delete checks) | Shrink incrementally on future agent branches (deferred) |
| med | `acceptance_pr_ci.json` | 13 suites vs ~85 manifest entries — nightly-only coverage by design | Document in CONTRIBUTING; rotate PR suites on minor releases (deferred) |
| low | `resource_new_surfaces_tf_acc_test.go` | `dockhand_job` skipped when batch completes inline | Optional coverage when API stable (deferred from v0.1.86) |
| low | `examples/scenarios/registry-and-image/main.tf:34` | `timestamp()` trigger causes perpetual diff | Replace with static trigger in follow-up (deferred) |
| — | `.github/workflows/issue-agent-intake.yml` | Issue title/body/comment passed via temp files — fixes backtick shell breakage | Fixed (#222) |
| — | `.github/workflows/agent-pr-approve-ci.yml` | Skips merge attempts on draft PRs (405 noise) | Fixed (#222) |

---

### 2026-07-21 — Security engineer

**Scope:** `internal/provider/provider.go`, `internal/provider/auth.go`, `internal/provider/client.go`, secret-bearing resources (sample grep), `.github/workflows/*`, `scripts/issue_agent_intake*.py`, `go.mod`, `.gitignore`.

**Summary:** No provider auth or secret-handling code changed since `v0.1.88`. CI hardening (#222) passes issue content via temp files instead of inline shell args (reduces quoting/injection risk from markdown backticks). Agent PR Approve CI skips draft merges. Release Drafter excludes compat-sync PRs from release notes. No `pull_request_target` workflows.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `.github/workflows/acceptance-ci.yml` | Throwaway CI password in workflow env (public repo) | Acceptable ephemeral creds; optional generated secret (deferred) |
| med | `internal/provider/auth.go:132` | Error bodies may echo server JSON | Truncate/sanitize diagnostics in future hardening (deferred) |
| low | CI failure artifacts | Session cookies possible in `dh-*.json` uploads | Scrub artifacts in follow-up (deferred) |
| — | `.github/workflows/issue-agent-intake.yml:93-101` | Issue body/title/labels/comment written to temp files before Python intake | Fixed (#222) |
| — | `.github/workflows/*.yml` | No `pull_request_target`; workflow integrity checks maintained | Current |

---

### 2026-07-21 — Release & upgrade

**Scope:** `CHANGELOG.md`, `README.md`, `docs/testing/release-gate.md`, `scripts/release_gate_check.py`, `.github/release-drafter.yml`, `.github/workflows/compat-reports-sync.yml`, unreleased commit log since `v0.1.88`.

**Summary:** Patch-tier maintenance release: 3 commits since `v0.1.88` with no provider schema or client changes. CI improvements: agent intake temp-file hardening, draft-PR merge skip, compat-sync excluded from Release Drafter notes. `release_gate_check.py` reports `ci_gates_pass: true`, `tier: patch`, zero blockers. `./scripts/verify.sh --quality` passes locally.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `CHANGELOG.md` | Unreleased section empty; v0.1.89 notes not yet written | Post-tag housekeeping via **Agent Release Tag** / `release_housekeeping.py` |
| low | `README.md:23` | `version = ">= 0.1.63"` constraint | Bump to `>= 0.1.89` after tag (optional doc follow-up) |
| — | `.github/release-drafter.yml:45-48` | `compat-reports-sync` and `skip-changelog` labels excluded from release notes | Fixed (#222) |
| — | `scripts/release_gate_check.py` | Gates green; `ready_for_lens_dispatch: true` | Satisfied |

---

### Release v0.1.89 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** probe fixture 404 tuning; backup API routes for future provider expansion; shrink `manifestOperationExemptions`; registry-and-image `timestamp()` example; README version pin bump; optional artifact scrubbing and auth error sanitization — carry forward from v0.1.85–v0.1.88 (no new high-severity regressions)

---

## Issue #252 — API drift detected: dockhand latest

- **Branch:** `agent/issue-252-api-drift-detected-dockhand-latest`
- **Lenses:** API compatibility; Ops / SRE; Acceptance & regression; Senior developer
- **Started:** 2026-08-09

### 2026-08-09 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`, `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`, `docs/reports/webui-api-endpoints.txt`, `internal/provider/client_settings_auth.go`.

**Summary:** Release Watch on Dockhand `latest` discovered `GET /api/settings/navigation` in the WebUI crawl; the route was relevant under `/api/settings` but absent from probe coverage and the allowlist. Added a safe GET probe entry alongside existing settings routes so the drift gate treats the path as tracked.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `scripts/endpoint-probe.py:18-19` | `/api/settings/navigation` missing from `ENDPOINTS`; caused `new_relevant_unallowlisted=1` on Release Watch | Add `GET /api/settings/navigation` (fixed) |
| med | `docs/reports/endpoint-probe.md` | Static probe report predates navigation route | **Compat Reports Sync** refreshes after green Acceptance Full / Release Watch |
| med | `docs/non-present-endpoints.md:16-18` | Backlog lists `/api/registry/tag-info` only; navigation is present on latest Dockhand | Probe tracking preferred over allowlist for live routes (fixed) |
| low | `docs/api-matrix.md` | No matrix row for navigation settings (UI-only) | Defer provider resource until operator demand — tracked via probe only |
| — | `scripts/api-drift-gate.py:26-52` | `/api/settings` prefix already in `RELEVANT_PREFIXES` | No change needed |

### 2026-08-09 — Ops / SRE

**Scope:** `.github/workflows/dockhand-release-watch.yml`, `.github/workflows/acceptance-full.yml`, `scripts/api-drift-gate.py`, `scripts/verify.sh`, `scripts/lens_sweep_gate.py`.

**Summary:** Drift failure is isolated to a single newly discovered WebUI route; probe list update is sufficient for Release Watch and Acceptance Full gates. No harness, workflow, or timeout changes required.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| med | `.github/workflows/dockhand-release-watch.yml` | Drift gate opens agent issue when `new_relevant_unallowlisted > 0` | Working as designed; this branch resolves #252 |
| low | `scripts/verify.sh` | `--endpoint-probe` remains debug-only (Dockhand env required) | No local probe run needed; CI validates |
| — | `scripts/lens_sweep_gate.py` | Requires `agent-review-log.md` update on agent branch | Satisfied by this log |

### 2026-08-09 — Acceptance & regression

**Scope:** `internal/provider/testdata/acceptance_manifest.json`, `internal/provider/testdata/acceptance_pr_ci.json`, `scripts/endpoint-probe.py`, `scripts/test_compat_reports_changed.py`, `scripts/test_release_watch_state.py`.

**Summary:** No provider resource or acceptance test changes required — navigation settings are UI configuration without a Terraform surface. Probe-only coverage matches prior drift fixes (e.g. `/api/settings/scanner/cache` in v0.1.86).

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| low | `acceptance_manifest.json` | No `dockhand_settings_navigation` resource (not requested) | Track navigation via probe only; open enhancement issue if Terraform surface is needed |
| — | `scripts/endpoint-probe.py` | Safe-mode GET probe for navigation | Fixed |
| — | `scripts/test_compat_reports_changed.py` | Compat report normalization unchanged | Skim confirmed |

### 2026-08-09 — Senior developer

**Scope:** `scripts/endpoint-probe.py` (settings block), `scripts/api-drift-gate.py` (path extraction), `CHANGELOG.md`.

**Summary:** Minimal one-line probe addition grouped with existing settings entries; no Go client changes. Matches repo convention for WebUI-only routes discovered by drift automation.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `scripts/endpoint-probe.py:20` | Single GET entry added after `settings/general` | Fixed |
| — | `CHANGELOG.md` | Unreleased entry documents drift fix | Fixed |
| low | `internal/provider/client_*.go` | No client method for navigation (UI-only) | Defer until provider resource is scoped |

---

## Release v0.1.90 — lens review

- **Lenses:** Release engineering; Ops / SRE; API compatibility; Senior developer
- **Started:** 2026-08-09

### 2026-08-09 — Release engineering

**Scope:** `scripts/release_gate_check.py`, `.github/workflows/agent-merge-cleanup.yml`, `.github/workflows/agent-autonomy-watchdog.yml`, `scripts/agent_merge_cleanup_watchdog.py`, `scripts/agent_release_loop_watchdog.py`, `CHANGELOG.md`.

**Summary:** v0.1.90 is a patch over v0.1.89: probe coverage for `GET /api/settings/navigation` (#252 / #254) plus release-loop recovery so bot squash-merges cannot leave compatibility issues open and stall tagging. No Terraform provider schema or client behavior change.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `scripts/release_gate_check.py` | Open `compatibility` issues blocked tagging even after `Fixes #N` merged | Skip issues with a merged closing PR (fixed) |
| high | `.github/workflows/agent-merge-cleanup.yml` | `GITHUB_TOKEN` merge skipped `push`/`pull_request` triggers | Add `workflow_run` after Agent PR Approve CI + 15m watchdog dispatch (fixed) |
| high | `.github/release-drafter.yml` | `exclude-labels` + `pre-exclude` category made Release Drafter v7 fail; no `v0.1.90` draft | Remove deprecated `exclude-labels` (fixed) |
| med | Release Drafter | Bot pushes never ran drafter after v0.1.89 | Watchdog dispatches `release-drafter.yml` when commits exist without a draft (fixed) |
| — | `docs/reports/agent-review-log.md` | Verdict recorded for tag workflow path filter | This section |

### 2026-08-09 — Ops / SRE

**Scope:** `docs/AGENT_AUTONOMY.md`, `docs/testing/release-gate.md`, `.github/workflows/agent-autonomy-watchdog.yml`.

**Summary:** Documented the token chicken-egg and hooked recovery into the existing 15m autonomy watchdog rather than a new cron.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `docs/testing/release-gate.md` | Gate checklist now excludes merged-but-open compat issues | Fixed |
| low | Watchdog cadence | Up to 15m delay before recovery | Acceptable; `workflow_run` covers the common Approve-CI merge path immediately |

### 2026-08-09 — API compatibility

**Scope:** `scripts/endpoint-probe.py`, issue #252, PR #254.

**Summary:** Navigation settings route is tracked in the safe probe list. No provider resource. Drift gate should pass on the next Release Watch / Acceptance Full against Dockhand `latest`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `scripts/endpoint-probe.py:20` | `GET /api/settings/navigation` present on `main` | Satisfied |
| low | `docs/api-matrix.md` | No Terraform surface for navigation | Defer |

### 2026-08-09 — Senior developer

**Scope:** Unreleased commits since `v0.1.89`, unit tests for gate + watchdogs.

**Summary:** Patch-tier release. Tests cover merged-fix skip, missing-draft commit counting, cleanup target filters, and lens/tag dispatch decisions.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `scripts/test_release_gate_check.py` | Merged #252 is not a blocker; missing draft still records unreleased commits | Fixed |
| — | Required CI workflows | Latest scheduled Acceptance Full / Release Watch green; Go CI on this PR | Validate before tag |

### Release v0.1.90 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** navigation settings Terraform resource; 15m watchdog delay when `workflow_run` does not see a completed merge yet; README version pin bump — carry forward, no high-severity product regressions

## Release v0.1.91 — lens review

- **Tier:** patch
- **Started:** 2026-08-10
- **Base commit:** `32702bceb7f8edad38fb14474dd3288e98e937a7`
- **CI gates:** pass (`release_gate_check.py`)
- **Status:** clear to tag

### 2026-08-10 — API compatibility

**Scope:** Unreleased commit since `v0.1.90` (`32702bc`), `scripts/endpoint-probe.py`, `docs/reports/endpoint-probe.md`, `docs/non-present-endpoints.md`, `docs/reports/dockhand-last-tested.json`.

**Summary:** Patch contains no provider client or probe changes. Dockhand Release Watch validated against `fnsys/dockhand:latest` on 2026-08-10; endpoint probe baseline unchanged (139 present, 2 not-present backlog).

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `32702bc` | CI-only Dependabot bump; no API surface change | No action |
| low | `docs/api-matrix.md` | Navigation settings still has no Terraform resource | Defer (carried from v0.1.90) |

### 2026-08-10 — Terraform schema & state

**Scope:** `internal/provider/` (grep for changes since `v0.1.90`), `internal/provider/provider.go`.

**Summary:** No provider resource, data source, or schema files changed in the unreleased window. State semantics and Plugin Framework behavior unchanged from v0.1.90.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `internal/provider/` | Zero diff since `v0.1.90` tag | No action |

### 2026-08-10 — Acceptance & regression

**Scope:** `internal/provider/testdata/acceptance_manifest.json`, `internal/provider/testdata/acceptance_pr_ci.json`, unreleased diff, CI gate output.

**Summary:** No acceptance test, manifest, or PR CI subset changes. Release gate reports CI green; Acceptance Full and Release Watch validated on current `main`.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | Manifest / PR CI | Unchanged since v0.1.90 | No action |
| — | `release_gate_check.py` | `ci_gates_pass: true`, `unreleased_commits_on_main: 1` | Satisfied |

### 2026-08-10 — Security engineer

**Scope:** `.github/workflows/release-artifacts.yml` (attest bump), provider auth/client files (skim), `.github/workflows/*` (skim for secret exfil patterns).

**Summary:** Dependabot bumps `actions/attest-build-provenance` v4.1.1 → v4.2.2 in the release artifact workflow only. Permissions remain scoped (`contents: write`, `id-token: write`, `attestations: write`); GPG signing and subject-path attestation unchanged. No provider secret-handling code touched.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `.github/workflows/release-artifacts.yml:134` | Attest action minor bump; same step inputs and OIDC permissions | Ship |
| — | Provider auth/client | No diff since v0.1.90 | No action |
| low | README | No pinned provider version example | Defer (carried from v0.1.90) |

### 2026-08-10 — Release & upgrade

**Scope:** `CHANGELOG.md` Unreleased, `scripts/release_gate_check.py`, `.github/workflows/agent-release-tag.yml`, Release Drafter draft for v0.1.91.

**Summary:** Patch-tier dependency-only release. Gate script reports tier `patch`, one unreleased commit, no open `awaiting-release` issues, no blocking compatibility issues. CHANGELOG Unreleased documents v0.1.90 product fixes; v0.1.91 will add the attest bump via release housekeeping on tag.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `release_gate_check.py` | `ready_for_lens_dispatch: true`, `lens_verdict_clear: false` until this log merges | This section |
| — | User impact | No Terraform operator action required beyond routine provider upgrade | Document in release notes on tag |
| low | `CHANGELOG.md` | Unreleased still lists v0.1.90 fixes until Agent Release Tag cuts the section | Expected; housekeeping on tag |

### Release v0.1.91 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
- **Deferred medium/low:** navigation settings Terraform resource; README version pin example — carried forward from v0.1.90, no new medium/low items

## Issue #267 — [Bug]: Can't use a `dockhand_git_repository` with existing repo, but new branch

### 2026-08-19 — Acceptance & regression

**Scope:** `internal/provider/resource_git_stack.go`, `internal/provider/resource_git_stack_test.go`, `internal/provider/resource_git_stack_tf_acc_test.go`, `internal/provider/testdata/acceptance_manifest.json`, `internal/provider/testdata/acceptance_pr_ci.json`

**Summary:** Root cause was a schema default (`branch = "main"`) on `dockhand_git_stack` that planned `main` even when only `repository_id` was set, then failed apply with inconsistent result when Dockhand returned the linked repository branch. Fix removes the default, adds unit tests for repository-id payload/state mapping, and adds `TestAccGitStackRepositoryIDInheritsBranchTerraform` acceptance coverage.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high | `resource_git_stack.go:120-124` | `branch` schema default `main` caused inconsistent result after apply when `repository_id` references a non-main branch | Fixed: remove default; branch computed from linked repo |
| — | `resource_git_stack_test.go` | Added unit tests for repository-id branch inheritance and URL default | Fixed in this branch |
| — | `resource_git_stack_tf_acc_test.go` | Added acceptance test for repository_id without explicit branch | Fixed in this branch |
| low | `acceptance_pr_ci.json` | New acceptance test not in PR CI subset (runs on full acceptance only) | Defer — existing git stack destroy test remains in PR CI |

### 2026-08-19 — Senior developer

**Scope:** `internal/provider/resource_git_stack.go` (schema, `buildGitStackPayload`, `mergeGitStackState`, `modelFromGitStackResponse`), `docs/resources/git_stack.md`

**Summary:** Bug was localized to Plugin Framework schema defaults conflicting with computed API-derived attributes. URL-based creation still defaults branch to `main` in payload builder; repository-id path correctly omits branch from POST body. No broader client or lookup-by-URL conflation found.

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| — | `resource_git_stack.go` | URL-mode branch default correctly handled in `buildGitStackPayload` | No action |
| — | `mergeGitStackState` | Remote branch already wins when API returns nested repository | No action |
| low | `resource_git_stack.go` | `url`/`repo_name`/`credential_id` also computed when using `repository_id`; same class of issue if defaults added later | Document only (done in git_stack.md) |

