# Agent Review Lenses

Focused reviews for the AI agent. **Automated on every agent issue** via Issue Agent Intake; **full pass required before every release tag**.

Each lens lists **Does not cover** to avoid overlap with other lenses.

## When to run

### Every agent issue (automated)

**Issue Agent Intake** selects lenses from issue labels/title (`scripts/issue_agent_intake_lenses.py`) and embeds required sweep steps in the Cloud Agent prompt.

**Agent Validate** enforces that `docs/reports/agent-review-log.md` is updated on the agent branch before acceptance tests run (`scripts/lens_sweep_gate.py`).

| Issue signal | Lenses (typical) |
|--------------|------------------|
| `compatibility`, `api-drift`, compatibility failure title | API compatibility, Ops / SRE (+ bug lenses) |
| `bug`, `[Bug]:` | Acceptance & regression, Senior developer |
| `enhancement`, `[Feature]:` | Terraform schema & state, Acceptance & regression, GitOps / IaC practitioner |
| `regression` | Acceptance & regression (+ issue-type lenses) |
| `security` | Security engineer |
| Default | Senior developer, Acceptance & regression |

### Before every `v*` release (required)

Run the lens set for the **release tier** below. Append each sweep to `docs/reports/agent-review-log.md` under a single release header.

- **Block the tag** if any finding is **high** severity and unresolved.
- **Medium** findings: fix before tag or document explicit deferral in the release notes / review log.
- **Low** findings: may ship with issues filed for follow-up.

The maintainer says *"prepare release vX.Y.Z"* or *"run release lens review"* to start the release-tier pass (not covered by per-issue automation above).

See `docs/testing/release-lens-review.md` for tier definitions and lens order.

### Ad hoc (maintainer override)

| Trigger | Suggested lens |
|---------|----------------|
| Maintainer names a lens | That lens only |
| After auth/credential change | Security engineer |
| After new resource/data source | Terraform schema & state + Acceptance & regression |
| After CI/harness change | Ops / SRE |

Per-issue automation already runs API compatibility on Dockhand drift failures. Use ad hoc when you need an extra lens beyond the intake mapping.

**Do not** run lenses on a timer or `/loop`.

## How to run a sweep

1. Read the lens **Goal**, **Does not cover**, **Priority paths**, and **Checklist** in this doc.
2. Search and read code; do not skim filenames only.
3. Append findings to `docs/reports/agent-review-log.md`.
4. Open GitHub issues for actionable items.
5. Fix only small, obvious problems in the same sweep; defer large work to issues/branches.

## Lens index

| Lens | When it's most useful |
|------|------------------------|
| Terraform schema & state | Resource/data source changes, apply drift reports |
| Security engineer | Auth, secrets, dependencies, workflow changes |
| Dockhand domain / runtime | Environment, stack, git, container, hawser work |
| Async & long-running operations | Action resources, jobs, streams, polling |
| Acceptance & regression | New tests, manifest changes, CI test gaps |
| Ops / SRE | Harness, workflows, flakes |
| API compatibility | Dockhand upgrades, probe/drift failures |
| Senior developer | Refactors, large files, general Go quality |
| GitOps / IaC practitioner | Docs, examples, HCL ergonomics |
| Entry-level developer | README, CONTRIBUTING, onboarding |
| Release & upgrade | Before tagging a release |

Full checklists below.

### Finding format

```markdown
### YYYY-MM-DD — <lens name>

**Scope:** <what you read>
**Summary:** <1-2 sentences>

| Severity | Location | Finding | Suggested action |
|----------|----------|---------|------------------|
| high/med/low | path:line | ... | ... |
```

---

## Lens: Terraform schema & state

**Goal:** Plugin Framework correctness — plan, apply, read, import, and state stability.

**Does not cover:** API endpoint presence (API compatibility), example HCL quality (GitOps), security of secrets (Security), Dockhand connection semantics (Dockhand domain).

**Priority paths:**

- `internal/provider/provider.go`
- `internal/provider/resource_*.go`, `data_source_*.go` (sample 5–8 recently changed or high-traffic)
- Shared schema helpers if present

**Checklist:**

- [ ] `Computed` / `Optional` / `Default` — no inconsistent result after apply?
- [ ] Read preserves user-configured values when API returns partial or empty fields?
- [ ] Write-only / sensitive attributes not leaked into state?
- [ ] Import ID format stable, documented, and matches `ImportState` logic?
- [ ] Plan modifiers / `UseStateForUnknown` used where API omits fields?
- [ ] Action vs stateful resource boundaries correct (no state where there should be none)?
- [ ] `RequiresReplace` / `ForceNew` aligned with immutable Dockhand fields?

---

## Lens: Security engineer

**Goal:** No credential leakage; safe defaults; minimal attack surface.

**Does not cover:** Schema computed/optional correctness (Terraform schema), CI flake debugging (Ops), whether tests exist (Acceptance).

