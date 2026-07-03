# Agent Coding Standards

Canonical engineering practices for humans and autonomous agents working on `terraform-provider-dockhand`.

**Read first:** `docs/AGENT_RUNBOOK.md` (workflow) and this document (how to write code).

## Principles

1. **Minimal, correct diffs** — fix the reported issue; do not refactor unrelated code.
2. **Match existing patterns** — read neighboring resources/tests before adding new ones.
3. **Fail loudly** — actionable diagnostics beat silent drift or swallowed API errors.
4. **Prove behavior** — unit tests for pure logic; acceptance tests for API contracts.
5. **Document what operators need** — schema docs, import formats, examples, and changelog entries.

## Before writing code

1. Link work to a GitHub issue (`Fixes #N` in PR body).
2. Use branch `agent/issue-<n>-<slug>` for agent work (`codex/*` for human-maintainer branches).
3. Run `./scripts/verify.sh --quality` before every push.
4. Skim `AGENTS.md` provider surface list and `docs/api-matrix.md` if touching HTTP routes.

## Go provider structure

| Area | Location | Notes |
|------|----------|-------|
| Provider wiring | `internal/provider/provider.go` | Register new resources/data sources here |
| API types | `internal/provider/client_types.go` | Shared request/response structs |
| HTTP core | `internal/provider/client.go` | `NewClient`, `doJSON*`, retries, env resolution |
| API domains | `internal/provider/client_*.go` | One file per domain (stack, git, container, …) |
| Resources | `internal/provider/resource_*.go` | Plugin Framework schema + CRUD |
| Runtime helpers | `internal/provider/runtime_helpers.go` | Shared env/enabled/job helpers |
| Unit tests | `internal/provider/*_test.go` | Fast, no Dockhand required |
| Acceptance | `internal/provider/*_tf_acc_test.go` | Requires harness / live Dockhand |

**Client rules:**

- Add new routes to the matching `client_*.go` file, not a catch-all.
- Use `httpClientWithTimeout` for long operations (image pull, git deploy stream).
- Pass `force=true` on destructive deletes when the API expects it (stacks, volumes, images).
- Prefer typed structs over `map[string]any` for stable responses.

## Resource implementation checklist

When adding or materially changing a resource or data source:

1. **Schema** — required/optional/computed, `Sensitive` for secrets, `RequiresReplace` for immutable fields.
2. **Plan modifiers** — `UseStateForUnknown` on stable computed IDs where appropriate.
3. **CRUD** — create/update/read/delete map to documented Dockhand endpoints.
4. **Import** — implement `ImportState` when the resource is adoptable; document ID format in `docs/resources/<name>.md`.
5. **Drift** — reconcile runtime state on read when Terraform should be source of truth (`enabled`, job status, etc.).
6. **One-shot flags** — deploy/trigger booleans must not re-fire every apply (reset after use or use `*_action` resources).
7. **Env scoping** — resolve `default_env`, persist resolved `env` in state, error clearly when env is required.

## Testing requirements

### Unit tests (`go test ./...`)

- Pure helpers, payload builders, parsers, retry logic.
- Table-driven where multiple cases exist.

### Acceptance manifest (`testdata/acceptance_manifest.json`)

Every provider resource and data source **must** appear with:

- `mode`: `stateful`, `action`, or `data_source`
- `operations`: lifecycle ops the resource supports
- `acceptance_test_regex`: explicit `TestAcc...` pattern (never bare `TestAcc`)

`TestAcceptanceManifestCoverage` also verifies matching acceptance sources exercise declared operations (`ImportState`, `CheckDestroy`, multi-step updates). Known gaps live in `manifestOperationExemptions` — **shrink this list, do not grow it without justification**.

### PR CI subset (`testdata/acceptance_pr_ci.json`)

Add the exact `TestAcc...` function name when a suite should run on every PR and on `agent/**` pushes.

### Acceptance test patterns

```go
resource.Test(t, resource.TestCase{
    CheckDestroy: testAccCheckThingDestroyed(...), // stateful resources
    Steps: []resource.TestStep{
        { Config: ..., Check: ... },              // create
        { Config: ..., Check: ... },              // update
        {
            ResourceName:      "dockhand_thing.test",
            ImportState:       true,
            ImportStateVerify: true,
            ImportStateVerifyIgnore: []string{"secret_field"},
        },
    },
})
```

