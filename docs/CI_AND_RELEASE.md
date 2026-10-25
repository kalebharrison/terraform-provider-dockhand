# CI and release

Two loops. No Cloud Agent factory.

## Loop A — Dockhand changes

1. **Dockhand Compat** polls Docker Hub every 6 hours.
2. New image (or nightly force at `02:43` UTC) → full `TestAcc` + endpoint probe + webui/docs audits.
3. Nightly and `strict_drift=true` dispatch fail on new relevant API routes (`API_DRIFT_FAIL_ON_NEW`).
4. Failures open a `compatibility` issue with what broke.
5. Green validate uploads reports → **Compat Reports Sync** refreshes `docs/reports/`.

## Loop B — People issues

1. Open/fix GitHub issues as usual.
2. Implement in a PR with `Fixes #N` (enforced by **PR Policy**).
3. **Go CI** + **Acceptance CI** + **Security** must pass.
4. Merge labels linked issues `awaiting-release`.
5. After **Go CI** and **Security** both succeed on that `main` push, **Release** publishes. Changelog cuts and compatibility-report refreshes do not publish again.

## Release

1. **Release** workflow updates the Release Drafter draft on every `main` push.
2. A user-fix or dependency merge publishes itself once Go CI and Security are green. **Actions → Release → Run workflow** still publishes on demand.
3. Gate (`scripts/release_gate_check.py --mode tag`) requires:
   - Go CI, Security, Dockhand Compat green
   - No blocking open compatibility issues (merged `Fixes #` counts as resolved)
   - A Release Drafter draft that is not yet published
   - `awaiting-release` issues **or** commits on `main` since the latest tag
4. On success: **Release Artifacts** publishes GPG-signed zips; housekeeping labels issues and cuts `CHANGELOG.md`; release comments are posted on linked issues.

```bash
gh workflow run release.yml --ref main
./scripts/release_gate_check.py --mode tag --json   # debug only
```

## Local checks

```bash
./scripts/verify.sh --quality
```

Full Dockhand acceptance is CI-only by default.
