# Release Lens Review

Required agent review pass **before** creating a signed `vX.Y.Z` tag.

Playbook detail for each lens: `docs/AGENT_REVIEW_LENSES.md`  
Log output: `docs/reports/agent-review-log.md`

Depth is **proportional to what changed** — not a fixed time per lens. See **Sweep depth** in `AGENT_REVIEW_LENSES.md`.

## When

After `main` is green and release CI gates pass (`docs/testing/release-gate.md`), **before** `git tag -s vX.Y.Z`.

Maintainer prompt examples:

- `prepare release v0.2.0`
- `run release lens review for v0.2.0`

## Release tier

Pick the tier from what is shipping. Maintainer can override (e.g. request full 11 on a patch).

| Tier | When | Lenses |
|------|------|--------|
| **Patch** | Bugfix, narrow change, docs/CI-only | **Core 5** (below) |
| **Minor / major** | New/changed resources, schema, or API client | **All 11** (order below) |
| **First release after agent CI** | Baseline audit | **All 11** at least standard depth |

### Core 5 (patch releases)

1. API compatibility  
2. Terraform schema & state  
3. Acceptance & regression  
4. Security engineer  
5. Release & upgrade  

Skip the other six unless the patch touches their area (then run those at **standard** depth too).

## Order (all 11 — minor/major / baseline)

| Step | Lens | Why this order |
|------|------|----------------|
| 1 | API compatibility | Confirm Dockhand contract before judging provider code |
| 2 | Terraform schema & state | Core correctness for Terraform users |
| 3 | Dockhand domain / runtime | Operator-facing behavior |
| 4 | Async & long-running operations | Actions and jobs |
| 5 | Acceptance & regression | Proof tests match reality |
| 6 | Security engineer | Secrets, auth, supply chain |
| 7 | Ops / SRE | CI and harness health |
| 8 | Senior developer | Go structure and maintainability |
| 9 | GitOps / IaC practitioner | Docs, examples, HCL UX |
| 10 | Entry-level developer | Onboarding and contributor path |
| 11 | Release & upgrade | This release's notes, version pins, gate checklist |

Steps may span multiple agent sessions. Track progress in the review log header.

## Review log header

Start each release pass with:

```markdown
## Release vX.Y.Z — lens review

- Tier: patch | minor/major
- Started: YYYY-MM-DD
- Base commit: <sha>
- CI gates: <links or pass/fail summary>
- Status: in progress | blocked | clear to tag

```

Append one `### YYYY-MM-DD — <lens>` section per step (format in `AGENT_REVIEW_LENSES.md`).

Close with:

```markdown
### Release vX.Y.Z — verdict

- **Clear to tag:** yes | no
- **Blocking findings:** <none or list>
- **Deferred medium/low:** <issue links>
```

## Gate rules

| Severity | Before tag |
|----------|------------|
| **High** | Must fix or explicitly abort release |
| **Medium** | Fix or document deferral in release notes + issue |
| **Low** | Issue filed; may ship |

## After lenses clear

Continue with `docs/testing/release-gate.md`:

1. `git tag -s vX.Y.Z -m "Release vX.Y.Z"`
2. `git push origin vX.Y.Z`
3. Wait for release artifacts workflow
4. `./scripts/release-test.sh X.Y.Z`
