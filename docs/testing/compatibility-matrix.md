# Compatibility Matrix

This project validates provider compatibility against Dockhand using two recurring workflows:

- `Acceptance Full` (nightly): runs recursive acceptance (`TestAcc`) plus endpoint probe against `fnsys/dockhand:latest`.
- `Dockhand Release Watch` (every 6 hours): resolves latest Dockhand release tag and runs the same recursive suite against that image tag.

## Validation Scope

1. Provider acceptance tests (`go test -run TestAcc ./internal/provider`)
2. Provider-surface manifest parity (`TestAcceptanceManifestCoverage`)
3. API endpoint compatibility probe (`scripts/endpoint-probe.py`)

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
