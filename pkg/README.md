# Package Documentation

This directory contains the core packages of the `gocfl` project. Each sub-package provides specific functionality for handling OCFL (Oxford Common Filesystem Layout) storage roots, objects, and related metadata.

## Overview of Packages

### [ocfl](./ocfl/README.md)
The main package for OCFL operations. It includes several sub-packages for different aspects of the OCFL specification.
- [extension](./ocfl/extension/README.md): OCFL extension management.
- [factory](./ocfl/factory/README.md): Factory for creating OCFL objects and storage roots.
- [initocfl](./ocfl/initocfl/README.md): Initializer for OCFL storage roots and objects.
- [inventory](./ocfl/inventory/README.md): Handling of OCFL inventory files.
- [object](./ocfl/object/README.md): OCFL object management.
- [ocflactions](./ocfl/ocflactions/README.md): Common actions for OCFL objects.
- [storageroot](./ocfl/storageroot/README.md): OCFL storage root management.
- [validation](./ocfl/validation/README.md): OCFL validation logic.
- [version](./ocfl/version/version.go): OCFL version definitions and specifications ([v1.0](./ocfl/version/ocfl_spec_1.0.md), [v1.1](./ocfl/version/ocfl_spec_1.1.md)).

### [extensions](./extensions/README.md)
Contains various extensions for the `gocfl` tool.

### [ocfllogger](./ocfllogger)
A specialized logger package used throughout the `gocfl` project to ensure consistent logging and error reporting.

### [util](./util)
General utility functions used across the project, including file system helpers and program detection.

---

For more general information about the project, please refer to the main [README.md](../README.md) in the root directory.
