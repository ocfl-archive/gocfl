# CLI helpers

set shell := ["bash", "-uc"]

# Get help
help:
  @just -l

# Fix imports
imports:
  goimports -w ./...

# Snapshot
snapshot:
  goreleaser build --snapshot --single-target --clean -f .goreleaser.yml

# Version
version:
  dist/*/gocfl --version

# Release
release:
  goreleaser release --skip=publish --clean -f .goreleaser.yml

# Single-target release
target:
  goreleaser build --single-target --clean -f .goreleaser.yml
