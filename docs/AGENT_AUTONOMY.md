# Agent Autonomy (No Maintainer Machine Required)

**Audience:** Agents and automation — how this repo stays healthy without a developer laptop, production Dockhand, or home network.

## Rule 1

**Everything is automated via Cursor.** Routine triage, fixes, lens reviews, compatibility validation, and provider releases must not depend on maintainer prompts, `/agent`, or manual tagging.

## Principle

Routine work runs entirely in **GitHub Actions** and **Cursor Cloud Agents** against **ephemeral** Dockhand + DinD containers on CI runners. No one needs:

- `DOCKHAND_*` credentials to a real/production instance
- `terraform/dockhand/test/env.sh` (gitignored)
- Local Docker acceptance harness runs
- Local endpoint probe runs
- `./scripts/release-test.sh` against persistent infrastructure

Optional local commands exist only for **debugging** when CI fails.

## What agents do for routine changes

1. Branch `agent/issue-<n>-<slug>`
2. Run required review lenses (automatic mapping + CI gate)
3. Implement fix (code, tests, docs, examples, manifest)
4. Run `./scripts/verify.sh --quality` when convenient (same checks as CI unit gate)
5. Push — **Agent Validate** + PR **acceptance-ci** run on GitHub runners
6. **Agent Open PR** → add `agent-auto-merge` when PR sections are filled
7. **Agent PR Approve CI** → approve blocked checks and squash-merge when ready
8. Fix failures from **CI logs and artifacts** — not by reproducing on a laptop

See `docs/AGENT_RUNBOOK.md` and `docs/AGENT_CODING_STANDARDS.md`.

## What CI owns (vacation-proof)

| Concern | Workflow | Notes |
|---------|----------|-------|
| Unit + lint + docs parity | `go-ci.yml` | Every PR |
| PR acceptance (ephemeral Dockhand) | `acceptance-ci.yml` | Targeted `TestAcc` suites |
| Agent pre-PR gate | `agent-validate.yml` | `agent/**` pushes + lens log gate |
| Issue → Cloud Agent dispatch | `issue-agent-intake.yml` | User issues, CI failures, release candidates |
| Full acceptance + probes | `acceptance-full.yml` | Nightly + `workflow_dispatch` |
| New Dockhand image detection | `dockhand-release-watch.yml` | Every 6h scheduled validate + `force_validate` dispatch |
| **Committed compatibility baselines** | `compat-reports-sync.yml` | After green full/release-watch runs |
| **Release lens dispatch** | `agent-release-orchestrate.yml` | After green Release Watch on `main` (not on every `main` push) |
| **Signed tag + artifacts + housekeeping** | `agent-release-tag.yml` | When lens verdict clears on `main` |
| **Automation health alert** | `automation-health-notify.yml` | Opens tracker issue when release gate blockers persist ≥24h |
| **Agent stall watchdog** | `agent-stall-watchdog.yml` | Re-dispatches when Cloud Agent progress stalls ≥24h |
| **Dependabot auto-merge** | `dependabot-auto-merge.yml` | Enables squash auto-merge for Dependabot PRs |
| **Secret smoke** | `secret-smoke.yml` | Weekly: secrets, Actions settings, disable Bugbot via API |
| Dependency vulnerabilities | `govulncheck.yml` | Weekly + PR |
| Release zips | `release-artifacts.yml` | GPG-signed checksums + GitHub artifact attestations |

### Compatibility reports (no local refresh)

After a green **Acceptance Full** or **Dockhand Release Watch** run:

1. Harness probes ephemeral Dockhand (`scripts/endpoint-probe.py`, webui/docs audits).
2. Workflow uploads artifact `compat-reports-sync`.
3. **Compat Reports Sync** opens a PR (`agent/compat-reports-<run>`) updating baselines.

PRs are labeled `agent` + `agent-auto-merge` so they merge without a human.

## Automated release path

1. Fixes merge to `main`; **Agent Merge Cleanup** (agent PRs) or **Issue Resolution Notify** (human PRs) labels linked issues `awaiting-release`.
2. **Release Drafter** maintains the next draft version on each `main` push.
3. **Agent Release Orchestrate** opens `release: prepare vX.Y.Z` when `scripts/release_gate_check.py --mode lens` passes (after a successful **validated** **Dockhand Release Watch** on `main`, or via manual dispatch).
4. **Issue Agent Intake** dispatches a Cloud Agent with the release-tier lens set.
5. Agent appends lens sweeps + `### Release vX.Y.Z — verdict` with **Clear to tag: yes** to `docs/reports/agent-review-log.md` and merges via the normal agent PR loop.
6. **Agent Release Tag** runs the full release pipeline when `scripts/release_gate_check.py --mode tag` passes: ensure Release Watch validated, **Release Artifacts** (GPG-signed checksums + attestations), label `awaiting-release` issues, cut `CHANGELOG.md`, and post release comments on linked issues.

No maintainer prompt, manual tag, or laptop required. Use `release-issue-notify.yml` (`workflow_dispatch`) only to recover missed release comments.

## Failure triage (CI-first)

1. Open the failed GitHub Actions run.
2. For acceptance: download `*-logs-*` artifacts (scrubbed JSON).
3. Re-run failed workflow via `workflow_dispatch` when flaky.
4. Local `./scripts/run-acceptance-harness.sh` only for deep debugging — **not** a merge gate.

## Explicit non-goals

- Agents do not need VPN/home network access to your Dockhand.
- Cloud Agents do not run DinD locally (`docs/AGENT_DEPLOYMENT.md`).
- Maintainers are not a release gate — automation is.
- **Cursor Bugbot** PR comments are disabled automatically — **Secret Smoke** calls `scripts/cursor_bugbot_settings.py --disable` weekly when `CURSOR_API_KEY` is configured. You can also turn off Bugbot manually at [cursor.com/dashboard/bugbot](https://cursor.com/dashboard/bugbot).

## Related docs

- `docs/AGENT_DEPLOYMENT.md` — one-time rollout of agent CI
- `docs/ENDPOINT_PROBE.md` — probe behavior (CI is the default execution path)
- `docs/testing/release-gate.md` — programmatic release gate (`scripts/release_gate_check.py`)
