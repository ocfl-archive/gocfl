# Inventory Architecture

This documentation describes the architecture of the interfaces in the `pkg/ocfl/inventory` package. The implementation follows the hierarchical structure of the OCFL specification and uses Go interfaces for a clear separation of concerns.

## Class Diagram

The following diagram shows the relationships between the central interfaces and how they are connected through composition and embedding.

```mermaid
classDiagram
    class Inventory {
        <<interface>>
        +GetVersions() Versions
        +GetManifest() Manifest
        +GetFixity() Fixity
        +AddFile()
        +Finalize()
    }
    class Versions {
        <<interface>>
        +GetVersion() Version
        +LatestVersionNumber()
        +NewVersion()
        +AddFile()
        +DeleteFile()
        +RenameFile()
        +CopyFile()
        +EchoDelete()
    }
    class Version {
        <<interface>>
        +GetState() State
        +GetUser() User
        +GetCreated()
        +AddFile()
        +DeleteFile()
        +RenameFile()
        +CopyFile()
        +EchoDelete()
    }
    class Manifest {
        <<interface>>
        +AddFile()
        +GetFiles()
    }
    class State {
        <<interface>>
        +GetFiles()
        +FileChecksum()
        +AddFile()
        +DeleteFile()
        +RenameFile()
        +CopyFile()
        +EchoDelete()
    }
    class Fixity {
        <<interface>>
        +AddFile()
        +GetFiles()
    }
    class User {
        <<interface>>
        +GetName()
        +GetAddress()
    }
    Inventory *-- Versions
    Inventory *-- Manifest
    Inventory *-- Fixity
    Versions "1" *-- "n" Version
    Version *-- State
    Version *-- User
```

## Explanation of Components

- **Inventory**: The main interface representing the entire OCFL inventory content.
- **Versions**: Manages the history of all versions of the object and provides file operations.
- **Version**: Represents a specific version with metadata (User, Message, Created), a state (State), and supports file operations.
- **Manifest**: Maps digests to physical paths (content of the object).
- **State**: Maps logical paths of a version to digests and supports file operations.
- **Fixity**: Contains optional additional checksums for physical files.

A detailed diagram of the high-level object structure and the operational interfaces can be found in the [**Object Architecture Documentation**](../../object/docs/architecture.md). Repository-wide management is explained in the [**Storage Root Architecture Documentation**](../../storageroot/docs/architecture.md).
