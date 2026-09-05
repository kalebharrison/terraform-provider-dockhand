# Governance

## Model

This project follows a **maintainer + CI** model:

- **GitHub Actions** watches Dockhand, runs acceptance/security CI, and publishes releases.
- **Humans (optionally with Cursor IDE)** triage issues, implement fixes, and dispatch **Release** when ready.
- There is no autonomous Cloud Agent intake or auto-merge loop.

See `docs/CI_AND_RELEASE.md` and `AGENTS.md`.

## Decision process

1. Open or link an issue.
2. Land a PR through protected-branch required checks.
3. Dispatch **Release** when the release gate is green and there is release work.

Ops-only human work: secrets rotation, org/branch settings, security advisories.
