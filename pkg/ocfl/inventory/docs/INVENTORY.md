# Inventory

The `Inventory` is the central document of an OCFL object. It contains all metadata required to reconstruct every version of the object.

## OCFL Specification (v1.1)

According to the OCFL 1.1 specification ([3.5 Inventory](../ocfl11.md#inventory)), an inventory must have a JSON structure and include the following mandatory fields:

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
- `GetVersions() Versions`: Access to the version history ([Version](VERSION.md)).
- `GetFixity() Fixity`: Access to optional [Fixity](FIXITY.md) information.
- `AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error`: Adds a file to the inventory.
- `Finalize(inCreation bool) error`: Completes editing and prepares the inventory for storage.

## Concrete Implementation: `InventoryBase`

The default implementation is found in `pkg/ocfl/inventory/inventoryimpl/inventorybase.go`. It handles:
- Serialization to JSON (`MarshalJSON`).
- Validation of the structure against OCFL rules.
- Path management (mapping between logical and physical paths).

### Example (JSON Structure)

```json
{
  "id": "ark:/12345/bcd987",
  "type": "https://ocfl.io/1.1/spec/#inventory",
  "digestAlgorithm": "sha512",
  "head": "v1",
  "contentDirectory": "content",
  "manifest": { ... },
  "versions": { ... }
}
```

## Navigation

- [Back to README](../README.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
