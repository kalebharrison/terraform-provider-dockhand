# Maintenance Playbook

Operational runbook for `terraform-provider-dockhand`.

**CI model:** `docs/CI_AND_RELEASE.md` — Dockhand validation and releases run in GitHub Actions.

## Canonical commands

```bash
./scripts/verify.sh --quality
./scripts/verify.sh --security
```

Dockhand-dependent commands (`--endpoint-probe`, `--acceptance`) are debug-only when CI is available. See `docs/ENDPOINT_PROBE.md`.

## Standard change flow

1. Branch from `main` (`codex/<slug>` or `agent/issue-<n>-<slug>`).
2. Link an issue; use `Fixes #<id>` in the PR body.
3. Implement narrowly (client/schema/tests/docs/examples together).
4. `./scripts/verify.sh --quality` then open a PR.
5. Merge when required checks are green (`Lint, Test, Build`, acceptance, security, PR policy).

Coding conventions: `docs/AGENT_CODING_STANDARDS.md`.

## Compatibility / Dockhand bumps

1. **Dockhand Compat** opens a `compatibility` issue when validation or drift fails.
2. Fix the provider (or allowlist intentional gaps in `docs/non-present-endpoints.md`).
3. Confirm **Dockhand Compat** is green; **Compat Reports Sync** refreshes `docs/reports/` when needed.

## Cutting a release

1. Confirm release work exists (`awaiting-release` issues or commits since last tag).
2. Optional check: `./scripts/release_gate_check.py --mode tag --json`
3. **Actions → Release → Run workflow** (`release.yml`).
4. Confirm GitHub release + artifacts; issues labeled `released` and commented.

## Related

- `docs/CI_AND_RELEASE.md`
- `docs/ENDPOINT_PROBE.md`
- `docs/REGISTRY_READINESS.md`
- `.github/workflows/` (9 workflows)
