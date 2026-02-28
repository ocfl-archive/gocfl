# OCFL Object Package

The `pkg/ocfl/object` package provides the core interfaces and abstractions for managing OCFL (Oxford Common File Layout) objects. Unlike the `inventory` package, which focuses on data structures and JSON representation, the `object` package handles the high-level orchestration of object operations like initialization, loading, updating, extraction, and validation.

## Key Features

- **Abstraction**: Decouples object operations from specific OCFL versions and storage implementations.
- **Modularity**: Operations are split into specialized functional modules: [Loader](docs/LOADER.md), [Initializer](docs/INITIALIZER.md), [VersionWriter](docs/VERSION_WRITER.md), [Checker](docs/CHECKER.md), and [Extractor](docs/EXTRACTOR.md).
- **Function Modules**: Uses [Functional Modules](docs/MODULES.md) to provide a modular architecture for object lifecycle management.
- **Factory Pattern**: Integrates with [Factory](../factory/README.md) to instantiate the correct implementation based on the OCFL version.

## Documentation

- [The Object Interface](docs/OBJECT.md): The central interface representing an OCFL object.
- [Functional Modules](docs/MODULES.md): Overview of the specialized modules for object operations.
  - [Loader](docs/LOADER.md): Reading and parsing objects.
  - [Initializer](docs/INITIALIZER.md): Creating new objects.
  - [VersionWriter](docs/VERSION_WRITER.md): Adding new versions.
  - [Checker](docs/CHECKER.md): Validating object integrity.
  - [Extractor](docs/EXTRACTOR.md): Retrieving object content.

## Related Components

- [OCFL Functions](../functions/README.md): High-level, project-wide orchestration functions for common object tasks.
- [OCFL Factory](../factory/README.md): Responsible for creating `Object` instances.
- [OCFL Inventory](../inventory/README.md): Handles the `inventory.json` structure and logic.
- [OCFL 1.1 Specification](../../../data/specs/ocfl_1.1.md): The underlying specification this package implements (also supports OCFL 1.0 and 2.0).

## Usage Overview

Common high-level operations like loading and checking objects are often performed via `pkg/ocfl/functions`.

```go
import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/functions"

// Loading an object
obj, err := functions.LoadObject(ctx, sourceFS, extensionFactory, logger)
if err != nil {
    // handle error
}

// Accessing metadata
metadata, err := obj.GetMetadata()
```

For more complex scenarios, you can interact with the specialized functional modules directly. See [Function Modules](docs/MODULES.md) for details.
