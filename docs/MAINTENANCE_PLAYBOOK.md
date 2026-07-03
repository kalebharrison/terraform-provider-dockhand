# Maintenance Playbook

Operational runbook for keeping `terraform-provider-dockhand` stable, repeatable, and continuously updated.

**Vacation-proof maintenance:** see `docs/AGENT_AUTONOMY.md` — routine Dockhand validation runs in CI, not on a maintainer machine.

## Canonical Commands

Run from repo root.

```bash
# Core gate (fast, no Dockhand) — matches CI unit/docs checks
./scripts/verify.sh --quality

# Dependency/toolchain/security gate
./scripts/verify.sh --security
```

Dockhand-dependent commands (`--endpoint-probe`, `--acceptance`) are **debug-only** when CI is available. See `docs/ENDPOINT_PROBE.md` and `docs/AGENT_AUTONOMY.md`.

## Standard Change Flow

Human-maintainer branches use `codex/<short-change-name>`. Agent-managed work uses `agent/issue-<number>-<slug>` — see `docs/AGENT_RUNBOOK.md` and `docs/AGENT_CODING_STANDARDS.md`.

1. Sync and branch

```bash
git checkout main
git pull origin main
git checkout -b codex/<short-change-name>
```

2. Link work to an issue
- Create/update issue first.
- Use `Fixes #<id>` in PR body for user-facing fixes/features.

3. Implement narrowly scoped changes
- Keep API client/schema/tests/docs/examples in the same PR.

4. Run validation

**Agent branches:** push → Agent Validate + PR CI (no local Dockhand).

**Human branches:** optional `./scripts/verify.sh --quality` then push; CI runs acceptance.

```bash
./scripts/verify.sh --quality
./scripts/verify.sh --security
```

5. If API behavior changed, update `scripts/endpoint-probe.py` and merge — **Compat Reports Sync** refreshes `docs/reports/` after the next green nightly/release-watch run.

6. If resource/data source behavior changed, ensure acceptance coverage and `acceptance_pr_ci.json` when the suite should run on PRs. Full recursive acceptance runs nightly in CI.

7. Commit, push, open PR

```bash
git add -A
git commit -m "<type>: <summary>"
git push -u origin codex/<short-change-name>
gh pr create
```

8. Watch checks and fix until green

```bash
gh pr checks <pr-number>
```

## Required Surface Parity (Always)

When adding/changing a resource or data source:

1. Provider registration (`internal/provider/provider.go`)
2. Client/API mapping (`internal/provider/client.go` and/or resource/data source file)
3. Acceptance coverage (`internal/provider/*_tf_acc_test.go`)
4. Acceptance manifest mapping (`internal/provider/testdata/acceptance_manifest.json`)
5. Docs page (`docs/resources/*.md` or `docs/data-sources/*.md`)
6. Example file (`examples/resources/.../resource.tf` or `examples/data-sources/.../data-source.tf`)
7. Coverage matrix updates when endpoint status changes:
   - `docs/api-matrix.md`
   - `docs/non-present-endpoints.md` (if still missing/not exposed)

`/usr/bin/python3 scripts/check-doc-example-coverage.py` enforces docs/examples parity.
`TestAcceptanceManifestCoverage` enforces manifest parity and rejects bare `TestAcc` mappings. Every provider surface must map to an explicit targeted acceptance suite name.

## Release-First Validation Flow

For release candidate `X.Y.Z`:

1. Merge to `main`; confirm CI release gates in `docs/testing/release-gate.md`.
2. Agent runs full lens review (`docs/testing/release-lens-review.md`); log **clear to tag** in `docs/reports/agent-review-log.md`.
3. Tag and push:

```bash
git checkout main
git pull origin main
git tag -s vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

4. Wait for `.github/workflows/release-artifacts.yml` completion.
5. Confirm latest **Acceptance Full** and **Dockhand Release Watch** runs are green; **Compat Reports Sync** baseline PR merged or unchanged.

Optional staging check against a long-lived Dockhand (not a release gate):

```bash
./scripts/release-test.sh X.Y.Z
```

6. If release validation fails, fix on branch and cut next patch tag.

## Continuous Drift Detection (Hands-Off Loop)

Automated workflows already in place:

- `.github/workflows/dockhand-release-watch.yml` (every 6 hours)
  - Detects new Dockhand image tags.
  - Skips re-test when tag+digest match cached state (`dockhand-release-watch-state` Actions cache).
  - Runs acceptance harness + endpoint/webui/docs/private probes when Dockhand changes.
  - On success, **Compat Reports Sync** updates `docs/reports/dockhand-last-tested.json` (last validated Dockhand tag/digest).
  - Uploads compatibility artifacts.
- `.github/workflows/acceptance-full.yml` (nightly)
  - Full recursive acceptance run.
  - Uploads compat artifact for **Compat Reports Sync**.
- `.github/workflows/compat-reports-sync.yml`
  - Opens PR to refresh `docs/reports/` after green full/release-watch runs.

On failures:

1. Open/update issue with failing endpoint/resource details.
2. Patch provider + tests.
3. Merge with green checks.
4. Release patch version.

## Triage Cheatsheet

- `Schema Using Attribute Default For Non-Computed Attribute`
  - Mark attribute `Computed: true` when `Default` is set.
- `Provider produced inconsistent result after apply`
  - Preserve configured/state values when API response is partial.
- Acceptance flake around DinD/Hawser
  - Check `acceptance-ci` logs for env bootstrap, token, and ws URL wiring.
- `actionlint` shellcheck failures
  - Group repeated redirects (`{ ... } >> file`) and quote vars.

## Non-Negotiables

- No secret/token commits.
- No release without green required checks.
- No resource/data source merge without docs + examples + acceptance mapping.
- Prefer additive, backward-compatible schema changes unless intentionally breaking.
