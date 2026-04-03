# dockhand_git_preview_env (Data Source)

Previews Git-backed stack environment variables via `POST /api/git/preview-env`.

## Example Usage

```hcl
data "dockhand_git_preview_env" "example" {
  url          = "https://github.com/docker/awesome-compose.git"
  branch       = "master"
  compose_path = "nginx-flask-mysql/compose.yaml"
}
```

## Schema

### Required

- `compose_path` (String) Repository-relative compose file path to inspect.

### Optional

- `repository_id` (String) Existing Dockhand Git repository ID to preview. Mutually exclusive with `url`.
- `url` (String) Git clone URL to preview when `repository_id` is not used.
- `branch` (String) Git branch to preview when `url` is used.
- `credential_id` (String) Optional Dockhand Git credential ID used when `url` is set.

### Read-Only

- `id` (String) Synthetic preview ID.
- `vars_json` (String) Raw `vars` object JSON returned by Dockhand.
- `sources_json` (String) Raw `sources` object JSON returned by Dockhand.
- `variable_names` (List of String) Sorted keys from the `vars` object.
- `source_names` (List of String) Sorted keys from the `sources` object.
