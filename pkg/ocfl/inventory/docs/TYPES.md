# Helper Types and Structures

In addition to the main interfaces, there are several data types that play an important role in the OCFL specification and the implementation.

## User

A `User` represents the creator of a version. It consists of a name and an optional address (usually a mailto URI).

- **Specification**: [3.5.3.1 Version](../../../../data/specs/ocfl_1.1.md#3531-version) (section `user`)
- **Interface**: `User` (`pkg/ocfl/inventory/inventorytypes.go` / `inventoryimpl/userbase.go`)
- **Fields**:
  - `name`: Full name of the user.
  - `address`: URI (e.g., `mailto:user@example.com`).

## VersionNumber

Represents an OCFL version number (e.g., `v1`, `v2`, `v0001`).

- **Specification**: [3.3.1 Version Directories](../../../../data/specs/ocfl_1.1.md#331-version-directories)
- **Implementation**: `VersionNumber` (`pkg/ocfl/inventory/versionnumber.go`)
- **Functions**:
  - Parsing version strings.
  - Generating the next version number (`Next()`).
  - Comparing versions.

## State

The `State` of a version maps logical file paths to their digests. Unlike the manifest, which shows physical locations, the state shows how the object looks to the user in that specific version.

- **Detailed Documentation**: [STATE.md](STATE.md)
- **Specification**: [3.5.3.2 Version State](../../../../data/specs/ocfl_1.1.md#3532-version-state)
- **Interface**: `State` (`pkg/ocfl/inventory/state.go`)
- **Methods**:
  - `GetDigest(path string) string`: Returns the digest for a logical path.
  - `Paths(digest string) []string`: Returns all logical paths for a digest (aliasing/deduplication).

## InventorySpec

Defines the supported OCFL versions.

- **Values**:
  - `https://ocfl.io/1.0/spec/#inventory`
  - `https://ocfl.io/1.1/spec/#inventory`

## Navigation

- [Back to README](../README.md)
- [Go to Inventory Documentation](INVENTORY.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to State Documentation](STATE.md)
