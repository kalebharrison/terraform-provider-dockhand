# Contributing

Thanks for contributing to `terraform-provider-dockhand`.

## Workflow

1. Open or identify a GitHub issue for the change.
2. Create a branch from `main`.
3. Make focused changes and include tests.
4. Open a pull request with a closing reference in the body, for example:
   - `Fixes #123`
5. Wait for CI to pass before merge.
6. Merge to `main`.
7. Cut a release tag (`vX.Y.Z`) when changes are ready for distribution.

Operational details and repeatable release flow: `docs/MAINTENANCE_PLAYBOOK.md`.

## Required PR Content

- A clear summary of behavior changes.
- Linked issue in PR body (`Fixes #...`).
- Validation evidence (`go test ./...`, acceptance tests when applicable).

## Local Validation

Run from repo root:

```bash
./scripts/verify.sh --quality

# If API contracts changed:
./scripts/verify.sh --endpoint-probe

# If behavior changes require runtime coverage:
./scripts/verify.sh --acceptance --test-regex 'TestAcc(<targeted-regex>)'
```

## Release Notes

- User-visible behavior changes should be called out in the PR.
- Release artifacts are published from git tags via GitHub Actions.

## Security

- Never commit secrets, API tokens, private keys, or local override files.
- If you discover a security issue, open a private report with the repository owner.
