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
| **Automated handoff watchdog** | `agent-autonomy-watchdog.yml` | Every 15m: unblock trusted agent/automation/Cursor PRs; dispatch eligible queued issues (CI, release, compatibility); recover missed merge cleanup, Release Drafter drafts, and release lens/tag dispatches after `GITHUB_TOKEN` merges |
| **Dependabot auto-merge** | `dependabot-auto-merge.yml` | Enables squash auto-merge for Dependabot PRs |
| **Secret smoke** | `secret-smoke.yml` | Weekly: secrets, Actions settings, disable Bugbot via API |
| Dependency vulnerabilities | `govulncheck.yml` | Weekly + PR |
| Release zips | `release-artifacts.yml` | GPG-signed checksums + GitHub artifact attestations |

### Compatibility reports (no local refresh)

After a green **Acceptance Full** or **Dockhand Release Watch** run:

1. Harness probes ephemeral Dockhand (`scripts/endpoint-probe.py`, webui/docs audits).
2. Workflow uploads artifact `compat-reports-sync`.
3. **Compat Reports Sync** commits refreshed baselines to `main` when branch rules allow; otherwise it opens one shared auto-merge PR (`automation/compat-reports-sync`), approves blocked checks, and merges without a maintainer. Those PRs are labeled `compat-reports-sync` + `skip-changelog` so **Release Drafter** omits them from release notes.
4. Sync is a **no-op** when only volatile metadata changed (`updated_at` / `source_run` / the non-present `- Date:` line). **Dockhand Release Watch** also skips re-validation when `main` advanced only via prior compat-sync commits (same Dockhand digest).

No manual merge or workflow approval is required for routine report refresh.

## Automated release path

1. Fixes merge to `main`; **Agent Merge Cleanup** (agent PRs) or **Issue Resolution Notify** (human PRs) labels linked issues `awaiting-release`.
2. **Release Drafter** maintains the next draft version on each `main` push.
3. **Agent Release Orchestrate** opens `release: prepare vX.Y.Z` when `scripts/release_gate_check.py --mode lens` passes (validated **Dockhand Release Watch** on `main`, or manual dispatch). Release work includes `awaiting-release` issues **or** any commits on `main` since the latest published tag.
4. **Issue Agent Intake** dispatches a Cloud Agent with the release-tier lens set.
5. Agent appends lens sweeps + `### Release vX.Y.Z — verdict` with **Clear to tag: yes** to `docs/reports/agent-review-log.md` and merges via the normal agent PR loop.
6. **Agent Release Tag** runs the full release pipeline when `scripts/release_gate_check.py --mode tag` passes: ensure Release Watch validated, **Release Artifacts** (GPG-signed checksums + attestations), label `awaiting-release` issues, cut `CHANGELOG.md`, and post release comments on linked issues.

No maintainer prompt, manual tag, or laptop required. Use `release-issue-notify.yml` (`workflow_dispatch`) only to recover missed release comments.

## Failure triage (CI-first, agent-owned)

**Do not use maintainer chat or direct `main` pushes for routine repair.** The automation loop owns:

| Failure | Automated path |
|---------|----------------|
| `main` CI / Workflow Lint / actionlint | **Automation Issue Notify** → `CI failure: …` issue → **Issue Agent Intake** → agent PR → **Agent PR Approve CI** |
| Trusted automated PR waiting on checks/merge | **Agent Autonomy Watchdog** (every 15m) approves blocked runs and re-dispatches merge |
| Eligible `agent` issue not dispatched (CI, release, compatibility) | **Agent Autonomy Watchdog** re-queues **Issue Agent Intake** |
| Cloud Agent stall (no branch/commits) | **Agent Stall Watchdog** clears `agent-dispatched` and re-triggers intake |

Maintainer chat and laptop commands are **debug-only** when automation is broken or secrets are missing.

When investigating manually:

1. Open the failed GitHub Actions run.
2. For acceptance: download `*-logs-*` artifacts (scrubbed JSON).
3. Re-run failed workflow via `workflow_dispatch` when flaky.
4. Local `./scripts/run-acceptance-harness.sh` only for deep debugging — **not** a merge gate.

## Explicit non-goals

- Agents do not need VPN/home network access to your Dockhand.
- Cloud Agents do not run DinD locally (`docs/AGENT_DEPLOYMENT.md`).
- Maintainers are not a release gate — automation is.
- **Cursor Bugbot** PR comments are disabled when the Bugbot admin API is available — **Secret Smoke** calls `scripts/cursor_bugbot_settings.py --disable --best-effort` weekly. That API is Team/Enterprise-only; on solo/Pro+ accounts the step skips without failing (turn Bugbot off manually at [cursor.com/dashboard/bugbot](https://cursor.com/dashboard/bugbot) if needed). `CURSOR_API_KEY` validity is checked via Cloud Agents `GET /v1/me`, not Bugbot.

## Related docs

- `docs/AGENT_DEPLOYMENT.md` — one-time rollout of agent CI
- `docs/ENDPOINT_PROBE.md` — probe behavior (CI is the default execution path)
- `docs/testing/release-gate.md` — programmatic release gate (`scripts/release_gate_check.py`)
