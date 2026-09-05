# Agent context

Start here for AI agents (IDE chat) working in this repo.

## Repository

- Open this repo at its root, not a parent folder.
- Do not commit secrets, local Terraform state, or `.codex/` machine config.
- Match existing Go and Terraform provider conventions in this repo.

## Cursor

- **Rules:** `.cursor/rules/repo-basics.mdc`
- **Ignore:** `.cursorignore` — excludes build artifacts, Terraform cache, secrets from indexing

## Operating model (two loops)

Cursor IDE chat helps write code. **GitHub Actions** owns detection, validation, and release.

1. **Dockhand Compat** — watch Dockhand versions; full API inventory; open issues when drift/compat fails
2. **People issues** — fix in a PR (you or Cursor), link `Fixes #N`
3. **Validate** — Go CI + Acceptance CI on PRs; Dockhand Compat for full inventory
4. **Release** — when green and there is release work: `gh workflow run release.yml` (manual dispatch)

There is **no** Cloud Agent intake, auto-merge factory, or lens-log release gate.

### Active workflows (9)

| Workflow | Role |
|----------|------|
| Go CI | Lint, unit, build, shellcheck, actionlint |
| Acceptance CI | PR Dockhand+DinD subset |
| Dockhand Compat | Watch Dockhand + full TestAcc + API probe/drift |
| Compat Reports Sync | Commit probe baselines after green Compat |
| Security | Gitleaks, CodeQL, govulncheck, dependency-review |
| PR Policy | Conventional title + linked issue |
| Release | Drafter on `main` push; tag/publish on `workflow_dispatch` |
| Release Artifacts | Signed zips (called by Release) |
| Maintenance | Stale, hygiene, GPG/settings smoke |

### Quick loop

1. Branch `agent/issue-<n>-<slug>` or a normal feature branch
2. Implement + `./scripts/verify.sh --quality`
3. Open PR with `Fixes #<n>`
4. Green CI → merge
5. When ready: **Actions → Release → Run workflow**

Optional: `Co-authored-by: Cursor Agent <noreply@cursor.com>` on agent-assisted commits (`./scripts/agent-commit-msg.sh`).

## Docs

- `docs/CI_AND_RELEASE.md` — workflows and release gate
- `docs/AGENT_CODING_STANDARDS.md` — how to write provider code
- `docs/MAINTENANCE_PLAYBOOK.md` — maintainer ops

## Validation

```bash
./scripts/verify.sh --quality
```

Dockhand-dependent acceptance and endpoint probes run on GitHub Actions.
