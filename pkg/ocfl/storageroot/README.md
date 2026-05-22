# OCFL Storage Root Package

The `pkg/ocfl/storageroot` package provides the core interfaces and abstractions for managing OCFL (Oxford Common File Layout) Storage Roots. While the [Object](../object/README.md) package handles individual OCFL objects, the `storageroot` package manages the higher-level structure that contains these objects, including storage layouts and root-level extensions.

## Key Features

- **Abstraction**: Decouples storage root operations from specific OCFL versions and storage implementations.
- **Modularity**: Operations are split into specialized functional modules:
    - [Loader](docs/LOADER.md): For reading and parsing an existing storage root.
    - [Initializer](docs/INITIALIZER.md): For creating a new storage root.
- **Storage Layouts**: Supports OCFL storage layouts through the [Extension Manager](docs/STORAGEROOT.md#extension-manager) and `ExtensionStorageRootPath` interface.
- **Factory Pattern**: Integrates with [OCFL Factory](../factory/README.md) to instantiate the correct implementation based on the OCFL version.

## Documentation

For a comprehensive overview of the storage root components and their technical details, see the [Storage Root Documentation Overview](docs/README.md).

## Related Components

- [OCFL Object](../object/README.md): Manages individual OCFL objects within the storage root.
- [OCFL Functions](../ocflactions/README.md): High-level orchestration functions that often involve storage roots.
- [OCFL Factory](../factory/README.md): Responsible for creating `StorageRoot` instances.
- [OCFL Extension](../extension/README.md): Provides the base for storage root extensions, particularly layouts.
- [OCFL 1.1 Specification](../../../data/specs/ocfl_1.1.md): The underlying specification this package implements, specifically the [Storage Root section](../../../data/specs/ocfl_1.1.md#4-storage-root).

## Usage Overview

Common high-level operations like loading a storage root are typically performed via the `initocfl` package.

```go
import (
    "github.com/ocfl-archive/gocfl/v3/pkg/initocfl"
)

// Example: Loading a storage root (recommended way)
sr, srCloser, err := initocfl.LoadStorageRoot(ctx, sourceFS, nil, logger)
if err != nil {
    // handle error
}
defer srCloser.Close()

// Check if an object exists by ID
exists, err := sr.ObjectExists("some-object-id")
```

For more complex scenarios, you can interact with the specialized functional modules directly. See [The StorageRoot Interface](docs/STORAGEROOT.md) for details.
