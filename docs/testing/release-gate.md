# Release Gate

Canonical docs: `docs/CI_AND_RELEASE.md`.

Programmatic gate: `scripts/release_gate_check.py`

| Mode | Meaning |
|------|---------|
| `status` | CI gates pass? |
| `tag` | Ready to publish (CI green + draft + release work) |
| `lens` | Deprecated alias; same readiness signal as historical lens dispatch |

Required workflows on `main`: **Go CI**, **Security**, **Dockhand Compat** (validated run).

Publish via **Actions → Release → Run workflow** (`.github/workflows/release.yml`).
