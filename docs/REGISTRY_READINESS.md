# Registry Readiness Checklist

This checklist is for preparing a public release path for Terraform Registry and OpenTofu ecosystem usage.

## Current State

- Build/test workflow exists: `.github/workflows/go-ci.yml`
- Release zip workflow exists: `.github/workflows/release-artifacts.yml`
- Release validation helper exists: `scripts/release-test.sh`

## Before Public Release

1. Repository visibility and naming
- Ensure repository is public.
- Keep provider source address stable (`kalebharrison/dockhand`).

2. Release artifact quality
- Verify release assets include all supported platform zips plus:
  - `terraform-provider-dockhand_<version>_SHA256SUMS`
  - `terraform-provider-dockhand_<version>_SHA256SUMS.sig`
- Keep naming pattern: `terraform-provider-dockhand_<version>_<os>_<arch>.zip`.

3. GitHub Actions signing setup
- Add repository secret `GPG_PRIVATE_KEY`:
  - ASCII-armored private key for the signing identity.
- Add repository secret `GPG_PASSPHRASE`:
  - Passphrase for the private key.
- Keep the matching public key available for Terraform Registry onboarding.

4. Provider documentation completeness
- Keep `docs/index.md` resource/data source list current.
- Keep `docs/api-matrix.md` and `docs/non-present-endpoints.md` current.
- Include acceptance test prerequisites in `README.md`.

5. Security and hygiene
- No secrets or local override files in git history.
- Keep `.gitignore` covering local test state and mirrors.
- Keep endpoint probes and acceptance tests running against non-production fixtures.

6. Registry onboarding tasks
- Terraform Registry:
  - Follow HashiCorp provider publishing/onboarding steps for namespace ownership and releases.
  - Configure the provider signing key in registry onboarding using the public key that matches release signatures.
- OpenTofu:
  - Decide distribution model (filesystem mirror, release assets, or OpenTofu registry integration).
  - Publish matching versioned artifacts and checksums.

## Ongoing Release Gate

Run this sequence for each release candidate:

```bash
go test ./...
go build ./...
source terraform/dockhand/test/env.sh
/usr/bin/python3 scripts/endpoint-probe.py
./scripts/release-test.sh <version>
```

Then verify:

- GitHub Release assets are present and downloadable.
- `docs/reports/endpoint-probe.md` reflects current endpoint contract.
