# Compatibility Matrix

This project validates provider compatibility against Dockhand using two recurring workflows:

- `Acceptance Full` (nightly): runs recursive acceptance (`TestAcc`) plus endpoint probe against `fnsys/dockhand:latest`.
- `Dockhand Release Watch` (every 6 hours): resolves latest Dockhand release tag and runs the same recursive suite against that image tag.

## Validation Scope

1. Provider acceptance tests (`go test -run TestAcc ./internal/provider`)
2. Provider-surface manifest parity (`TestAcceptanceManifestCoverage`)
3. API endpoint compatibility probe (`scripts/endpoint-probe.py`)
4. API drift gate (`scripts/api-drift-gate.py`)

## Coverage Contract

- Every provider resource/data source in `internal/provider/provider.go` must exist in:
  - `internal/provider/testdata/acceptance_manifest.json`
- Manifest entries must define:
  - mode (`stateful`, `action`, `data_source`)
  - required lifecycle/read operations
  - acceptance test regex mapping
- Every manifest entry must point at an explicit targeted acceptance suite. Bare `TestAcc` catch-all mappings are rejected.

If a new provider surface is added without manifest coverage, CI fails.

## Failure Policy

If `Dockhand Release Watch` fails, the workflow opens a `compatibility` issue automatically.

`scripts/api-drift-gate.py` flags recurring compatibility runs when all of the following are true:

1. A route is newly discovered versus committed baseline snapshots in `docs/reports/`.
2. The route matches Terraform-relevant API prefixes (environments, stacks, containers, images, registry, schedules, auth, batch/jobs, etc.).
3. The route is not already tracked in `scripts/endpoint-probe.py`.
4. The route is not allowlisted in `docs/non-present-endpoints.md`.

When drift is detected in `Dockhand Release Watch`, the workflow opens or updates a `compatibility` + `api-drift` issue with the endpoint list.

This keeps existing known gaps non-blocking while making new Dockhand API drift immediately actionable.
