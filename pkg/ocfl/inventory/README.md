# OCFL Inventory Package

This package defines the structures and interfaces for the OCFL (Oxford Common File Layout) Inventory. The inventory is the heart of an OCFL object and contains all metadata about the object's versions, files, and fixity information.

## Overview

The inventory structure is divided into several structural modules to maintain clarity and follow the OCFL specification. For a visual overview, see the [**Architecture Diagram**](docs/architecture.md).

- **Inventory**: The central management unit.
- **Versions**: Management of the version history and the state of each version.
- **Manifest**: Mapping of content digests to physical file paths.
- **Fixity**: Optional additional fixity information for physical files.

## Instantiation and Factory

The recommended way to create inventory components is through the `pkg/ocfl/factory` module. The factory ensures that components are created correctly according to the desired OCFL version (1.0, 1.1, etc.).

When using the factory to create an inventory, the individual components (Manifest, Versions, Fixity) must be linked using the `With...()` methods (e.g., `WithManifest()`, `WithVersions()`, `WithFixity()`).

For more details on instantiation, see the [Factory documentation](../factory/README.md).

## GoDoc Documentation

This package uses standard GoDoc comments. You can view the detailed API documentation by running:
```bash
go doc -all pkg/ocfl/inventory
```

## Data Type Documentation (Detailed)

Detailed documentation is available for each module in the [**docs**](docs/README.md) directory:

- [**Architecture**](docs/architecture.md): Visual representation of the interface hierarchy.
- [**Inventory**](docs/INVENTORY.md): The central document of the OCFL object.
- [**Version**](docs/VERSION.md): Metadata and state of a single version.
- [**Manifest**](docs/MANIFEST.md): Mapping of digests to physical files.
- [**Fixity**](docs/FIXITY.md): Optional additional checksums for integrity assurance.
- [**State**](docs/STATE.md): Logical view of the files in a version.
- [**Types**](docs/TYPES.md): Information about `User`, `VersionNumber`, and `InventorySpec`.

## Directory Structure

- `pkg/ocfl/inventory`: Contains core interfaces and type definitions.
- `pkg/ocfl/inventory/inventoryimpl`: Contains the default implementations of these interfaces.
- `pkg/ocfl/inventory/docs`: Detailed specification-oriented documentation of the data types.

## Implementation

The implementation in `inventoryimpl` is modular and uses base structures (`InventoryBase`, `VersionBase`) that implement the OCFL specification.
More details on the implementation can be found in [inventoryimpl/README.md](inventoryimpl/README.md).
