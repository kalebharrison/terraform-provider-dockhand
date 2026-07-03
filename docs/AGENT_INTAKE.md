# Agent Issue Intake

How work enters the autonomous agent loop for `terraform-provider-dockhand`.

## Event-driven intake (default)

**Issue Agent Intake** (`.github/workflows/issue-agent-intake.yml`) dispatches a **Cursor Cloud Agent** automatically when:

| Trigger | Condition |
|---------|-----------|
| Issue **opened** by a user | Bug/feature reports from templates; no `/agent` required |
| Issue **opened** by CI with `compatibility` or `api-drift` | Release Watch compatibility/drift failures |
| Issue **labeled** `agent`, `compatibility`, `api-drift`, or `regression` | Re-dispatch or CI follow-up |
| Comment on open **regression** issue | Human follow-up re-dispatches |
| **`workflow_dispatch`** | Manual retry with issue number |

**Skipped automatically:**

- Labels `released`, `awaiting-release`, or `no-agent`
- Titles like `[Automation] Workflow failing: …` (workflow health trackers)
- Bot issues without `compatibility` / `api-drift` (unless manually dispatched)
- Issues already labeled `agent-dispatched` (unless `regression` or manual)

Eligibility rules live in `scripts/issue_agent_intake_eligibility.py` (unit-tested).

The workflow:

1. Creates branch `agent/issue-<n>-<slug>` from `main` (if missing)
2. Adds `agent`, `agent-dispatched`, and `in-progress` labels
3. Selects review lenses from labels/title and calls Cursor Cloud Agents API with runbook + lens sweep instructions
4. Comments with next automated steps

**Agent Validate** requires an update to `docs/reports/agent-review-log.md` on the agent branch before acceptance tests run.

**Required secret:** `CURSOR_API_KEY` (Cursor Dashboard → Integrations / API Keys). Add in GitHub **Settings → Secrets and variables → Actions**.

After the agent pushes:

1. **Agent Validate** → **Agent Open PR** (prefills **What was fixed** / **User impact** from issue body)
2. **Agent Approve CI** — approves pending PR workflow runs (no more yellow “Action required” gate)
3. **Agent Auto Merge** — when checks pass, PR has `agent-auto-merge`, and resolution sections are filled

## CI failure → agent bridge

| Source | Issue | Agent dispatch |
|--------|-------|----------------|
| **Dockhand Release Watch** `report_failure` | `Compatibility failure: dockhand …` | Yes — structured Problem/Done when + `agent` label |
| **Dockhand Release Watch** API drift gate | `API drift detected: dockhand …` | Yes — same pattern |
| **Automation Issue Notify** | `[Automation] Workflow failing: …` | No — `no-agent` tracker only (Release Watch removed from this notifier) |

## Issue quality bar

GitHub issue templates already include enough detail for user reports. Thin issues get an **Agent intake skipped** comment asking for Problem + Done when.

Good agent issues include:

| Field | Example |
|-------|---------|
| **Problem** | `dockhand_batch_action reports success when job failed` |
| **Scope** | `internal/provider/resource_batch_action.go` + acceptance test |
| **Done when** | `./scripts/verify.sh --quality` passes; new test asserts failure |
| **Out of scope** | unrelated client refactors |

Labels (informational):

- `bug`, `enhancement`, `documentation` — type
- `agent` — added automatically on dispatch (optional on open for visibility)
- `agent-auto-merge` — added automatically when PR sections are prefilled (opt-out by removing label)
- `no-agent` — excludes automation tracker issues from intake

## Manual override

**Actions → Issue Agent Intake → Run workflow** with an issue number bypasses most eligibility gates.

Optional: comment `/agent` on an open issue to re-dispatch after removing `agent-dispatched`.

## What stays outside routine automation

These are intentionally not auto-dispatched (security / blast radius):

- Branch protection or org-level settings changes
- Credential rotation or production Dockhand access changes
- Large cross-cutting refactors without a dedicated epic issue

## After merge

1. **Close linked issues after merge** (`agent-open-pr.yml`) closes issues, labels `awaiting-release`
2. **Issue Resolution Notify** posts resolution comments on linked issues
3. **Agent Release Orchestrate** → release lens issue → **Agent Release Tag** when gates pass
4. **Release Issue Notify** comments with version and upgrade steps

See `docs/AGENT_ISSUE_RESPONSE.md` for the full issue communication standard.

## Regression / feedback on closed issues

**Issue Regression Intake** watches **new comments on closed issues**. When feedback indicates the fix did not work, it reopens the issue, removes `agent-dispatched`, and adds `regression` + `agent` — which triggers **Issue Agent Intake** again.

## Smoke test

Four checks after enabling agent CI on `main`:

1. **CI loop** — branch `agent/issue-0-smoke-test` per `docs/AGENT_DEPLOYMENT.md` (Validate → Open PR; close without merge).
2. **Issue intake** — open a user-style smoke issue (e.g. `#126`) with done-when criteria; confirm dispatch without `/agent`.
3. **Fully automated loop** — smoke issue through **Agent Auto Merge** (e.g. `#129`).
4. **Issue auto-close** — smoke issue through merge plus **Close linked issues after merge** (e.g. `#132`).
