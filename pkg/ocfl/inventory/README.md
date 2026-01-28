# OCFL Inventory Package

This package defines the structures and interfaces for the OCFL (Oxford Common File Layout) Inventory. The inventory is the heart of an OCFL object and contains all metadata about the object's versions, files, and fixity information.

The documentation is based on the [OCFL 1.1 Specification](ocfl11.md).

## Data Type Documentation

Detailed documentation has been created for each data type:

- [**Inventory**](docs/INVENTORY.md): The central document of the OCFL object.
- [**Version**](docs/VERSION.md): Metadata and state of a single version.
- [**Manifest**](docs/MANIFEST.md): Mapping of digests to physical files.
- [**Fixity**](docs/FIXITY.md): Optional additional checksums for integrity assurance.
- [**Types**](docs/TYPES.md): Information about `User`, `VersionNumber`, `State`, and `InventorySpec`.

## Directory Structure

- `pkg/ocfl/inventory`: Contains core interfaces and type definitions.
- `pkg/ocfl/inventory/inventoryimpl`: Contains the default implementations of these interfaces.
- `pkg/ocfl/inventory/docs`: Detailed documentation of the data types.

## Implementation

The implementation in `inventoryimpl` is modular and uses base structures (`InventoryBase`, `VersionBase`) that implement the OCFL specification.
More details on the implementation can be found in [inventoryimpl/README.md](inventoryimpl/README.md).