**Priority paths:**

- `internal/provider/provider.go`, `internal/provider/client.go`, `internal/provider/auth.go`
- Resources with secrets: `*_git_*`, `*_registry*`, `*_auth_*`, `*_environment*`, `*_stack_env*`
- `.github/workflows/*`, `scripts/*`, `.gitignore`, `examples/`

**Checklist:**

- [ ] Sensitive schema attributes marked `Sensitive: true`?
- [ ] Passwords/tokens write-only where appropriate; not in logs/diagnostics?
- [ ] TLS `insecure` default safe; documented?
- [ ] Error messages avoid echoing secrets or full cookie headers?
- [ ] CI uses only throwaway creds; no real secrets in repo?
- [ ] No `pull_request_target` secret exfiltration patterns?
- [ ] Dependencies: note any govulncheck / dependency-review concerns?

---

## Lens: Dockhand domain / runtime

**Goal:** Provider models Dockhand the way operators actually use it.

**Does not cover:** Generic Terraform schema rules (Terraform schema), HTTP retry code (Senior dev), endpoint probe lists (API compatibility), test manifest rows (Acceptance).

**Priority paths:**

- `dockhand_environment`, `dockhand_stack`, `dockhand_git_stack`, `dockhand_container*`
- `dockhand_schedule*`, `dockhand_registry*`, hawser-related tests/resources
- Matching `docs/resources/*.md` for sampled resources

**Checklist:**

- [ ] `connection_type` socket / direct / agent (`hawser-edge`) maps correctly?
- [ ] Hawser agent token create/rotate/read behavior matches Dockhand UI/API?
- [ ] Stack vs git stack vs stack_action — right tool for deploy vs compose-on-disk?
- [ ] `default_env` / per-resource `env` consistent across resources?
- [ ] Container lifecycle (create, actions, rename, update) matches Docker/Dockhand semantics?
- [ ] Registry/git credential flows match Dockhand's credential storage model?
- [ ] Prune/batch/schedule behavior matches operator expectations?

---

## Lens: Async & long-running operations

**Goal:** Jobs, streams, polls, and actions fail clearly and eventually complete or timeout.

**Does not cover:** Static CRUD schema (Terraform schema), whether API route exists (API compatibility), CI harness boot (Ops).

**Priority paths:**

- `*_action.go` resources (batch, deploy, scan, push, schedule run, git deploy/webhook)
- `internal/provider/client.go` — job polling, stream readers
- Data sources: `dockhand_job`, schedule stream/executions

**Checklist:**

- [ ] Poll loops have max attempts/duration and actionable timeout errors?
- [ ] Terminal job statuses (`done`, `failed`, `cancelled`, etc.) handled consistently?
- [ ] Stream endpoints (deploy-stream, logs) fully read/closed?
- [ ] User docs say when apply returns before background work finishes?
- [ ] Idempotent re-run of actions safe where Dockhand allows it?

---

## Lens: Acceptance & regression

**Goal:** Tests prove real behavior against Dockhand — not just that CI is green.

**Does not cover:** Workflow YAML reliability (Ops), schema design (Terraform schema), user-facing doc prose (Entry-level / GitOps).

**Priority paths:**

- `internal/provider/*_tf_acc_test.go` (sample 5 suites)
- `internal/provider/testdata/acceptance_manifest.json`
- `internal/provider/testdata/acceptance_pr_ci.json`
- `internal/provider/acceptance_manifest_test.go`, `acceptance_pr_ci_test.go`

**Checklist:**

- [ ] Every manifest entry regex matches at least one real `TestAcc*` function?
- [ ] Stateful resources: tests cover create + read + destroy (or import)?
- [ ] Action resources: tests cover create + read + delete lifecycle?
- [ ] PR CI subset still exercises env, hawser/agent, registry, git, container paths?
- [ ] `t.Skip` for missing env — acceptable, or masking untested surface?
- [ ] New critical resources added to `acceptance_pr_ci.json` when appropriate?
- [ ] Flakes tied to bootstrap order documented in playbook or issues?

---

## Lens: Ops / SRE

**Goal:** CI and harness are reliable, fast enough, and debuggable.

**Does not cover:** Test assertion quality (Acceptance), API drift content (API compatibility), provider schema (Terraform schema).

**Priority paths:**

- `.github/workflows/*`
- `scripts/run-acceptance-harness.sh`, `scripts/verify.sh`
- Agent workflows: `agent-validate.yml`, `agent-open-pr.yml`, `agent-auto-merge.yml`

**Checklist:**

- [ ] Harness boot order: registry → DinD → Dockhand → Hawser — healthy?
- [ ] Failure artifacts (`*-logs-*`) sufficient to debug without re-run?
- [ ] Timeouts and concurrency groups appropriate?
- [ ] `verify.sh` flags align with CI jobs?
- [ ] Agent loop wired: validate → open PR → auto-merge?
- [ ] Repeated flakes have open issues or playbook notes?

