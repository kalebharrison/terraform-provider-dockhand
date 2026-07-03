# Release Gate

Provider releases should only be cut when all of the following pass on `main`:

1. `Go CI` (includes vet, golangci-lint, staticcheck, and shellcheck)
2. `Govulncheck`
3. `Workflow Lint`
4. `Gitleaks`
5. `dependency-review`
6. `Acceptance Full` (most recent scheduled/dispatch run)
7. `Dockhand Release Watch` (most recent run)

## Release lens review (agent)

Before creating `vX.Y.Z`, the agent runs the **release-tier lens set** per `docs/testing/release-lens-review.md` (core 5 for patch, all 11 for minor/major) and logs to `docs/reports/agent-review-log.md`.

Do not tag while **high** severity findings are open.

## Operational Gate Checklist

Before creating `vX.Y.Z`:

1. Ensure no open `compatibility` issues for current Dockhand release.
2. Confirm `TestAcceptanceManifestCoverage` passes.
3. Confirm docs/examples parity check passes (`/usr/bin/python3 scripts/check-doc-example-coverage.py`).
4. Confirm committed `docs/reports/endpoint-probe.md` is current **or** a green **Acceptance Full** / **Release Watch** run completed since the last probe script change ( **Compat Reports Sync** PR merged or pending).
5. Confirm latest `Dockhand Release Watch` run produced compatibility artifacts (`endpoint-probe.*`, `webui-api-endpoints.txt`, `webui-endpoint-gap-audit.md`, `docs-reference-*`, `private-endpoint-probe.*`, `api-drift-gate.md`).
6. Complete release lens review (`docs/testing/release-lens-review.md`) — **clear to tag** verdict in review log.
7. Cut signed tag:

```bash
git tag -s vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

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
