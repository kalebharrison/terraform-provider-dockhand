# Agent Issue Intake

How work enters the autonomous agent loop for `terraform-provider-dockhand`.

## Event-driven intake (default)

**Issue Agent Intake** (`.github/workflows/issue-agent-intake.yml`) dispatches a **Cursor Cloud Agent** when:

| Trigger | Condition |
|---------|-----------|
| Issue **labeled** `agent` | Label added to an open issue |
| Issue **opened** with `agent` | Rare; same handler |
| Comment **`/agent`** on open issue | Adds `agent` label if missing, then dispatches |
| **`workflow_dispatch`** | Manual retry with issue number |

The workflow:

1. Creates branch `agent/issue-<n>-<slug>` from `main` (if missing)
2. Calls Cursor Cloud Agents API (`POST /v1/agents`) with runbook prompt
3. Labels issue `agent-dispatched` + `in-progress`
4. Comments with next automated steps

**Required secret:** `CURSOR_API_KEY` (Cursor Dashboard → Integrations / API Keys). Add in GitHub **Settings → Secrets and variables → Actions**.

After the agent pushes:

1. **Agent Validate** → **Agent Open PR** (prefills **What was fixed** / **User impact** from issue body)
2. **Agent Approve CI** — approves pending PR workflow runs (no more yellow “Action required” gate)
3. **Agent Auto Merge** — when checks pass, PR has `agent-auto-merge`, and resolution sections are filled

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
- `agent` — triggers Cloud Agent dispatch
- `agent-auto-merge` — added automatically when PR sections are prefilled (opt-out by removing label)
- `good first issue` — small, well-bounded

## Manual dispatch options

### A. Label-only (recommended)

1. Open or edit issue with clear acceptance criteria
2. Add label **`agent`**
3. Wait for intake comment + Cloud Agent run

### B. Comment dispatch

Comment on an open issue:

```text
/agent
```

### C. Cursor Cloud Agent (direct)

1. Connect GitHub read/write on this repo (Cursor dashboard)
2. Assign issue to Cloud Agent or paste issue URL into agent chat
3. Ensure branch follows `agent/issue-<n>-<slug>` contract

### D. Local Cursor agent

1. `git checkout main && git pull`
2. `git checkout -b agent/issue-<n>-<slug>`
3. Implement fix per coding standards
4. `./scripts/agent-commit-msg.sh "fix(provider): ..." | git commit -F -`
5. `git push -u origin HEAD`

## What agents should not pick up without human review

- Release tagging or registry publishing
- Branch protection or org-level settings changes
- Credential rotation or production Dockhand access changes
- Large cross-cutting refactors without a dedicated epic issue
- Issues requiring unaudited third-party API behavior

## After merge

1. **Agent Close Linked Issues** (in `agent-auto-merge.yml`) closes issues linked via `Fixes #N` when **Agent Auto Merge** squash-merges a bot PR (GitHub does not auto-close on `github-actions[bot]` merges)
2. **Issue Resolution Notify** posts on linked issues (what was fixed, awaiting release, reopen instructions)
3. Maintainer cuts `vX.Y.Z` when ready (`docs/MAINTENANCE_PLAYBOOK.md`)
4. **Release Issue Notify** comments with version and upgrade steps
5. **Lens review** before tag per `docs/testing/release-lens-review.md`

See `docs/AGENT_ISSUE_RESPONSE.md` for the full issue communication standard.

## Regression / feedback on closed issues

**Issue Regression Intake** watches **new comments on closed issues**. When feedback indicates the fix did not work, it reopens the issue, removes `agent-dispatched`, and adds `regression` + `agent` — which triggers **Issue Agent Intake** again.

## Re-dispatch

Remove label `agent-dispatched`, then comment `/agent` or re-add `agent`, or use **Actions → Issue Agent Intake → Run workflow**.

## Smoke test

Four checks after enabling agent CI on `main`:

1. **CI loop** — branch `agent/issue-0-smoke-test` per `docs/AGENT_DEPLOYMENT.md` (Validate → Open PR; close without merge).
2. **Issue intake** — open a smoke issue labeled `agent` (e.g. `#126`) with done-when criteria for dispatch → Cloud Agent commit → Validate → Open PR → Approve CI; close PR and issue without merge when green.
3. **Fully automated loop** — open a smoke issue labeled `agent` (e.g. `#129`) with done-when criteria through **Agent Auto Merge** (Validate → Open PR → Approve CI → PR CI green → squash merge); no manual merge or workflow approval.
4. **Issue auto-close** — open a smoke issue labeled `agent` (e.g. `#132`) with done-when criteria through merge plus **Close linked issues after merge** (issue closed, `awaiting-release` label, resolution comment from Agent Close Linked Issues); no manual `gh issue close`.
