# OCFL Core Packages

This directory contains the core packages of the `gocfl` project, which implement the [OCFL (Oxford Common File Layout) specification](https://ocfl.io/).

## Package Overview

The functionality is divided into several specialized packages covering different aspects of the OCFL specification:

### [Extension](./extension/README.md)
The `extension` package provides core interfaces and management logic for OCFL extensions. Extensions are the official way to add additional functionality to OCFL objects and storage roots.
- [Extension Interface](./extension/docs/EXTENSION.md)
- [Manager Interface](./extension/docs/MANAGER.md)
- [Factory Interface](./extension/docs/FACTORY.md)

### [Factory](./factory/README.md)
The `factory` module provides a unified factory mechanism to correctly instantiate components according to the various OCFL specification versions (e.g., 1.0, 1.1, 2.0).

### [OCFL Actions](./ocflactions/README.md)
Contains high-level orchestration functions for common tasks such as loading, creating, checking, or extracting OCFL objects. It serves as the primary API for many use cases.

### [Inventory](./inventory/README.md)
This package defines the structures and interfaces for the OCFL Inventory (`inventory.json`). It is the heart of an OCFL object and contains all metadata about versions, files, and fixity information.
- [**Architecture Diagram**](./inventory/docs/architecture.md)
- [Inventory](./inventory/docs/INVENTORY.md)
- [Version](./inventory/docs/VERSION.md)
- [Manifest](./inventory/docs/MANIFEST.md)
- [Fixity](./inventory/docs/FIXITY.md)

### [Object](./object/README.md)
Manages the high-level operations of OCFL objects. While the `inventory` package focuses on data structures, `object` handles the loading, initializing, updating, and validating of objects.
- [Object Interface](./object/docs/OBJECT.md)
- [Functional Modules](./object/docs/MODULES.md) (Loader, Initializer, etc.)

### [StorageRoot](./storageroot/README.md)
Manages the OCFL Storage Root, the top-level structure containing OCFL objects. It handles the Storage Layout and root-level extensions.
- [StorageRoot Interface](./storageroot/docs/STORAGEROOT.md)

### [Validation](./validation/README.md)
Provides structures and functions for handling OCFL validation errors and warnings according to the specification.

---

## Utility Packages

- **[ocflerrors](./ocflerrors/errors.go)**: Central definition of OCFL-specific error types.
- **[util](./util/helper.go)**: Helper functions for OCFL operations (e.g., version detection in the file system).
- **[version](./version/README.md)**: Definition of supported OCFL versions.

---
- [Back to parent package documentation](../README.md)
- [Back to project main page](../../README.md)
