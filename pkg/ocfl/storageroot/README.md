# OCFL Storage Root Package

The `pkg/ocfl/storageroot` package provides the core interfaces and abstractions for managing OCFL (Oxford Common File Layout) Storage Roots. While the [Object](../object/README.md) package handles individual OCFL objects, the `storageroot` package manages the higher-level structure that contains these objects, including storage layouts and root-level extensions.

## Key Features

- **Abstraction**: Decouples storage root operations from specific OCFL versions and storage implementations.
- **Modularity**: Operations are split into specialized functional modules: [Loader](docs/LOADER.md) and [Initializer](docs/INITIALIZER.md).
- **Storage Layouts**: Supports OCFL storage layouts through the [Extension Manager](docs/STORAGEROOT.md#extension-manager) and `ExtensionStorageRootPath` interface.
- **Factory Pattern**: Integrates with [Factory](docs/FACTORY.md) to instantiate the correct implementation based on the OCFL version.

## Documentation

- [The StorageRoot Interface](docs/STORAGEROOT.md): The central interface representing an OCFL storage root and its Extension Manager.
- [Functional Modules](docs/STORAGEROOT.md#functional-modules): Overview of the specialized modules for storage root operations.
  - [Loader](docs/LOADER.md): Reading and parsing an existing storage root.
  - [Initializer](docs/INITIALIZER.md): Creating a new storage root.
- [Factory Interface](docs/FACTORY.md): Details on the factory used to create `StorageRoot` instances.

## Related Components

- [OCFL Object](../object/README.md): Manages individual OCFL objects within the storage root.
- [OCFL Inventory](../inventory/README.md): While primarily for objects, storage roots also interact with inventory concepts for object discovery.
- [OCFL Extension](../extension/README.md): Provides the base for storage root extensions, particularly layouts.
- [OCFL 1.1 Specification](../../../data/specs/ocfl_1.1.md): The underlying specification this package implements, specifically the [Storage Root section](../../../data/specs/ocfl_1.1.md#4-storage-root).

## Usage Overview

Common high-level operations like loading a storage root are typically performed via the [Factory](docs/FACTORY.md).

```go
// Example: Loading a storage root
sr := factory.NewStorageRoot(ctx)
loader := sr.GetLoader(extensionFactory).WithReadFS(sourceFS)
if err := loader.Load(); err != nil {
    // handle error
}

// Check if an object exists by ID
exists, err := sr.ObjectExists("some-object-id")
```

---
- [Back to Project Root](../../../README.md)
