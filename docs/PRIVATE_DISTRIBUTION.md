# Private Distribution

Goal: use this provider "like normal" (via `init`) without publishing a public provider.

## Option A: Filesystem Mirror (Recommended)

This works for both Terraform and OpenTofu and is the lowest-ops option. You build provider zip packages and place them into a shared directory (SMB/NFS/etc). Consumers point their CLI configuration at that directory.

### 1) Build Packages And Mirror

From the repo root:

```bash
make packages VERSION=0.1.0
make mirror VERSION=0.1.0 NAMESPACE=kalebharrison
```

This writes a mirror directory at `./mirror` using the packed mirror layout.

Alternative: download zip artifacts from a GitHub Release (tag `v0.1.0`) and then place them into your mirror.

### 2) Consumer CLI Config

Terraform: `~/.terraformrc` (or set `TF_CLI_CONFIG_FILE`).

OpenTofu: `~/.tofurc` (or set `TOFU_CLI_CONFIG_FILE`).

Example:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/path/to/shared/mirror"
    include = ["registry.terraform.io/kalebharrison/dockhand"]
  }
  direct {
    exclude = ["registry.terraform.io/kalebharrison/dockhand"]
  }
}
```

Then your Terraform/OpenTofu configuration can use:

```hcl
terraform {
  required_providers {
    dockhand = {
      source  = "kalebharrison/dockhand"
      version = "0.1.0"
    }
  }
}
```

## Option B: Private Terraform Cloud / Terraform Enterprise Registry

This gives the most "normal" experience for teams, but has more setup and requires signing (GPG). Treat this as the longer-term goal if you want a true private registry with UI, version management, and access controls.