- **Action resources** — at least one `TestStep`; destroy is implicit at test end.
- **Skip guards** — use `t.Skip` with clear env var names; export fixtures from `scripts/run-acceptance-harness.sh` when possible.
- **Never** silently pass when fixtures are missing unless the test is explicitly optional and documented.

## Documentation requirements

| Change | Update |
|--------|--------|
| New/changed resource | `docs/resources/<name>.md` + `examples/` |
| New/changed data source | `docs/data-sources/<name>.md` |
| Import support | `## Import` section with exact ID format |
| API route added/changed | `scripts/endpoint-probe.py`, `docs/api-matrix.md`; CI refreshes reports (see `docs/AGENT_AUTONOMY.md`) |
| User-visible behavior | `CHANGELOG.md` Unreleased section |
| Agent/process change | relevant `docs/AGENT_*.md` |

Run `scripts/check-doc-example-coverage.py` via `./scripts/verify.sh --quality`.

## API & drift tooling

After client route changes:

1. Add endpoint to `scripts/endpoint-probe.py` `ENDPOINTS`.
2. Ensure `scripts/api-drift-gate.py` prefixes cover the path family.
3. Merge via CI — **Acceptance Full** / **Release Watch** regenerate reports; **Compat Reports Sync** opens a baseline PR.

Do not require a local `./scripts/verify.sh --endpoint-probe` run.

## Quality gates

**Routine (agents):** push → CI.

**Optional local (no Dockhand):**

```bash
./scripts/verify.sh --quality
```

**CI must pass:** `Lint, Test, Build`, `Dockhand + DinD Acceptance Tests`, `dependency-review`, and agent workflows for `agent/**`.

**Dockhand-dependent checks** (endpoint probe, full acceptance, drift audits) run only on GitHub runners — see `docs/AGENT_AUTONOMY.md`.

Debug-only local commands (not merge gates):

```bash
./scripts/verify.sh --acceptance --test-regex 'TestAccYourSuite'
./scripts/verify.sh --endpoint-probe   # needs DOCKHAND_* ; use only when debugging CI
```

## Commits and pull requests

- **Agent commits** must include `Co-authored-by: Cursor Agent <noreply@cursor.com>` (use `./scripts/agent-commit-msg.sh`).
- **Conventional commits** — `fix(provider):`, `feat:`, `docs:`, `test:`, `chore:`.
- **PRs** — opened by Agent Open PR for `agent/**`; humans use normal PR flow with `Fixes #N`.
- **Issue responses** — follow `docs/AGENT_ISSUE_RESPONSE.md`; fill PR resolution sections before merge.
- **Do not** merge, tag releases, or push secrets manually — **Agent Auto Merge** and **Agent Release Tag** handle routine merge and release publish

## Anti-patterns (never)

- Bare `TestAcc` regex in manifest
- Secrets in code, examples, or commit messages
- New API methods only in `client.go` when a domain file exists
- Persistent `deploy_now` / `force_redeploy` true in state across applies
- Missing `Sensitive` on password/token/env/job JSON fields
- Skipping manifest/docs/examples updates when adding provider surface
- Broad refactors mixed into issue-scoped fixes
- Amending pushed commits or force-pushing `main`

## Related docs

- `docs/AGENT_RUNBOOK.md` — branch/CI loop
- `docs/AGENT_IDENTITY.md` — attribution model
- `docs/AGENT_INTAKE.md` — how work enters the agent queue
- `docs/AGENT_AUTONOMY.md` — CI-first model (no maintainer machine)
- `docs/AGENT_ISSUE_RESPONSE.md` — issue comments, release notices, regression handling
- `docs/AGENT_DEPLOYMENT.md` — one-time rollout checklist
- `docs/AGENT_SWEEP.md` — readiness tracker
- `docs/MAINTENANCE_PLAYBOOK.md` — releases and harness
- `CONTRIBUTING.md` — human contributor entry point
