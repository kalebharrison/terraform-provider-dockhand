# Agent Workflow Deployment

**Audience:** Maintainers — one-time rollout after merging agent CI to `main`.

One-time checklist after merging agent CI changes to `main`.

## 1. Sync GitHub branch protection

`.github/settings.yml` is the source of truth. Apply to GitHub (settings bot or manual edit):

**Remove required check:**

- `Vet, Lint, Staticcheck` (merged into `Lint, Test, Build`)

**Add required check:**

- `dependency-review`

**Keep existing checks:**

- `Lint, Test, Build`
- `Dockhand + DinD Acceptance Tests`
- `Require linked issue and mark in-progress`
- `Conventional PR title`
- `Secret Scan`
- `actionlint`
- `Analyze (Go) (go)`
- `Go Vulnerability Check`

Verify:

```bash
gh api repos/kalebharrison/terraform-provider-dockhand/branches/main/protection \
  --jq '.required_status_checks.contexts'
```

## 2. Ensure labels exist

- `agent` — automated agent PRs; **triggers Issue Agent Intake**
- `agent-auto-merge` — opt-in label that enables squash auto-merge when checks pass
- `agent-dispatched` — intake already spawned a Cloud Agent for this issue
- `in-progress`, `awaiting-release`, `regression` — issue lifecycle

## 2b. Configure Cursor API secret

Add repository secret **`CURSOR_API_KEY`** (Cursor Dashboard → Integrations / API Keys).

Without it, **Issue Agent Intake** fails fast with an actionable error. The rest of the agent loop (Validate, Open PR, Approve CI, Auto Merge) still works for manually pushed `agent/**` branches.

Also enable **Settings → Actions → General → Allow GitHub Actions to create and approve pull requests** (required for Agent Open PR and Agent Approve CI).

## 3. Smoke test the loop (branch push)

```bash
git checkout main
git pull
git checkout -b agent/issue-0-smoke-test
# make a trivial docs-only change
./scripts/agent-commit-msg.sh "docs: agent workflow smoke" | git commit -F -
git push -u origin agent/issue-0-smoke-test
```

Expected:

1. **Agent Validate** passes
2. **Agent Open PR** creates a PR labeled `agent`
3. PR checks run (same as normal PRs)
4. **Agent Auto Merge** enables auto-merge when green

Close the smoke PR without merging if it was only for validation.

## 3b. Smoke test issue intake

1. Open a test issue with **Problem** and **Done when** sections (no label required)
2. Confirm **Issue Agent Intake** comments and dispatches Cloud Agent on `issues.opened`
3. Optional: add label **`agent`** or comment **`/agent`** to re-dispatch later

Requires **`CURSOR_API_KEY`** secret (step 2b).

## 4. Cursor Cloud Agent setup

- Connect GitHub in Cursor dashboard (read/write on this repo)
- Optional: store `CURSOR_API_KEY` only in Cursor secrets / password manager — never in the repo
- Point agents at `docs/AGENT_RUNBOOK.md` and `docs/AGENT_CODING_STANDARDS.md`

Cloud agents do not run the DinD acceptance harness locally; they rely on **Agent Validate** and PR CI on GitHub runners.

See `docs/AGENT_AUTONOMY.md` for the full vacation-proof CI model.
