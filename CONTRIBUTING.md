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

Autonomous agent loop: `docs/AGENT_RUNBOOK.md` (coding standards: `docs/AGENT_CODING_STANDARDS.md`; deploy: `docs/AGENT_DEPLOYMENT.md`; intake: `docs/AGENT_INTAKE.md`).

## Agent-managed changes

This repository uses a transparent agent workflow for some fixes:

- Branches: `agent/issue-<number>-<slug>`
- Validation: **Agent Validate** (GitHub Actions)
- Pull requests: opened by **Agent Open PR** as `github-actions[bot]` with the `agent` label
- Commits: include `Co-authored-by: Cursor Agent <noreply@cursor.com>`

Details: `docs/AGENT_IDENTITY.md`.

## First-time setup

```bash
git clone https://github.com/kalebharrison/terraform-provider-dockhand.git
cd terraform-provider-dockhand
./scripts/verify.sh --quality
```

Requires Go 1.25+ (see `go.mod`). For acceptance tests you also need Docker and Terraform; see `docs/LOCAL_DEV.md` and `docs/MAINTENANCE_PLAYBOOK.md`.

Human branches: `codex/<short-description>`. Agent branches: `agent/issue-<number>-<slug>` (see `docs/AGENT_RUNBOOK.md`).

## Glossary

| Term | Meaning |
|------|---------|
| **Dockhand** | The container management platform this provider configures |
| **Hawser** | Dockhand's agent that connects remote Docker hosts (`connection_type = "agent"`) |
| **DinD** | Docker-in-Docker; used in CI acceptance harness for isolated Docker environments |
| **Acceptance harness** | `scripts/run-acceptance-harness.sh` — spins up Dockhand + DinD for `TestAcc*` tests |
| **Manifest** | `acceptance_manifest.json` — maps each provider resource/data source to an acceptance test |

## Required PR Content

- A clear summary of behavior changes.
- Linked issue in PR body (`Fixes #...`).
- Validation evidence (`go test ./...`, acceptance tests when applicable).

## Local validation (optional)

Routine work does not require a local Dockhand. See `docs/AGENT_AUTONOMY.md`.

```bash
./scripts/verify.sh --quality
```

Dockhand-dependent checks run in CI. Debug-only locally when investigating failures:

```bash
./scripts/verify.sh --acceptance --test-regex 'TestAcc(<targeted-regex>)'
./scripts/verify.sh --endpoint-probe
```

## CI failures

When a GitHub Actions job fails:

1. Open the failed workflow run and read the job log (for acceptance, download `*-logs-*` artifacts when present).
2. Re-run via `workflow_dispatch` when appropriate.
3. Local reproduction is optional — not required before merge when CI passes on retry.
4. Agent-managed branches: see failure handling in `docs/AGENT_RUNBOOK.md`.

List PR checks from the CLI: `gh pr checks <number>`.

## Release Notes

- User-visible behavior changes should be called out in the PR.
- Release artifacts are published from git tags via GitHub Actions.

## Security

- Never commit secrets, API tokens, private keys, or local override files.
- Use environment variables or a local `.env` file for Dockhand credentials. See `.env.example` for variable names; `.env` is gitignored.
- Do not commit Terraform state (`.tfstate`), `*.tfvars` with real values, or local CLI override files (`.terraformrc*`).
- CI workflows use throwaway test credentials for ephemeral containers — not real Dockhand instances.
- Report security vulnerabilities privately via [SECURITY.md](SECURITY.md) (GitHub Security Advisories preferred). Do not file public issues for security bugs.
