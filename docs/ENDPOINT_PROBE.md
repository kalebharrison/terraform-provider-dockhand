# Endpoint Probe

Use this to verify Dockhand API endpoint presence against a live instance without mutating state.

## Command

From repository root:

```bash
source terraform/dockhand/test/env.sh
DOCKHAND_PROBE_ALLOW_MUTATION=false /usr/bin/python3 scripts/endpoint-probe.py
```

Outputs:

- `docs/reports/endpoint-probe.csv`
- `docs/reports/endpoint-probe.md`

## Safety

- Default mode is non-destructive.
- `POST`/`PUT`/`DELETE` singleton endpoints are probed with `OPTIONS`.
- Parameterized mutating routes use placeholder values.
- To allow real mutating calls (for controlled acceptance checks), set:

```bash
DOCKHAND_PROBE_ALLOW_MUTATION=true
```

Use mutation mode only in a disposable test environment.

## Result Categories

- `present`: endpoint responded with non-404.
- `not_present`: non-parameterized route returned `404`.
- `unverified_no_fixture`: parameterized route could not be resolved to a fixture object.
- `unexpected_404`: parameterized route still returned `404` when probed.
