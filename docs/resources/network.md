# dockhand_network

Manages a Dockhand network via `/api/networks`.

## Example Usage

```hcl
resource "dockhand_network" "shared" {
  name   = "tf-shared-net"
  driver = "bridge"
  env    = "1"
  internal   = false
  attachable = true
  options = {
    com.docker.network.bridge.enable_icc = "true"
  }
}
```

## Schema

### Required

- `name` (String) Network name.

### Optional

- `driver` (String) Network driver. Defaults to `bridge`.
- `env` (String) Optional environment ID. If omitted, provider `default_env` is used.
- `internal` (Boolean) Whether the network is internal.
- `attachable` (Boolean) Whether the network is attachable.
- `options` (Map of String) Driver option map.

### Read-Only

- `id` (String) Network ID.
- `internal` (Boolean) Whether the network is internal.
- `attachable` (Boolean) Whether the network is attachable.
- `options` (Map of String) Effective options.
- `labels` (Map of String) Network labels from inspect.
- `scope` (String) Network scope.
- `created_at` (String) Creation timestamp.

## Import

```bash
terraform import dockhand_network.shared <network-id>
```
