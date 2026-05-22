# OCFL Inventory Data Types Documentation

This directory contains detailed documentation for the various data types and modules that make up an OCFL Inventory.

## Documentation Files

- [**Architecture**](architecture.md): Visual representation of the interface hierarchy and their relationships.
- [**Object Architecture**](../../object/docs/architecture.md): How the inventory integrates into the Object structure.
- [**Inventory**](INVENTORY.md): Documentation of the central inventory structure, which serves as the main entry point for an OCFL object's metadata.
- [**Version**](VERSION.md): Details about individual version metadata, including the state of the object at that version.
- [**Manifest**](MANIFEST.md): Documentation of the manifest structure, which maps content digests to physical file paths.
- [**Fixity**](FIXITY.md): Information about optional fixity blocks used to provide additional checksums for integrity verification.
- [**State**](STATE.md): Describes the logical state of an object version, mapping logical paths to content digests.
- [**Types**](TYPES.md): Documentation of common types used throughout the inventory, such as `User`, `VersionNumber`, and `InventorySpec`.

## Structure

The inventory implementation in this package is designed to be modular, following the logical separation defined in the OCFL specification. Each of these documents describes the interfaces and data structures used to implement these components.
