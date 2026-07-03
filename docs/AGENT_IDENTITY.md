# Agent Identity and Transparency

This repository uses an **agent-managed workflow**: Cursor Cloud Agents implement changes, GitHub Actions validate and open pull requests, and maintainers retain merge/release control.

We are transparent about that. A separate GitHub machine user is **not** required.

## How attribution works

| Step | Who it shows as | How we note agent involvement |
|------|-----------------|-------------------------------|
| Branch commits | Maintainer (Cursor uses your GitHub login) | `Co-authored-by: Cursor Agent <noreply@cursor.com>` in every agent commit |
| Pull request | `github-actions[bot]` via **Agent Open PR** | `agent` label + PR body banner |
| CI validation | `github-actions[bot]` | **Agent Validate** workflow |
| Merge to `main` | Squash commit on `main` | Co-author trailer preserved when present on squashed commits |

## Required commit trailer

Every commit on `agent/**` branches must end with:

```text
Co-authored-by: Cursor Agent <noreply@cursor.com>
```

Append it in the commit message body (after a blank line following the subject).

Example:

```text
fix(provider): preserve agent token on environment read

Co-authored-by: Cursor Agent <noreply@cursor.com>
```

Helper:

```bash
./scripts/agent-commit-msg.sh "fix(provider): summary here"
```

## Pull request flow

Do **not** open agent pull requests manually when avoidable.

1. Push to `agent/issue-<number>-<slug>`
2. Wait for **Agent Validate**
3. **Agent Open PR** creates or updates the PR as `github-actions[bot]`
4. **Agent Auto Merge** enables squash auto-merge when checks pass

## Optional dedicated identity (not default)

If you later want commits to show a non-maintainer GitHub user, create a machine user and Cursor account for it. That is optional overhead for this project and is not required for transparency.

## Bot and branch policy exemptions

- `agent/**` branches skip human PR title and issue-link enforcement
- `Bot` users skip issue-link enforcement
- Agent PRs still include `Fixes #...` for traceability (**Agent Open PR** adds this automatically)
