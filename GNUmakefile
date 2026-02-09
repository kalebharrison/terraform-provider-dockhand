SHELL := /bin/bash

.PHONY: tidy build test install tf-dev-build tf-dev-plan tf-dev-apply tf-dev-destroy packages mirror

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./...

install:
	go install .

# Build a provider binary suitable for Terraform "dev_overrides".
# We keep caches inside the repo to avoid writing to global Go cache dirs.
tf-dev-build:
	@mkdir -p ./bin
	@GOCACHE=$$PWD/.cache/go-build GOMODCACHE=$$PWD/.cache/gomod \
		go build -o ./bin/terraform-provider-dockhand .
	@cp ./bin/terraform-provider-dockhand ./bin/terraform-provider-dockhand_v0.0.0

tf-dev-plan:
	@test -n "$(TF_DIR)" || (echo "Set TF_DIR=/path/to/terraform/config"; exit 2)
	@./scripts/tf-dev.sh --chdir "$(TF_DIR)" plan

tf-dev-apply:
	@test -n "$(TF_DIR)" || (echo "Set TF_DIR=/path/to/terraform/config"; exit 2)
	@./scripts/tf-dev.sh --chdir "$(TF_DIR)" apply

tf-dev-destroy:
	@test -n "$(TF_DIR)" || (echo "Set TF_DIR=/path/to/terraform/config"; exit 2)
	@./scripts/tf-dev.sh --chdir "$(TF_DIR)" destroy

packages:
	@test -n "$(VERSION)" || (echo "Set VERSION=X.Y.Z"; exit 2)
	@./scripts/build-packages.sh --version "$(VERSION)"

mirror:
	@test -n "$(VERSION)" || (echo "Set VERSION=X.Y.Z"; exit 2)
	@test -n "$(NAMESPACE)" || (echo "Set NAMESPACE=<your-namespace>"; exit 2)
	@./scripts/build-mirror.sh --version "$(VERSION)" --namespace "$(NAMESPACE)"
