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

- `agent` — automated agent PRs
- `agent-auto-merge` — opt-in label that enables squash auto-merge when checks pass
- `in-progress`, `awaiting-release` — issue lifecycle (if not already present)

## 3. Smoke test the loop

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

## 4. Cursor Cloud Agent setup

- Connect GitHub in Cursor dashboard (read/write on this repo)
- Optional: store `CURSOR_API_KEY` only in Cursor secrets / password manager — never in the repo
- Point agents at `docs/AGENT_RUNBOOK.md` and `docs/AGENT_CODING_STANDARDS.md`

Cloud agents do not run the DinD acceptance harness locally; they rely on **Agent Validate** and PR CI on GitHub runners.

See `docs/AGENT_AUTONOMY.md` for the full vacation-proof CI model.
