# Release Lens Review

Automated release-tier lens review before **Agent Release Tag** publishes `vX.Y.Z`.

Playbook detail for each lens: `docs/AGENT_REVIEW_LENSES.md`  
Log output: `docs/reports/agent-review-log.md`  
Gate script: `scripts/release_gate_check.py`

Depth is **proportional to what changed** — not a fixed time per lens. See **Sweep depth** in `AGENT_REVIEW_LENSES.md`.

## When (automated)

1. Fixes are on `main` with `awaiting-release` issues.
2. CI release gates pass (`docs/testing/release-gate.md`).
3. **Agent Release Orchestrate** opens `release: prepare vX.Y.Z`.
4. **Issue Agent Intake** dispatches a Cloud Agent with the tier lens set.
5. Agent logs sweeps and a **Clear to tag** verdict.
6. **Agent Release Tag** publishes the signed tag when the verdict is on `main`.

## Release tier

Tier is computed automatically from the draft release version vs the latest published tag (`scripts/release_semver.py`).

| Tier | When | Lenses |
|------|------|--------|
| **Patch** | Bugfix, narrow change, docs/CI-only | **Core 5** (below) |
| **Minor / major** | New/changed resources, schema, or API client | **All 11** (order below) |

### Core 5 (patch releases)

1. API compatibility  
2. Terraform schema & state  
3. Acceptance & regression  
4. Security engineer  
5. Release & upgrade  

### Order (all 11 — minor/major)

| Step | Lens |
|------|------|
| 1 | API compatibility |
| 2 | Terraform schema & state |
| 3 | Dockhand domain / runtime |
| 4 | Async & long-running operations |
| 5 | Acceptance & regression |
| 6 | Security engineer |
| 7 | Ops / SRE |
| 8 | Senior developer |
| 9 | GitOps / IaC practitioner |
| 10 | Entry-level developer |
| 11 | Release & upgrade |

## Review log header

```markdown
## Release vX.Y.Z — lens review

- Tier: patch | minor/major
- Started: YYYY-MM-DD
- Base commit: <sha>
- CI gates: pass (release_gate_check.py)
- Status: in progress | blocked | clear to tag
```

Append one `### YYYY-MM-DD — <lens>` section per step (format in `docs/AGENT_REVIEW_LENSES.md`).

Close with:

```markdown
### Release vX.Y.Z — verdict

- **Clear to tag:** yes | no
- **Blocking findings:** <none or list>
- **Deferred medium/low:** <issue links>
```

**Agent Release Tag** requires **Clear to tag: yes** (`scripts/release_verdict.py`).

## Gate rules

| Severity | Before tag |
|----------|------------|
| **High** | Must fix or block release (verdict stays **no**) |
| **Medium** | Fix or document deferral in release notes + issue |
| **Low** | Issue filed; may ship |

## After lenses clear

**Agent Release Tag** pushes the signed tag; then see `docs/testing/release-gate.md` for artifact validation.
