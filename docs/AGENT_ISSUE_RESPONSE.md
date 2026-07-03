# Agent Issue Response Standards

**Audience:** Agents and automation — how to communicate with issue reporters from fix through release and regression.

## Goals

- Reporters understand **what changed**, **when they get a fix**, and **how to upgrade**.
- Issues are not closed with only `Fixes #N` and no context.
- If a fix did not work, reporters know to **reopen** or comment; automation picks that up.

## Lifecycle

| Stage | Who | What happens |
|-------|-----|----------------|
| Work starts | Agent | Branch `agent/issue-<n>-<slug>`; PR links `Fixes #n` |
| PR open | Agent | PR body includes **What was fixed** and **User impact** (see below) |
| PR merged | **Issue Resolution Notify** workflow | Substantive comment on the issue; label `awaiting-release` |
| Tag published | **Release Issue Notify** workflow | Comment with version, upgrade hint, reopen instructions |
| User comments on closed issue | **Issue Regression Intake** workflow | Reopens + `regression` label when feedback indicates continued problem |

## PR body (required sections)

Agent PRs must include these sections ( **Agent Open PR** seeds the template; agents must fill them before merge):

```markdown
Fixes #123

## What was fixed

- **Problem:** …
- **Cause:** …
- **Change:** … (files/resources affected)

## User impact

- **Who:** …
- **Action required:** upgrade provider / change config / none
- **Workaround until release:** … (if any)

## Validation

- `./scripts/verify.sh --quality`
- Agent Validate: (link in Automation section)
```

Do not merge agent work with empty `What was fixed` / `User impact` placeholders.

## Issue comment on merge (automation)

**Issue Resolution Notify** posts on linked issues when a fix PR merges. It includes:

- PR link and summary (from PR body sections)
- **Release status:** merged to `main`, not yet tagged
- **If the problem continues:** reopen or comment after upgrading to the release that includes the fix

Agents do not need to post this comment manually.

## Issue comment on release (automation)

**Release Issue Notify** posts when **Release Artifacts** succeeds. It includes:

- Tag (e.g. `v0.1.48`) and release URL
- What was fixed (from merged PR)
- Terraform `required_providers` version hint
- **If the problem continues:** reopen or comment — **Issue Regression Intake** will reopen the issue

## Regression / feedback on closed issues

When a user comments on a **closed** issue:

1. **Issue Regression Intake** evaluates the comment (skips bots and short “thanks” replies).
2. If it looks like continued failure, the issue is **reopened**, labeled `regression` and `agent`.
3. A bot comment acknowledges reopen and points to next steps.
4. Agents pick up reopened issues with `regression` like new work (`agent/issue-<n>-…`).

### Keywords that trigger reopen

`still broken`, `not fixed`, `reopen`, `regression`, `persists`, `doesn't work`, `does not work`, `same issue`, `issue continues`, `not resolved`, `didn't fix`, `did not fix`

Users can also write **reopen** explicitly.

### Agent follow-up on regression

1. Read the new comment and prior fix PR.
2. Branch `agent/issue-<n>-regression-<short-slug>` or reuse pattern `agent/issue-<n>-<slug>`.
3. PR body must reference the original issue and what was tried before.
4. Post-merge and release comments follow the same standards.

## Labels

| Label | Meaning |
|-------|---------|
| `in-progress` | Linked open PR |
| `awaiting-release` | Fix on `main`, not tagged yet |
| `released` | Fix included in a published tag |
| `regression` | Reporter says fix did not work; needs follow-up |
| `agent` | Safe for autonomous agent pickup |

## What not to do

- Close issues with no resolution comment (automation handles merge/release comments).
- Say only “fixed in main” without describing the change.
- Omit release version once a tag exists (Release Issue Notify handles this).
- Ignore reopened/regression issues.

## Related docs

- `docs/AGENT_INTAKE.md` — how work enters the queue
- `docs/AGENT_RUNBOOK.md` — branch and CI loop
- `.github/workflows/issue-resolution-notify.yml`
- `.github/workflows/release-issue-notify.yml`
- `.github/workflows/issue-regression-intake.yml`
