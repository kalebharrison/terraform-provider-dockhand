# Local Development (Private)

This provider can be developed and tested locally without publishing to a registry by using Terraform `dev_overrides`.

## Build Provider Binary

From the repo root:

```bash
make tf-dev-build
```

This produces:

- `bin/terraform-provider-dockhand`
- `bin/terraform-provider-dockhand_v0.0.0`

## Run Terraform With Dev Overrides

In your Terraform configuration directory, run Terraform through the helper script:

```bash
REPO="/path/to/terraform-provider-dockhand"
"$REPO/scripts/tf-dev.sh" plan
"$REPO/scripts/tf-dev.sh" apply
```

The script writes a local CLI config file `.terraformrc.dockhand-dev` and sets `TF_CLI_CONFIG_FILE` for that invocation only. If `tofu` is installed, it will use `tofu`; otherwise it uses `terraform`.

## Dockhand Auth

Prefer env vars so credentials are not stored in `.tf` files:

```bash
export DOCKHAND_ENDPOINT="http://dockhand.example.internal:13001"
export DOCKHAND_USERNAME="your-username"
export DOCKHAND_PASSWORD="your-password"
export DOCKHAND_DEFAULT_ENV="1"

# Optional fresh-install bootstrap mode (no auth configured yet):
# export DOCKHAND_ALLOW_UNAUTHENTICATED="true"
```
