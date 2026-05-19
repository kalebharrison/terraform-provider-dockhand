# Release Gate

Provider releases should only be cut when all of the following pass on `main`:

1. `Go CI`
2. `Quality CI`
3. `Govulncheck`
4. `Workflow Lint`
5. `Shell Lint`
6. `Gitleaks`
7. `Acceptance Full` (most recent scheduled/dispatch run)
8. `Dockhand Release Watch` (most recent run)

## Operational Gate Checklist

Before creating `vX.Y.Z`:

1. Ensure no open `compatibility` issues for current Dockhand release.
2. Confirm `TestAcceptanceManifestCoverage` passes.
3. Confirm docs/examples parity check passes (`/usr/bin/python3 scripts/check-doc-example-coverage.py`).
4. Confirm endpoint probe report is clean for current Dockhand target.
5. Confirm latest `Dockhand Release Watch` run produced compatibility artifacts (`endpoint-probe.*`, `webui-api-endpoints.txt`, `webui-endpoint-gap-audit.md`, `docs-reference-*`, `private-endpoint-probe.*`, `api-drift-gate.md`).
6. Cut signed tag:

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
- Change detection: compares latest discovered Dockhand `tag` and image `digest` to the release-watch state issue body (`last_tag=<value>`, `last_digest=<value>`).
- Only runs full validation when a new tag is discovered (or manual override via `workflow_dispatch` input).
- On success, updates the release-watch state issue using `GITHUB_TOKEN` (`issues:write`), so no extra repository secret is required.
- Includes docs-reference drift audit from `https://dockhand.pro/manual/#api-reference`.
- Includes a targeted authenticated private endpoint probe (`GET /api/environments` by default).
- Includes API drift gating that opens/updates an issue when new relevant endpoints are discovered and not yet tracked/allowlisted.

## API Drift Baseline Refresh

When a compatibility run fails on new API drift:

1. Review `api-drift-gate.md` artifact from the failed run.
2. Integrate new routes into provider/probe coverage where appropriate.
3. For accepted backlog gaps, add them to `docs/non-present-endpoints.md`.
4. Refresh baseline snapshots in `docs/reports/` after triage:

```bash
DOCKHAND_ENDPOINT=<endpoint> \
DOCKHAND_USERNAME=<username> \
DOCKHAND_PASSWORD=<password> \
DOCKHAND_AUTH_PROVIDER=local \
RUN_ENDPOINT_PROBE=true \
RUN_WEBUI_AUDIT=true \
RUN_DOCS_REFERENCE_AUDIT=true \
./scripts/run-acceptance-harness.sh
```