---

## Lens: API compatibility

**Goal:** Provider client and surface track live Dockhand API reality.

**Does not cover:** Terraform UX of schema (Terraform schema / GitOps), test design (Acceptance), CI wiring (Ops).

**Priority paths:**

- `internal/provider/client.go` (request/response structs)
- `scripts/endpoint-probe.py`, `scripts/api-drift-gate.py`
- `docs/api-matrix.md`, `docs/non-present-endpoints.md`, `docs/reports/endpoint-probe.md`

**Checklist:**

- [ ] Probe report clean for target Dockhand version?
- [ ] New endpoints in drift gate integrated or allowlisted with reason?
- [ ] Client structs match live JSON shapes (null vs missing, nested fields)?
- [ ] `non-present-endpoints.md` backlog still accurate?
- [ ] Release watch / acceptance-full artifacts reviewed if recent failure?

---

## Lens: Senior developer

**Goal:** Go code quality, structure, and maintainability — not TF- or Dockhand-specific rules covered elsewhere.

**Does not cover:** Schema/state (Terraform schema), secrets (Security), Dockhand ops model (Dockhand domain), tests (Acceptance), API routes (API compatibility).

**Priority paths:**

- `internal/provider/client.go` (structure, not API parity)
- `internal/provider/request_retry.go`, `internal/provider/auth.go`
- Any file >500 LOC touched recently

**Checklist:**

- [ ] Error wrapping and messages consistent and actionable?
- [ ] Duplicated logic that should be shared helpers?
- [ ] God files growing without boundary (e.g. monolithic `client.go`)?
- [ ] Context passed through HTTP calls?
- [ ] Dead code or orphaned helpers?
- [ ] Matches repo naming and file layout conventions?

---

## Lens: GitOps / IaC practitioner

**Goal:** Terraform **users** get clear HCL, examples, and resource choice guidance.

**Does not cover:** Plugin Framework internals (Terraform schema), API mapping (API compatibility), README onboarding (Entry-level).

**Priority paths:**

- `docs/resources/`, `docs/data-sources/`, `examples/`
- `docs/index.md`, provider argument docs
- Sample: one stack scenario, one git stack scenario, one environment scenario

**Checklist:**

- [ ] Examples copy-pasteable; variables explained?
- [ ] When to use action vs resource vs data source — clear in docs?
- [ ] `default_env` and per-resource `env` documented for multi-env users?
- [ ] Import blocks documented where supported?
- [ ] Breaking attribute changes noted in resource docs?

---

## Lens: Entry-level developer

**Goal:** A new contributor succeeds without insider knowledge.

**Does not cover:** Deep schema/API review (other lenses), example HCL correctness (GitOps — only onboarding path).

**Priority paths:**

- `README.md`, `CONTRIBUTING.md`, `AGENTS.md`
- `docs/LOCAL_DEV.md`, `docs/MAINTENANCE_PLAYBOOK.md`, `docs/AGENT_*.md`
- `.env.example`, `scripts/verify.sh --help`

**Checklist:**

- [ ] First-time clone → validate path obvious?
- [ ] Jargon explained (Dockhand, Hawser, DinD, manifest, acceptance harness)?
- [ ] Human vs `agent/**` branch workflow clear?
- [ ] Where to look when CI fails?
- [ ] How to report security issues?

---

## Lens: Release & upgrade

**Goal:** Safe, documented provider releases for consumers upgrading versions.

**Does not cover:** Day-to-day code quality (other lenses). Run before tagging `v*`.

**Priority paths:**

- `README.md` (version pins), release workflow, recent CHANGELOG if present
- Schema changes in last release window
- `docs/testing/release-gate.md`

**Checklist:**

- [ ] User-visible changes documented for next release?
- [ ] Schema breaking changes called out with migration notes?
- [ ] Examples use realistic `version` constraints?
- [ ] Release gate checklist in `docs/testing/release-gate.md` satisfied?
- [ ] Compatibility issues closed or explicitly deferred?

---

## Sweep depth

No fixed time budget. This repo is a single Go provider package (~90 source files) — most lenses are a **focused pass**, not a multi-hour audit.

Scale depth to **what changed since the last tag**, not a clock:

| Depth | When | What to do |
|-------|------|------------|
| **Skim** | Lens area untouched in this release | Priority-path grep + confirm no accidental edits |
| **Standard** | Lens area touched or always-on for release type | Read changed files + checklist |
| **Deep** | Incident, major release, or findings in standard pass | Widen to full lens priority paths |

Do not read entire `client.go` every lens unless that lens owns it (Senior dev, API compatibility) **and** the file changed.

For release reviews, use the tier in `docs/testing/release-lens-review.md` (full 11 vs core 5) — not a per-lens minute target.
