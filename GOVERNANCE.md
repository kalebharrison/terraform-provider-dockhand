# Governance

## Model

This project follows an **agent-autonomous model** documented in `docs/AGENT_AUTONOMY.md`.

- Routine bugs, CI failures, compatibility drift, and releases are handled by **Cursor Cloud Agents** and **GitHub Actions**.
- Humans retain **ops-only** responsibilities: secrets rotation, org/branch settings, and security advisories that require judgment.
- Changes land on `main` through protected-branch CI and automated agent pull requests.

## Decision Process

1. Open or link an issue (templates provide Problem/Done when sections).
2. **Issue Agent Intake** dispatches a Cloud Agent when eligible.
3. Agent work merges via **Agent Auto Merge** when required checks pass.
4. Releases publish via **Agent Release Tag** after automated release lens review.

Human pull requests remain supported for ops and trust-boundary changes (for example agent workflow files).

## Evolution

As automation coverage grows, update `docs/AGENT_SWEEP.md` and this document to reflect new hands-off paths and remaining ops-only tasks.
