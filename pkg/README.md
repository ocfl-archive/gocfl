# Package Documentation

This directory contains the core packages of the `gocfl` project. Each sub-package provides specific functionality for handling OCFL (Oxford Common Filesystem Layout) storage roots, objects, and related metadata.

## Overview of Packages

### [dilcis](subsystem/dilcis/README.md)
Contains implementations for DILCIS standards, including EAD3, METS, and PREMIS, used for metadata handling.
- [EAD3](subsystem/dilcis/ead3)
- [METS](subsystem/dilcis/mets)
- [PREMIS](subsystem/dilcis/premis)

### [ocfl](./ocfl)
The main package for OCFL operations. It includes several sub-packages for different aspects of the OCFL specification.
- [extension](./ocfl/extension/README.md): OCFL extension management.
- [factory](./ocfl/factory/README.md): Factory for creating OCFL objects and storage roots.
- [inventory](./ocfl/inventory/README.md): Handling of OCFL inventory files.
- [object](./ocfl/object/README.md): OCFL object management.
- [storageroot](./ocfl/storageroot/README.md): OCFL storage root management.
- [validation](./ocfl/validation/README.md): OCFL validation logic.

### [ocfllogger](./ocfllogger)
A specialized logger package used throughout the `gocfl` project to ensure consistent logging and error reporting.

### [appendfs](./appendfs/README.md)
Provides an extended file system interface (`FS`) that supports both read and write operations, abstracting the underlying storage (e.g., local FS, S3).

### [subsystem](./subsystem/README.md)
Contains various subsystems used by `gocfl` for processing content.
- [migration](./subsystem/migration/README.md): Subsystem for file format migrations.
- [thumbnail](./subsystem/thumbnail/README.md): Subsystem for generating thumbnails.

### [extension](./extension)
General extension mechanisms used in the project.

---

For more general information about the project, please refer to the main [README.md](../README.md) in the root directory.
