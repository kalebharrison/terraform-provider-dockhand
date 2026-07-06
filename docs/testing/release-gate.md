# Release Gate

Provider releases should only be cut when all of the following pass on `main`:

1. `Go CI` (includes vet, golangci-lint, staticcheck, and shellcheck)
2. `Govulncheck`
3. `Workflow Lint`
4. `Gitleaks`
5. `dependency-review` (runs on PRs and on each `main` push with explicit base/head refs)
6. `Acceptance Full` (most recent scheduled/dispatch run)
7. `Dockhand Release Watch` — **validated** run on the current skip chain without intervening failures; full validation runs only when Dockhand tag/digest changes or provider `main` moves; gate-driven dispatches use `force_validate` only when the latest run failed

## Release lens review (automated)

**Agent Release Orchestrate** opens a `release: prepare vX.Y.Z` issue when CI gates pass and fixes are `awaiting-release`. It runs after a successful **Dockhand Release Watch** on `main` or via manual dispatch (no standalone cron). **Issue Agent Intake** dispatches a Cloud Agent with the release-tier lens set per `docs/testing/release-lens-review.md`. **Agent Release Tag** is the single release completion workflow: ensure Release Watch validated → **Release Artifacts** (GPG-signed checksums for Terraform Registry + GitHub artifact attestations) → label `awaiting-release` issues, cut `CHANGELOG.md`, and post release comments on linked issues.

Do not tag while **high** severity findings are open in the release verdict.

## Operational Gate Checklist

Programmatic gate: `scripts/release_gate_check.py`

- `--mode lens` — ready to open `release: prepare vX.Y.Z` for automated lens review (Release Watch: latest **validated** run strict)
- `--mode tag` — ready for **Agent Release Tag** to publish the GitHub release (Release Watch: **validated** success on current `main` SHA)

Before **Agent Release Tag** publishes `vX.Y.Z`:

1. No open `compatibility` issues for the current Dockhand release.
2. Required workflows green on `main` (Go CI, Govulncheck, Workflow Lint, Gitleaks, Dependency Review, Acceptance Full, Dockhand Release Watch).
3. Release Drafter draft exists for `vX.Y.Z` and the tag is not published yet.
4. At least one `awaiting-release` issue exists (for lens dispatch) **or** the review log already contains **Clear to tag: yes** for `vX.Y.Z`.
5. `docs/reports/agent-review-log.md` on `main` contains `### Release vX.Y.Z — verdict` with **Clear to tag: yes**.

Tagging and artifact publish are performed by `.github/workflows/agent-release-tag.yml`:

1. `scripts/ensure_release_watch_green.py` dispatches Release Watch when needed and waits for a **validated** run (with one retry in CI).
2. **Release Artifacts** builds zips, signs `SHA256SUMS` with GPG (Terraform Registry), attaches GitHub artifact attestations, and creates the GitHub release with `gh release create --target main`.
3. `scripts/release_housekeeping.py` labels all open `awaiting-release` issues `released` and cuts `CHANGELOG.md`.
4. **Agent Release Tag** posts upgrade/release comments on issues linked from the release notes (recovery-only: `release-issue-notify.yml` via `workflow_dispatch`).

## Why This Gate Exists

- Prevents publishing provider releases that regress on current Dockhand.
- Enforces parity between provider surface area and acceptance coverage metadata.
- Improves supply-chain confidence through GPG-signed Terraform Registry checksums and GitHub artifact attestations.

## Release Watch Behavior

- Workflow: `.github/workflows/dockhand-release-watch.yml`
- Poll cadence: every 6 hours at `:10` UTC (lightweight discover only; skips full validation when Dockhand tag/digest and provider `main` SHA are unchanged).
- Triggers: `schedule`, `workflow_dispatch` (`force_validate`, optional `image_tag`). Does **not** run on every `main` push (provider-only pushes would otherwise produce skip-only runs).
- Full validation runs when: Dockhand tag or digest changes, provider `main` SHA changes, state is unset, or `force_validate` / manual `image_tag` is used. **No time-based re-validation.**
- Periodic full coverage without a Dockhand bump: **Acceptance Full** (nightly).
- On discover cache miss, migrates legacy issue #38 state and **seeds** the Actions cache immediately so skip logic does not re-query issues every run.
- On success, saves state to the Actions cache (no tracking issue). **Compat Reports Sync** commits `docs/reports/dockhand-last-tested.json` for a human-readable last-validated record.
- Includes docs-reference drift audit from `https://dockhand.pro/manual/#api-reference`.
- Includes a targeted authenticated private endpoint probe (`GET /api/environments` by default).
- Includes API drift gating that opens/updates an issue when new relevant endpoints are discovered and not yet tracked/allowlisted.

## API Drift Baseline Refresh

When a compatibility run fails on new API drift:

1. Review `api-drift-gate.md` artifact from the failed run.
2. Integrate new routes into provider/probe coverage where appropriate.
3. For accepted backlog gaps, add them to `docs/non-present-endpoints.md`.
4. After a green **Acceptance Full** or **Release Watch** run, **Compat Reports Sync** updates `docs/reports/` baselines on `main` (or auto-merges a fallback PR when branch rules block direct push).

Manual local harness runs are debug-only; see `docs/AGENT_AUTONOMY.md`.
