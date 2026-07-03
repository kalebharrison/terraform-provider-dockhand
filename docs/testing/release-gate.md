# Release Gate

Provider releases should only be cut when all of the following pass on `main`:

1. `Go CI` (includes vet, golangci-lint, staticcheck, and shellcheck)
2. `Govulncheck`
3. `Workflow Lint`
4. `Gitleaks`
5. `dependency-review`
6. `Acceptance Full` (most recent scheduled/dispatch run)
7. `Dockhand Release Watch` (most recent run)

## Release lens review (automated)

**Agent Release Orchestrate** opens a `release: prepare vX.Y.Z` issue when CI gates pass and fixes are `awaiting-release`. **Issue Agent Intake** dispatches a Cloud Agent with the release-tier lens set per `docs/testing/release-lens-review.md`. **Agent Release Tag** publishes the signed tag when `docs/reports/agent-review-log.md` contains **Clear to tag: yes** for that version.

Do not tag while **high** severity findings are open in the release verdict.

## Operational Gate Checklist

Programmatic gate: `scripts/release_gate_check.py`

- `--mode lens` — ready to open `release: prepare vX.Y.Z` for automated lens review
- `--mode tag` — ready for **Agent Release Tag** to publish the signed tag

Before **Agent Release Tag** publishes `vX.Y.Z`:

1. No open `compatibility` issues for the current Dockhand release.
2. Required workflows green on `main` (Go CI, Govulncheck, Workflow Lint, Gitleaks, Acceptance Full, Dockhand Release Watch).
3. Release Drafter draft exists for `vX.Y.Z` and the tag is not published yet.
4. At least one `awaiting-release` issue exists (for lens dispatch) **or** the review log already contains **Clear to tag: yes** for `vX.Y.Z`.
5. `docs/reports/agent-review-log.md` on `main` contains `### Release vX.Y.Z — verdict` with **Clear to tag: yes**.

Tagging is performed by `.github/workflows/agent-release-tag.yml` (signed GPG tag in Actions). **Release Artifacts** runs on tag push.

## Why This Gate Exists

- Prevents publishing provider releases that regress on current Dockhand.
- Enforces parity between provider surface area and acceptance coverage metadata.
- Improves supply-chain confidence through signed tags and release provenance.

## Release Watch Behavior

- Workflow: `.github/workflows/dockhand-release-watch.yml`
- Poll cadence: every 6 hours.
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
