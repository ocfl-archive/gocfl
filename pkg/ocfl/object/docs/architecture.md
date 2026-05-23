# Object Architecture

This document describes the architecture of the `pkg/ocfl/object` package. While the `inventory` package defines OCFL data structures, the `object` package implements the logic and orchestration of object operations.

## Class Diagram

The following diagram illustrates the `Object` interface, its specialized modules, and its relationship to the inventory.

```mermaid
classDiagram
    class Object {
        <<interface>>
        +GetLoader() Loader
        +GetInitializer() Initializer
        +GetExtractor() Extractor
        +GetValidator() Validator
        +GetInventory() Inventory
        +StartUpdate(msg, name, address, echo) VersionWriter
        +GetExtensionManager() ExtensionManager
    }

    class ExtensionManager {
        <<interface>>
    }

    note for ExtensionManager "See Extension Docs"

    class Loader {
        <<interface>>
        +Load()
        +GetFS() FS
    }

    class Initializer {
        <<interface>>
        +Init(id, digest, fixity)
    }

    class Extractor {
        <<interface>>
        +Extract(version, withManifest, area)
        +GetMetadata() Metadata
        +GetFileReader(name) ReadCloser
    }

    class VersionWriter {
        <<interface>>
        +AddFile(sourceFS, path, checkDuplicate, area, noExt, isDir)
        +AddFolder(sourceFS, checkDuplicate, area)
        +AddData(data, path, checkDuplicate, area, noExt, isDir)
        +AddReader(r, files, area, noExt, isDir)
        +DeleteFile(name, digest)
        +RenameFile(src, dest, digest)
        +Close()
    }

    class Validator {
        <<interface>>
        +Validate()
    }

    class Inventory {
        <<interface>>
    }

    note for Inventory "See Inventory Docs"

    Object ..> Loader : provides
    Object ..> Initializer : provides
    Object ..> Extractor : provides
    Object ..> Validator : provides
    Object ..> VersionWriter : creates
    Object o-- Inventory : manages
    Object o-- ExtensionManager : manages
    Loader --|> Object : embeds
```

## Component Overview

- **Object**: Central orchestrator and entry point for all operations.
- **Loader**: Loads existing objects. See [Load Sequence](load_sequence.md).
- **Initializer**: Creates new objects. See [Creation Sequence](create_sequence.md).
- **Extractor**: Accesses content and metadata.
- **VersionWriter**: Handles write operations and versioning. See [Update Sequence](update_sequence.md).
- **Validator**: Validates compliance with the OCFL specification.
- **ExtensionManager**: Manages object extensions. See [Extension Architecture](../../extension/docs/architecture.md).
- **Inventory**: Represents the OCFL state and metadata. See [Inventory Architecture](../../inventory/docs/architecture.md).

## Implementation Details

The `object` package uses the `inventory` package for state persistence. High-level structure is governed by the [Storage Root](../../storageroot/docs/architecture.md).
