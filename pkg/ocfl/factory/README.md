# OCFL Unified Factory Module (`pkg/ocfl/factory`)

The `pkg/ocfl/factory` module provides a **Unified Factory** mechanism to instantiate components correctly according to different OCFL (Oxford Common File Layout) specification versions (e.g., 1.0, 1.1, 2.0).

## Goal
The primary objective of this module is to ensure that all OCFL-related objects—such as Inventories, Storage Roots, Objects, and Versions—are created with the correct implementation and configuration matching the desired OCFL version through a single, consistent interface.

## Documentation
Detailed architectural information and sequence diagrams can be found in the [docs](./docs) directory:
- [**Architecture Overview**](./docs/architecture.md)
- [**Object Factory Initialization Sequence**](./docs/initfactoryobject.md)
- [**Storage Root Factory Initialization Sequence**](./docs/initfactorystorageroot.md)

## Core Components

### `FactoryObject` and `FactoryStorageRoot` Interfaces
Defined in `pkg/ocfl/factory/factoryInterface.go`, these are the central interfaces that aggregate several sub-factories into a **Unified Factory**:
- `FactoryObject`:
    - `object.Factory`: For creating object loaders, initializers, checkers, extractors, and object instances.
    - `inventory.Factory`: For creating inventories, fixity information, users, manifests, versions, and states.
- `FactoryStorageRoot`:
    - `storageroot.Factory`: For creating storage roots, storage root loaders, and initializers.

They also provide methods for version management (`GetVersion`, `WithNewVersion`), configuration (`WithConfig`), and cloning (`Copy`).

### `FactoryBaseObject` and `FactoryBaseStorageRoot`
These classes (in `pkg/ocfl/factory/factoryimpl/`) provide the standard implementation of the Unified Factory interfaces. They hold the common dependencies like `extensionFactory` and `logger` and handle the delegation to component-specific implementations based on the OCFL version.

### Version-Specific Factories
The module provides specialized factory constructors for each supported OCFL version via the `initocfl` package:
- `NewFactoryObject(version, ...)`
- `NewFactoryStorageRoot(version, ...)`

These functions dispatch to version-specific wrappers in `factoryimpl` (e.g., `factoryObject11`, `factoryStorageRoot11`).

## Usage

### Creating a Fixed-Version Unified Factory
Use the `initocfl` package to create a factory for a specific OCFL version:

```go
import (
    "github.com/ocfl-archive/gocfl/v3/pkg/initocfl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// Create a unified factory for OCFL 1.1 objects
f := initocfl.NewFactoryObject(version.Version1_1, extFactory, logger)

// Use the factory to create a new object
obj := f.NewObject(ctx)
```

### `WithConfig`
Both `FactoryObject` and `FactoryStorageRoot` support a `WithConfig` method. This allows passing a configuration map to the factory, which will be used when creating components (like loaders, initializers, etc.).

```go
import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"

conf := map[object.ConfigName]any{
    object.LoaderName: &objectimpl.LoaderConfig{},
}
f := initocfl.NewFactoryObject(version.Version1_1, extFactory, logger).WithConfig(conf)
```

## Directory Structure

- `pkg/ocfl/factory/`: Contains the public interfaces (`factoryInterface.go`).
- `pkg/ocfl/factory/docs/`: Detaillierte Dokumentation und Diagramme.
- `pkg/ocfl/factory/factoryimpl/`: Contains concrete implementations:
    - `factorybaseobject.go` / `factorybasestorageroot.go`: Common implementation logic.
    - `factoryObject10.go`, `factoryObject11.go`, etc.: Version-specific wrappers.
- `pkg/ocfl/initocfl/`: Entry point for creating factory instances (`factory.go`).
