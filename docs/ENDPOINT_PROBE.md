# Endpoint Probe

Verifies Dockhand API endpoint presence against a live instance without mutating state.

## Default execution path (CI)

The probe runs automatically inside the acceptance harness on GitHub Actions via **Dockhand Compat** (every 6h on new images; nightly full inventory).

Reports are written to `docs/reports/` and synced to `main` by **Compat Reports Sync**. Maintainers do **not** need local `DOCKHAND_*` credentials for routine work. See `docs/CI_AND_RELEASE.md`.

## Debug-only local run

For investigating CI failures when you have a disposable Dockhand instance:

```bash
export DOCKHAND_ENDPOINT="http://127.0.0.1:13001"
export DOCKHAND_USERNAME="admin"
export DOCKHAND_PASSWORD="..."
DOCKHAND_PROBE_ALLOW_MUTATION=false /usr/bin/python3 scripts/endpoint-probe.py
```

Outputs:

- `docs/reports/endpoint-probe.csv`
- `docs/reports/endpoint-probe.md`

Do not commit local probe output unless it came from the same Dockhand version CI tests against; prefer waiting for **Compat Reports Sync**.

## Safety

- Default mode is non-destructive.
- `POST`/`PUT`/`DELETE` singleton endpoints are probed with `OPTIONS`.
- Parameterized mutating routes use placeholder values.
- To allow real mutating calls (disposable environments only):

```bash
DOCKHAND_PROBE_ALLOW_MUTATION=true
```

## Result categories

- `present`: endpoint responded with non-404.
- `not_present`: non-parameterized route returned `404`.
- `unverified_no_fixture`: parameterized route could not be resolved to a fixture object.
- `unexpected_404`: parameterized route still returned `404` when probed.
