# CLI helpers

set shell := ["bash", "-uc"]

# get help
help:
  @just -l

# fix imports
imports:
  goimports -w ./...

# run tests
test:
  go test ./...

# run tests with coverage (requires version alignment)
test-cov:
  go test -coverprofile= ./...

# clear test cache
rm-test-cache:
  go clean -testcache

# snapshot
snapshot:
  goreleaser build --snapshot --single-target --clean -f .goreleaser.yml

# version
version:
  dist/*/gocfl --version

# release
release:
  goreleaser release --skip=publish --clean -f .goreleaser.yml

# single-target release
target:
  goreleaser build --single-target --clean -f .goreleaser.yml

# docs
docs:
  godoc -http=localhost:6060

# init-submodules
init-sub:
  git submodule update --init --recursive.
