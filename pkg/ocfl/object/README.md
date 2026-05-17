# OCFL Object Package

The `pkg/ocfl/object` package provides the core interfaces and abstractions for managing OCFL (Oxford Common File Layout) objects. Unlike the `inventory` package, which focuses on data structures and JSON representation, the `object` package handles the high-level orchestration of object operations like initialization, loading, updating, extraction, and validation.

## Key Features

- **Abstraction**: Decouples object operations from specific OCFL versions and storage implementations.
- **Modularity**: Operations are split into specialized functional modules:
    - [Loader](docs/LOADER.md): For loading existing objects.
    - [Initializer](docs/INITIALIZER.md): For creating new objects.
    - [VersionWriter](docs/VERSION_WRITER.md): For adding or modifying files in a version.
    - [Checker](docs/CHECKER.md): For validating object integrity and compliance.
    - [Extractor](docs/EXTRACTOR.md): For extracting content from the object.
- **Extension Support**: Comprehensive [Extension System](docs/HOOKS.md) for customizing object behavior.
- **Factory Pattern**: Integrates with [OCFL Factory](../factory/README.md) to instantiate the correct implementation based on the OCFL version.

## Documentation

For a comprehensive overview of the object components and their technical details, see the [Object Documentation Overview](docs/README.md).

## Related Components

- [OCFL Functions](../ocflactions/README.md): High-level, project-wide orchestration functions for common object tasks.
- [OCFL Factory](../factory/README.md): Responsible for creating `Object` instances.
- [OCFL Inventory](../inventory/README.md): Handles the `inventory.json` structure and logic.
- [OCFL 1.1 Specification](../../../data/specs/ocfl_1.1.md): The underlying specification this package implements (also supports OCFL 1.0 and 2.0).

## Usage Overview

Common high-level operations like loading and checking objects are often performed via the `initocfl` and `ocflactions` packages.

```go
import (
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
)

// Loading an object (recommended way)
obj, objCloser, err := initocfl.LoadObject(ctx, sourceFS, nil, logger)
if err != nil {
    // handle error
}
defer objCloser.Close()

// Accessing metadata
metadata, err := obj.GetExtractor().GetMetadata()
```

For more complex scenarios, you can interact with the specialized functional modules directly. See [Function Modules](docs/MODULES.md) for details.
