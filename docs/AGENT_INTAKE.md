# Agent Issue Intake

How work enters the autonomous agent loop for `terraform-provider-dockhand`.

## Current model (manual dispatch)

There is **no** always-on watcher for **new** open issues yet. Work starts when:

1. A maintainer or agent **creates or selects a GitHub issue** with clear acceptance criteria.
2. An agent creates branch `agent/issue-<number>-<slug>`.
3. Push triggers **Agent Validate** → **Agent Open PR** → **Agent Auto Merge** when green.

**Closed-issue feedback** is handled by **Issue Regression Intake** (reopens + `regression` label). See `docs/AGENT_ISSUE_RESPONSE.md`.

This keeps humans in the loop for prioritization while automation handles validation and merge.

## Issue quality bar

Good agent issues include:

| Field | Example |
|-------|---------|
| **Problem** | `dockhand_batch_action reports success when job failed` |
| **Scope** | `internal/provider/resource_batch_action.go` + acceptance test |
| **Done when** | `./scripts/verify.sh --quality` passes; new test asserts failure |
| **Out of scope** | unrelated client refactors |

Labels (recommended):

- `bug`, `enhancement`, `documentation` — type
- `agent` — safe for autonomous pickup (no secrets, no release tagging)
- `good first issue` — small, well-bounded

## Dispatch options

### A. Cursor Cloud Agent (recommended)

1. Connect GitHub read/write on this repo (Cursor dashboard).
2. Open issue → **Assign to Cloud Agent** or paste issue URL into agent chat.
3. Agent reads `docs/AGENT_RUNBOOK.md`, `docs/AGENT_CODING_STANDARDS.md`, and `docs/AGENT_ISSUE_RESPONSE.md`.
4. Agent pushes `agent/issue-<n>-<slug>`; CI completes the loop.

### B. Local Cursor agent

1. `git checkout main && git pull`
2. `git checkout -b agent/issue-<n>-<slug>`
3. Implement fix per coding standards.
4. `./scripts/agent-commit-msg.sh "fix(provider): ..." | git commit -F -`
5. `git push -u origin HEAD`

### C. Future: GitHub Actions intake on new issues (optional)

Not implemented. **Issue Regression Intake** handles feedback on closed issues today.

## What agents should not pick up without human review

- Release tagging or registry publishing
- Branch protection or org-level settings changes
- Credential rotation or production Dockhand access changes
- Large cross-cutting refactors without a dedicated epic issue
- Issues requiring unaudited third-party API behavior

## After merge

1. **Issue Resolution Notify** posts on linked issues (what was fixed, awaiting release, reopen instructions).
2. Issue may auto-close via `Fixes #N` when the PR merges.
3. Maintainer cuts `vX.Y.Z` when ready (`docs/MAINTENANCE_PLAYBOOK.md`).
4. **Release Issue Notify** comments with version and upgrade steps.
5. **Lens review** before tag per `docs/testing/release-lens-review.md`.

See `docs/AGENT_ISSUE_RESPONSE.md` for the full issue communication standard.

## Regression / feedback on closed issues

**Issue Regression Intake** watches **new comments on closed issues**. When feedback indicates the fix did not work, it reopens the issue and adds `regression` + `agent` labels.

Agents should pick up reopened issues the same way as new work (`agent/issue-<n>-<slug>`).

## Future: GitHub Actions intake (optional)

A workflow could comment `/agent` or react 🤖 on **new** issues to spawn a branch. **Issue Regression Intake** already handles closed-issue feedback. Track optional `/agent` on open issues in `docs/AGENT_SWEEP.md`.

## Smoke test

Use `agent/issue-0-smoke-test` per `docs/AGENT_DEPLOYMENT.md` after enabling agent CI on `main`.
