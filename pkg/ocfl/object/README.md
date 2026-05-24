# OCFL Object Package

The `pkg/ocfl/object` package provides the core interfaces for managing OCFL (Oxford Common File Layout) objects. It handles high-level orchestration for operations such as initialization, loading, updating, extraction, and validation.

## Key Features

- **Abstraction**: Decouples object operations from specific OCFL versions and storage implementations.
- **Modularity**: Operations are divided into specialized modules:
    - [Loader](docs/LOADER.md): Loads existing objects.
    - [Initializer](docs/INITIALIZER.md): Creates new objects.
    - [VersionWriter](docs/VERSION_WRITER.md): Adds or modifies files in a version.
    - [Validator](docs/VALIDATOR.md): Validates object integrity and OCFL compliance.
    - [Extractor](docs/EXTRACTOR.md): Retrieves content and metadata.
- **Extension Support**: A robust [Extension System](docs/HOOKS.md) for customizing behavior.
- **Factory Integration**: Works with the [OCFL Factory](../factory/README.md) to instantiate the appropriate version-specific implementations.

## Documentation

See the [Object Documentation Overview](docs/README.md) for technical details and architecture diagrams.

## Related Components

- [OCFL Core](./README.md): High-level orchestration for common tasks.
- [OCFL Factory](../factory/README.md): Responsible for creating `Object` instances.
- [OCFL Inventory](../inventory/README.md): Manages the `inventory.json` structure.
- [OCFL Specification](https://ocfl.io/): The underlying specification (supports 1.0, 1.1, and 2.0).

## Usage Overview

High-level operations are typically performed via the `ocfl` package.

```go
import (
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
)

// Load an object (recommended)
obj, objCloser, err := ocfl.LoadObject(ctx, sourceFS, nil, logger)
if err != nil {
    // handle error
}
defer objCloser.Close()

// Access metadata via the Extractor
metadata, err := obj.GetExtractor().GetMetadata()
```

For advanced scenarios, interact directly with specialized modules. See [Function Modules](docs/MODULES.md) for more information.
