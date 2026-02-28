# Inventory Implementation (inventoryimpl)

This package provides the default implementations for the interfaces defined in the parent package.

## Core Implementations

The implementations strictly follow the OCFL specification and use the interfaces defined in `inventory`. Detailed descriptions of the data types can be found in the respective documentation files:

- [**InventoryBase**](../docs/INVENTORY.md#concrete-implementation-inventorybase): Central class for `inventory.json`.
- [**VersionBase**](../docs/VERSION.md#concrete-implementation-versionbase): Management of version metadata.
- [**ManifestBase**](../docs/MANIFEST.md#concrete-implementation-manifestbase): Management of physical file mapping.
- [**FixityBase**](../docs/FIXITY.md#concrete-implementation-fixitybase): Optional integrity check.
- [**StateBase**](../docs/STATE.md#concrete-implementation-statebase): Logical view of the files in a version.

## Implementation Features

### Serialization
The `MarshalJSON()` method ensures compliant output according to the OCFL specification. It takes into account specific requirements for date formats (RFC3339) and path structures.

### Validation
The implementations include integrated validation logic to ensure the inventory remains consistent (e.g., checking for duplicate paths, correctness of digests).

### Management of File Operations
Methods such as `AddFile`, `DeleteFile` (logical), and `RenameFile` allow for the manipulation of the object state across versions, with deduplication in the manifest automatically handled.

## Usage

Instances of these implementations are typically created via a factory that configures the correct versions (1.0 vs 1.1) and algorithms.

## Navigation

- [Back to main Inventory README](../README.md)
- [Back to Inventory Data Type Documentation](../docs/INVENTORY.md)
