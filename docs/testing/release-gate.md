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
3. Confirm endpoint probe report is clean for current Dockhand target.
4. Cut signed tag:

```bash
git tag -s vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

## Why This Gate Exists

- Prevents publishing provider releases that regress on current Dockhand.
- Enforces parity between provider surface area and acceptance coverage metadata.
- Improves supply-chain confidence through signed tags and release provenance.

