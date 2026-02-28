# Inventory

The `Inventory` is the central document of an OCFL object. It contains all metadata required to reconstruct every version of the object.

## Modular Structure

The inventory is split into several structural modules. Each module is documented in its own file:

- [**Versions**](VERSION.md): Management of the version history and version states.
- [**Manifest**](MANIFEST.md): Mapping of content digests to physical file paths.
- [**Fixity**](FIXITY.md): Optional additional fixity information.

## Instantiation via Factory

Direct instantiation of implementation classes (like `InventoryBase`) is discouraged. Instead, use the `pkg/ocfl/factory` package to create an inventory. This ensures that the correct implementation for the target OCFL version is used.

For more details, see the [Factory documentation](../../factory/README.md).

### Using `With...()` methods

When using the factory to build an inventory, you must link the constituent components using the appropriate `With...()` methods. These methods are designed for fluid composition:

```go
// 1. Create components via the factory
inv := f.NewInventory(ctx)
manifest := f.NewManifest(ctx)
versions := f.NewVersions(ctx)
fixity := f.NewFixity(ctx)

// 2. Link components to the inventory
inv.WithManifest(manifest).
    WithVersions(versions).
    WithFixity(fixity).
    WithID("ark:/12345/bcd987").
    WithDigestAlgorithm(checksum.DigestSHA512)
```

## OCFL Specification

The Inventory is the central metadata file for an OCFL object, as described in the specification.

- **Specification**: [3.5 Inventory](../../../../data/specs/ocfl_1.1.md#35-inventory)

According to the OCFL specification, an inventory must have a JSON structure and include the following mandatory fields:

- `id`: Unique identifier of the object (should be a URI).
- `type`: Type URI corresponding to the OCFL version (e.g., `https://ocfl.io/1.1/spec/#inventory`).
- `digestAlgorithm`: The algorithm used for the content-addressing scheme (e.g., `sha512`).
- `head`: The version number of the most recent version (e.g., `v1`).
- `manifest`: A mapping of digests to physical paths within the object.
- `versions`: A list of version objects.

Optionally, a `contentDirectory` can be defined (default is `content`).

## Implementation: `Inventory` Interface

In the code, the inventory is represented by the `Inventory` interface (`pkg/ocfl/inventory/inventory.go`).

### Important Methods

- `GetID() string`: Returns the object's ID.
- `GetDigestAlgorithm() checksum.DigestAlgorithm`: Returns the used digest algorithm.
- `GetHead() *VersionNumber`: Returns the current head version.
- `GetManifest() Manifest`: Access to the [Manifest](MANIFEST.md).
- `GetVersions() Versions`: Access to the version history ([Versions](VERSION.md)).
- `GetFixity() Fixity`: Access to optional [Fixity](FIXITY.md) information.
- `AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error`: Adds a file to the inventory.
- `Finalize(inCreation bool) error`: Completes editing and prepares the inventory for storage.

## Concrete Implementation: `InventoryBase`

The default implementation is found in `pkg/ocfl/inventory/inventoryimpl/inventorybase.go`. It handles serialization to JSON, validation against OCFL rules, and path management.

## Navigation

- [Back to README](../README.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
