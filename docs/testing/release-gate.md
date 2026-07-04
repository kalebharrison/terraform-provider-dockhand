# Release Gate

Provider releases should only be cut when all of the following pass on `main`:

1. `Go CI` (includes vet, golangci-lint, staticcheck, and shellcheck)
2. `Govulncheck`
3. `Workflow Lint`
4. `Gitleaks`
5. `dependency-review`
6. `Acceptance Full` (most recent scheduled/dispatch run)
7. `Dockhand Release Watch` — **strict** for lens dispatch (latest run on `main` must succeed); **main SHA** for tag publish (any successful run on current `main` HEAD counts, even if a later run failed)

## Release lens review (automated)

**Agent Release Orchestrate** opens a `release: prepare vX.Y.Z` issue when CI gates pass and fixes are `awaiting-release`. It runs after a successful **Dockhand Release Watch** on `main`, on schedule, or via manual dispatch (not on every `main` push or Release Drafter alone). **Issue Agent Intake** dispatches a Cloud Agent with the release-tier lens set per `docs/testing/release-lens-review.md`. **Agent Release Tag** is the single release completion workflow: ensure Release Watch green → **Release Artifacts** (GPG-signed checksums for Terraform Registry + GitHub artifact attestations) → label `awaiting-release` issues and cut `CHANGELOG.md`.

Do not tag while **high** severity findings are open in the release verdict.

## Operational Gate Checklist

Programmatic gate: `scripts/release_gate_check.py`

- `--mode lens` — ready to open `release: prepare vX.Y.Z` for automated lens review (Release Watch: latest run strict)
- `--mode tag` — ready for **Agent Release Tag** to publish the GitHub release (Release Watch: success on current `main` SHA)

Before **Agent Release Tag** publishes `vX.Y.Z`:

1. No open `compatibility` issues for the current Dockhand release.
2. Required workflows green on `main` (Go CI, Govulncheck, Workflow Lint, Gitleaks, Acceptance Full, Dockhand Release Watch).
3. Release Drafter draft exists for `vX.Y.Z` and the tag is not published yet.
4. At least one `awaiting-release` issue exists (for lens dispatch) **or** the review log already contains **Clear to tag: yes** for `vX.Y.Z`.
5. `docs/reports/agent-review-log.md` on `main` contains `### Release vX.Y.Z — verdict` with **Clear to tag: yes**.

Tagging and artifact publish are performed by `.github/workflows/agent-release-tag.yml`:

1. `scripts/ensure_release_watch_green.py` dispatches Release Watch when needed (with one retry in CI).
2. **Release Artifacts** builds zips, signs `SHA256SUMS` with GPG (Terraform Registry), attaches GitHub artifact attestations, and creates the GitHub release with `gh release create --target main`.
3. `scripts/release_housekeeping.py` labels all open `awaiting-release` issues `released` and cuts `CHANGELOG.md`.

## Why This Gate Exists

- Prevents publishing provider releases that regress on current Dockhand.
- Enforces parity between provider surface area and acceptance coverage metadata.
- Improves supply-chain confidence through GPG-signed Terraform Registry checksums and GitHub artifact attestations.

## Release Watch Behavior

- Workflow: `.github/workflows/dockhand-release-watch.yml`
- Poll cadence: every 6 hours.
- Harness retry: failed acceptance harness runs once automatically before marking the workflow failed.
- Change detection: compares latest discovered Dockhand `tag` and image `digest` to cached release-watch state (`last_tag`, `last_digest` in Actions cache key `dockhand-release-watch-state`).
- Only runs full validation when a new tag is discovered (or manual override via `workflow_dispatch` input).
- On success, saves state to the Actions cache (no tracking issue). **Compat Reports Sync** commits `docs/reports/dockhand-last-tested.json` for a human-readable last-validated record.
- Includes docs-reference drift audit from `https://dockhand.pro/manual/#api-reference`.
- Includes a targeted authenticated private endpoint probe (`GET /api/environments` by default).
- Includes API drift gating that opens/updates an issue when new relevant endpoints are discovered and not yet tracked/allowlisted.

## API Drift Baseline Refresh

When a compatibility run fails on new API drift:

1. Review `api-drift-gate.md` artifact from the failed run.
2. Integrate new routes into provider/probe coverage where appropriate.
3. For accepted backlog gaps, add them to `docs/non-present-endpoints.md`.
4. After a green **Acceptance Full** or **Release Watch** run, **Compat Reports Sync** opens a PR updating `docs/reports/` baselines — merge it (or let `agent-auto-merge` handle it).

Manual local harness runs are debug-only; see `docs/AGENT_AUTONOMY.md`.
