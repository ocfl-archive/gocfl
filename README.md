# GOCFL - Go OCFL Library

`gocfl` is a high-performance Go library for the [Oxford Common Filesystem Layout (OCFL)](https://ocfl.io/). It focuses on the creation, update, validation, and extraction of OCFL Storage Roots and Objects, with a strong emphasis on extensibility, I/O efficiency, and technical metadata indexing.

For additional extensions, please refer to the [gocfl-extensions](https://github.com/ocfl-archive/gocfl-extensions) repository.

> **Note**: This repository contains the `gocfl` library. For the command-line interface, please refer to the [gocfl-cli](https://github.com/ocfl-archive/gocfl-cli) repository.

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

The library provides convenient functions in the `initocfl` and `functions` packages for common tasks.

### Initializing and Loading a Storage Root

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

ctx := context.Background()
logger := ocfllogger.NewNopLogger()

// Initialize a new storage root (needs an appendfs.FS)
sr, err := initocfl.InitStorageRoot(ctx, fsys, version.Version1_1, logger)

// Load an existing storage root (needs an fs.FS)
sr, err := initocfl.LoadStorageRoot(ctx, fsys, logger)
```

### Working with Objects

```go
// Load an object from a filesystem
obj, err := initocfl.LoadObject(ctx, objFsys, logger)

// Add files to an object
writer, err := obj.StartUpdate("initial commit", "user", "user@example.com", false)
err = writer.AddFolder(sourceFS, true, "")
err = writer.Close()
```

### Validation and Extraction

```go
import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/functions"

// Validate an object
err := functions.CheckObject(ctx, objFsys, nil, logger)

// Extract an object
err := functions.Extract(ctx, objectFS, destFS, "object_path", nil, true, "", nil, logger)
```

## Supported Extensions

### Community Extensions
- [0001-digest-algorithms](https://ocfl.io/1.1/spec/extensions/0001-digest-algorithms.html)
- [0002-flat-direct-storage-layout](https://ocfl.io/1.1/spec/extensions/0002-flat-direct-storage-layout.html)
- [0003-hash-and-id-n-tuple-storage-layout](https://ocfl.io/1.1/spec/extensions/0003-hash-and-id-n-tuple-storage-layout.html)
- [0004-hashed-n-tuple-storage-layout](https://ocfl.io/1.1/spec/extensions/0004-hashed-n-tuple-storage-layout.html)
- [0006-flat-omit-prefix-storage-layout](https://ocfl.io/1.1/spec/extensions/0006-flat-omit-prefix-storage-layout.html)
- [0007-n-tuple-omit-prefix-storage-layout](https://ocfl.io/1.1/spec/extensions/0007-n-tuple-omit-prefix-storage-layout.html)

### Local & Special Extensions
- **NNNN-mets**: Generation of METS and PREMIS files.
- **NNNN-indexer**: Technical metadata extraction.
- **NNNN-migration**: Automatic file format migration on ingest.
- **NNNN-thumbnail**: Generation of thumbnails for video, images, and PDFs.
- **NNNN-filesystem**: Preserves filesystem metadata.
- **NNNN-pairtree-storage-layout**: Support for pairtree layouts.
- **NNNN-gocfl-extension-manager**: Internal manager for extension execution order.

## Documentation

- [Package Overview](./pkg/README.md)
- [OCFL Factory Documentation](./pkg/ocfl/factory/README.md)
- [Detailed Extension Docs](./docs/)

## License

GOCFL is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for details.
