# GOCFL - Go OCFL Library

`gocfl` is a high-performance Go library for the [Oxford Common Filesystem Layout (OCFL)](https://ocfl.io/). It focuses on the creation, update, validation, and extraction of OCFL Storage Roots and Objects, with a strong emphasis on extensibility, I/O efficiency, and technical metadata indexing.

> **Note**: This repository contains the `gocfl` library. For the command-line interface, please refer to the [gocfl-cli](https://github.com/ocfl-archive/gocfl-cli) repository.
> 
> For additional extensions, please refer to the [gocfl-extensions](https://github.com/ocfl-archive/gocfl-extensions) repository.

## Features

- **OCFL Support**: Full support for OCFL v1.0 and v1.1, with experimental support for v2.0.
- **Storage Backends**:
  - Local Filesystem
  - S3 Cloud Storage (via MinIO Client SDK)
  - Serialization into ZIP Containers (with optional AES Encryption)
- **High Performance**: 
  - Optimized I/O: Files are read and written as few times as possible.
  - Concurrent checksum generation and processing.
- **Extensibility**: Implements a flexible extension hook system (7 different hooks) for both Storage Root and Object extensions.
- **Advanced Capabilities**:
  - Technical metadata extraction and indexing.
  - File format migration.
  - Thumbnail generation.
  - METS/PREMIS generation.

## Installation

```bash
go get github.com/ocfl-archive/gocfl/v3
```

## Basic Usage

The library provides convenient functions in the `initocfl` and `ocflactions` packages for common tasks.

### Initializing and Loading a Storage Root

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/initocfl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
    "github.com/je4/utils/v2/pkg/checksum"
)

ctx := context.Background()
logger := ocfllogger.NewNopLogger()

// Initialize a new storage root (needs an appendfs.FS)
sr, err := initocfl.InitStorageRoot(ctx, fsys, nil, version.Version1_1, checksum.DigestSHA512, nil, logger)

// Load an existing storage root (needs an fs.FS)
sr, srCloser, err := initocfl.LoadStorageRoot(ctx, fsys, nil, logger)
defer srCloser.Close()
```

### Working with Objects

```go
// Load an object from a filesystem
obj, objCloser, err := initocfl.LoadObject(ctx, objFsys, nil, logger)
defer objCloser.Close()

// Add files to an object
writer, err := obj.StartUpdate("initial commit", "user", "user@example.com", false)
err = writer.AddFolder(sourceFS, true, "")
err = writer.Close()
```

### Validation and Extraction

```go
import (
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflactions"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
)

// Validate an object
err := ocflactions.CheckObject(ctx, objFsys, logger)

// Extract an object
err := ocflactions.Extract(ctx, objectFS, destFS, "object_path", inventory.NewVersionNumber(), true, "", logger)
```

## Examples

The [examples](./examples/README.md) directory contains runnable Go programs for common tasks:

- **[Storage Root Initialization](./examples/storageroot_init/main.go)**: How to initialize a new OCFL Storage Root.
- **[Object Creation](./examples/object_add/main.go)**: Creating and initializing a new OCFL Object.
- **[Adding and Updating Content](./examples/object_upd/main.go)**: Adding files and creating new versions.
- **[Object Validation](./examples/object_validate/main.go)**: Validating an existing OCFL Object.
- **[Extracting Objects](./examples/object_extract/main.go)**: Extracting object content to a filesystem.

## Documentation

- [Quickstart Guide](./docs/quickstart.md) - Start here for a quick overview and code examples.
- [Documentation Overview](./docs/README.md) - Links to detailed component documentation.
- [Package Overview](./pkg/README.md)
- [OCFL Factory Documentation](./pkg/ocfl/factory/README.md)

## License

GOCFL is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for details.
