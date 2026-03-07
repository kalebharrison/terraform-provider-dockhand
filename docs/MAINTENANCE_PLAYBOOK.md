# Maintenance Playbook

Operational runbook for keeping `terraform-provider-dockhand` stable, repeatable, and continuously updated.

## Canonical Commands

Run from repo root.

```bash
# Core local gate (fast, repeatable)
./scripts/verify.sh

# Extended gate (matches CI quality checks where local tools exist)
./scripts/verify.sh --quality

# API contract check against live Dockhand
./scripts/verify.sh --endpoint-probe

# Full acceptance harness (ephemeral Dockhand + DinD + Hawser)
./scripts/verify.sh --acceptance --test-regex 'TestAcc'
```

## Standard Change Flow

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

4. Run local validation

```bash
./scripts/verify.sh --quality
```

5. If API behavior changed, run live probe

```bash
./scripts/verify.sh --endpoint-probe
```

6. If resource/data source behavior changed, run acceptance harness

```bash
./scripts/verify.sh --acceptance --test-regex 'TestAcc(<targeted-regex>)'
```

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

## Release-First Validation Flow

For release candidate `X.Y.Z`:

1. Merge to `main`.
2. Tag and push:

```bash
git checkout main
git pull origin main
git tag -s vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

3. Wait for `.github/workflows/release-artifacts.yml` completion.
4. Validate the published artifact, not local binaries:

```bash
./scripts/release-test.sh X.Y.Z
```

5. If release validation fails, fix on branch and cut next patch tag.

## Continuous Drift Detection (Hands-Off Loop)

Automated workflows already in place:

- `.github/workflows/dockhand-release-watch.yml` (every 6 hours)
  - Detects new Dockhand image tags.
  - Runs acceptance harness + endpoint/webui/docs/private probes.
  - Uploads compatibility artifacts.
- `.github/workflows/acceptance-full.yml` (nightly)
  - Full recursive acceptance run.

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
