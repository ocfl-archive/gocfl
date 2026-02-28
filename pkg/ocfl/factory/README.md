# OCFL Unified Factory Module (`pkg/factory`)

The `pkg/factory` module provides a **Unified Factory** mechanism to instantiate components correctly according to the different OCFL (Oxford Common File Layout) specification versions (e.g., 1.0, 1.1, 2.0).

## Goal
The primary objective of this module is to ensure that all OCFL-related objects—such as Inventories, Storage Roots, Objects, and Versions—are created with the correct implementation and configuration matching the desired OCFL version through a single, consistent interface.

## Core Components

### `Factory` Interface (Unified Factory)
Defined in `pkg/ocfl/factory/factory.go`, this is the central interface that aggregates several sub-factories into a single **Unified Factory**:
- `inventory.Factory`: For creating inventories, fixity, users, manifests, versions, and states.
- `object.Factory`: For creating object loaders, initializers, checkers, extractors, and object instances.
- `storageroot.Factory`: For creating storage roots, storage root loaders, and initializers.

It also provides methods to get/set the OCFL version and to create a copy of the factory.

### `FactoryBase`
The `FactoryBase` (in `pkg/ocfl/factory/factoryimpl/factorybase.go`) provides the standard implementation for the **Unified Factory** interface. It holds the `OCFLVersion`, the `InventorySpec`, an `extensionFactory`, and a `logger`. It is responsible for calling the appropriate constructors in `inventoryimpl`, `objectimpl`, and `storagerootimpl` with the correct version parameters.

### Version-Specific Factories
The module provides specialized factory constructors for each supported OCFL version, each returning an instance of the **Unified Factory**:
- `NewFactory10`: Configured for OCFL v1.0.
- `NewFactory11`: Configured for OCFL v1.1.
- `NewFactory20`: Configured for OCFL v2.0.

### `DynamicFactory`
The `DynamicFactory` is a variant of the **Unified Factory** that allows for switching the OCFL version at runtime using the `SetVersion` method. When the version is changed, it re-instantiates the underlying factory while preserving common dependencies like the extension factory and logger.

## Usage

### Creating a Fixed-Version Unified Factory
If you know the OCFL version beforehand, use the dispatcher `NewFactory`:

```go
import (
    "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
    "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

// Create a unified factory for OCFL 1.1
f := factoryimpl.NewFactory(version.Version1_1, extFactory, logger)

// Use the factory to create a new inventory
inv := f.NewInventory(ctx)
```

### Creating a Dynamic Unified Factory
If you need to change the version later (e.g., after detecting it from a storage root or object):

```go
f := factoryimpl.NewDynamicFactory(version.Version1_1, extFactory, logger)

// ... detect version from filesystem ...
err := f.SetVersion(version.Version1_0)
if err != nil {
    // handle error
}

// Now the unified factory will create 1.0 components
obj := f.NewObject(ctx)
```

## Directory Structure

- `pkg/ocfl/factory/`: Contains the public `Factory` interface.
- `pkg/ocfl/factory/factoryimpl/`: Contains concrete implementations:
    - `factorybase.go`: Common implementation logic.
    - `factory.go`: Dispatcher for fixed-version factories.
    - `dynamicFactory.go`: Runtime-adjustable factory.
    - `factory10.go`, `factory11.go`, `factory20.go`: Version-specific wrappers.
