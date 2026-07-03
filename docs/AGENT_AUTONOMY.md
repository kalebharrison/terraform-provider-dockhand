# Agent Autonomy (No Maintainer Machine Required)

**Audience:** Agents and maintainers — how this repo stays healthy without a developer laptop, production Dockhand, or home network.

## Principle

Routine work runs entirely in **GitHub Actions** against **ephemeral** Dockhand + DinD containers on CI runners. Maintainers and agents never need:

- `DOCKHAND_*` credentials to a real/production instance
- `terraform/dockhand/test/env.sh` (gitignored)
- Local Docker acceptance harness runs
- Local endpoint probe runs
- `./scripts/release-test.sh` against persistent infrastructure

Optional local commands exist only for **debugging** when CI fails.

## What agents do for routine changes

1. Branch `agent/issue-<n>-<slug>`
2. Implement fix (code, tests, docs, examples, manifest)
3. Run `./scripts/verify.sh --quality` when convenient (same checks as CI unit gate)
4. Push — **Agent Validate** + PR **acceptance-ci** run on GitHub runners
5. **Agent Open PR** → add `agent-auto-merge` when appropriate
6. Fix failures from **CI logs and artifacts** — not by reproducing on a laptop

See `docs/AGENT_RUNBOOK.md` and `docs/AGENT_CODING_STANDARDS.md`.

## What CI owns (vacation-proof)

| Concern | Workflow | Notes |
|---------|----------|-------|
| Unit + lint + docs parity | `go-ci.yml` | Every PR |
| PR acceptance (ephemeral Dockhand) | `acceptance-ci.yml` | Targeted `TestAcc` suites |
| Agent pre-PR gate | `agent-validate.yml` | `agent/**` pushes |
| Full acceptance + probes | `acceptance-full.yml` | Nightly + `workflow_dispatch` |
| New Dockhand image detection | `dockhand-release-watch.yml` | Every 6h |
| **Committed compatibility baselines** | `compat-reports-sync.yml` | After green full/release-watch runs |
| Dependency vulnerabilities | `govulncheck.yml` | Weekly + PR |
| Release zips | `release-artifacts.yml` | On `v*` tag (GPG in repo secrets) |

### Compatibility reports (no local refresh)

After a green **Acceptance Full** or **Dockhand Release Watch** run:

1. Harness probes ephemeral Dockhand (`scripts/endpoint-probe.py`, webui/docs audits).
2. Workflow uploads artifact `compat-reports-sync`.
3. **Compat Reports Sync** opens a PR (`agent/compat-reports-<run>`) updating:
   - `docs/reports/endpoint-probe.md` / `.csv`
   - `docs/reports/webui-api-endpoints.txt`
   - `docs/reports/docs-reference-api-endpoints.txt`
   - `docs/non-present-endpoints.md` (last verified date)

PRs are labeled `agent` + `agent-auto-merge` so they can merge without a human.

Agents updating `scripts/endpoint-probe.py` should **not** run the probe locally — merge the code change; the next green nightly run refreshes reports.

## API / drift changes

When adding client routes:

1. Update `scripts/endpoint-probe.py` and `docs/api-matrix.md`.
2. Merge via normal agent PR + CI.
3. Let **Acceptance Full** / **Release Watch** + **Compat Reports Sync** refresh baselines.
4. If drift gate fails, triage from the `api-drift-gate.md` **artifact**; update allowlists in `docs/non-present-endpoints.md` in the same PR as probe changes when possible.

## Failure triage (CI-first)

1. Open the failed GitHub Actions run.
2. For acceptance: download `*-logs-*` artifacts (scrubbed JSON).
3. Re-run failed workflow via `workflow_dispatch` when flaky.
4. Local `./scripts/run-acceptance-harness.sh` only for deep debugging — **not** a merge gate.

## Release path (minimal human touch)

1. Green required checks on `main` (see `docs/testing/release-gate.md`).
2. Latest **Acceptance Full** + **Release Watch** green; compat report PR merged or current.
3. Release lens review logged in `docs/reports/agent-review-log.md`.
4. Signed tag via maintainer or future dispatch workflow (`release-artifacts.yml` uses repo `GPG_*` secrets — not a laptop keychain).

`./scripts/release-test.sh` is **optional** staging validation against a long-lived Dockhand — not required for the automated gate.

## Explicit non-goals

- Agents do not need VPN/home network access to your Dockhand.
- Agents do not tag releases without maintainer instruction (`docs/AGENT_INTAKE.md`).
- Cloud Agents do not run DinD locally (`docs/AGENT_DEPLOYMENT.md`).

## Related docs

- `docs/AGENT_DEPLOYMENT.md` — one-time rollout of agent CI
- `docs/ENDPOINT_PROBE.md` — probe behavior (CI is the default execution path)
- `docs/testing/release-gate.md` — pre-tag checklist (CI artifacts, not local env)
