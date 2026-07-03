# Agent Runbook

Operational guide for autonomous agents managing `terraform-provider-dockhand`.

**Required reading:** `docs/AGENT_CODING_STANDARDS.md` (engineering practices), `docs/AGENT_IDENTITY.md` (attribution).

For how issues are picked up, see `docs/AGENT_INTAKE.md`.

## Branch and issue contract

1. Pick or create a GitHub issue for the work.
2. Create a branch named:

```text
agent/issue-<number>-<short-slug>
```

Example: `agent/issue-42-fix-environment-agent-token`

3. Push to `origin`. CI runs **Agent Validate** on every `agent/**` push.
4. When validation succeeds, **Agent Open PR** creates or updates a pull request with `Fixes #<number>`.
5. **Agent Auto Merge** enables squash auto-merge when checks go green, the PR has the `agent-auto-merge` label, and the PR body includes filled **What was fixed** / **User impact** sections.

Human PR policy checks (`Fixes #`, conventional title) are skipped for `agent/**` branches.

## Issue communication

Follow `docs/AGENT_ISSUE_RESPONSE.md`:

- Fill **What was fixed** and **User impact** in the PR before merge (required for auto-merge).
- Do not post bare “fixed” comments; **Issue Resolution Notify** and **Release Issue Notify** handle issue threads.
- Pick up `regression`-labeled reopened issues like new work.

## Commit transparency

Every commit on `agent/**` branches must include this trailer in the message body:

```text
Co-authored-by: Cursor Agent <noreply@cursor.com>
```

Build a message with:

```bash
./scripts/agent-commit-msg.sh "fix(provider): short summary"
```

Or source helpers and commit:

```bash
source ./scripts/agent-git-env.sh
agent_commit "fix(provider): short summary"
```

See `docs/AGENT_IDENTITY.md` for the full attribution model.

## Validation commands

**CI is the Dockhand gate.** Agents push and rely on GitHub Actions (see `docs/AGENT_AUTONOMY.md`).

Optional before push (no Dockhand required):

```bash
./scripts/verify.sh --quality
/usr/bin/python3 scripts/acceptance-pr-ci-regex.py
```

Do **not** require local `--endpoint-probe` or `--acceptance` when CI is available. On GitHub, `agent/**` pushes run the acceptance subset via `agent-validate.yml`.

## Provider surface changes

When adding or changing a resource or data source, follow the full checklist in `docs/AGENT_CODING_STANDARDS.md`. At minimum, update all of:

1. `internal/provider/provider.go`
2. Client/API mapping
3. Acceptance test (`internal/provider/*_tf_acc_test.go`)
4. `internal/provider/testdata/acceptance_manifest.json`
5. Docs (`docs/resources/` or `docs/data-sources/`)
6. Example (`examples/...`)
7. API matrix docs when endpoint status changes

If the new acceptance suite should run on PRs, add the exact `TestAcc...` function name to:

`internal/provider/testdata/acceptance_pr_ci.json`

## CI map

| Workflow | When | Purpose |
|----------|------|---------|
| `go-ci.yml` | PR + push to `main` | fmt, tidy, docs/examples, vet, golangci-lint, staticcheck, shellcheck, unit tests, build |
| `acceptance-ci.yml` | PR | Dockhand + DinD + Hawser targeted acceptance |
| `agent-validate.yml` | push to `agent/**` | Agent pre-PR validation |
| `agent-open-pr.yml` | after successful Agent Validate | Opens/updates PR as `github-actions[bot]` |
| `agent-approve-ci.yml` | agent PR opened/updated | Approves pending workflow runs |
| `agent-auto-merge.yml` | agent PR events | Enables auto-merge |
| `issue-agent-intake.yml` | issue labeled `agent`, `/agent` comment | Dispatches Cursor Cloud Agent |
| `acceptance-full.yml` | nightly | Full `TestAcc` + drift audits |
| `dockhand-release-watch.yml` | every 6h | New Dockhand image compatibility |
| `compat-reports-sync.yml` | after full/release-watch success | PR to refresh `docs/reports/` baselines |
| `issue-resolution-notify.yml` | fix PR merged | Substantive comment on linked issues |
| `issue-regression-intake.yml` | comment on closed issue | Reopen + `regression` when fix failed |
| `release-issue-notify.yml` | release published | Version + upgrade comment on issues |

## Failure handling

1. Read the failed job log in GitHub Actions.
2. For acceptance failures, download the `*-logs-*` artifact when present.
3. Fix narrowly, push to the same `agent/issue-*` branch, wait for **Agent Validate** again.
4. Do not open a new branch for the same issue unless the old one is abandoned.

## Never

- Commit secrets, `.env`, Terraform state, or local override files
- Merge or tag releases without explicit maintainer instruction
- Use bare `TestAcc` manifest mappings
- Bypass manifest/docs/examples parity

## Maintainer-only

- Signed release tags (`vX.Y.Z`) and release artifact validation
- **Release lens review** before every tag (`docs/testing/release-lens-review.md`)
- Rotating production Dockhand credentials
- Changing branch protection or required checks

## Related docs

- `docs/AGENT_AUTONOMY.md` — vacation-proof CI model (no maintainer machine)
- `docs/AGENT_DEPLOYMENT.md` — one-time agent CI rollout
