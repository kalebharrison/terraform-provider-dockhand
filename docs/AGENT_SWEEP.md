# Agent Readiness Sweep

Living checklist for the agent-management rollout. Update as items complete.

## Done

- [x] Agent CI: `agent-validate`, `agent-open-pr`, `agent-auto-merge`
- [x] Merge `quality-ci` into `go-ci` (shellcheck included)
- [x] Manifest-driven PR acceptance (`acceptance_pr_ci.json` + script + tests)
- [x] Manifest operations enforcement in acceptance tests
- [x] Composite `setup-terraform` action
- [x] Acceptance failure log artifacts
- [x] `agent/**` exemptions for PR policy workflows
- [x] `agent` label + stale-bot exemption
- [x] `dependency-review` always on
- [x] Hygiene: `.env.example`, `.gitignore`, CONTRIBUTING security
- [x] Transparency: `Co-authored-by` helpers + docs
- [x] `.cursor/rules/agent-workflow.mdc`
- [x] `scripts/test-agent-helpers.sh` in Go CI + Agent Validate
- [x] Co-author trailer enforced on `agent/**` pushes (Agent Validate)
- [x] Docs: RUNBOOK, IDENTITY, DEPLOYMENT, INTAKE, CODING_STANDARDS, README, MAINTENANCE_PLAYBOOK
- [x] Client split (`client_*.go`), endpoint probe expansion, import doc sweep
- [x] Issue communication: Resolution Notify, Release Issue Notify (expanded), Regression Intake (`docs/AGENT_ISSUE_RESPONSE.md`)
- [x] **Commit and push** agent CI + lens fixes to `main`
- [x] **Branch protection sync** per `docs/AGENT_DEPLOYMENT.md` (removed `Vet, Lint, Staticcheck`; added `dependency-review`)
- [x] **Smoke test** `agent/issue-0-smoke-test`: Agent Validate green, Agent Open PR opened #125, closed without merge
- [x] GitHub Actions setting: allow workflows to create pull requests

## Next

- [ ] Confirm **Agent Auto Merge** on first real `agent/issue-<n>-*` PR (with filled resolution sections + `agent-auto-merge` label)
- [ ] First nightly **Acceptance Full** → Compat Reports Sync PR

## Backlog (optional polish)

- [ ] Path filter: skip `acceptance-ci` for docs-only PRs (careful — easy to miss provider doc drift)
- [ ] `.cursor/environment.json` for Cloud Agent install hints (Go, Terraform)
- [ ] Post-merge issue: document agent workflow in GitHub wiki / discussions
- [ ] Add `agent-commit-msg.sh --check` to a pre-push hook template (optional)
- [ ] Monitor first 3 agent PRs for CI time / flake rate
- [ ] Decide policy for bot-opened PRs: approve-and-run vs auto-run same-repo agent branches

## Focused review lenses

**Release:** lens review before every tag — tiered (`docs/testing/release-lens-review.md`).

**Ad hoc:** single lens when a trigger in `docs/AGENT_REVIEW_LENSES.md` applies. Log: `docs/reports/agent-review-log.md`.

## Explicit non-goals

- Separate GitHub machine user (transparency via co-author + bot PR is enough)
- Running full DinD acceptance inside Cursor Cloud Agent VM (use GitHub Actions)
- Auto-tagging releases without maintainer action
