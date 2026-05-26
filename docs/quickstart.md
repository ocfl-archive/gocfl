# Quickstart Guide

This guide provides simple code examples for common OCFL operations using the `gocfl` library. You can find fully runnable versions of these examples in the [examples](../examples/) directory.

## Virtual Filesystem (vfsrw)

OCFL operations in `gocfl` are performed via a filesystem abstraction layer called `vfsrw`. This allows the library to work with various storage backends (local, S3, memory, etc.) using a unified interface.

```go
import (
    "github.com/je4/filesystem/v4/pkg/vfsrw"
    "github.com/je4/filesystem/v4/pkg/appendfs"
    "github.com/je4/utils/v2/pkg/zLogger"
    "github.com/rs/zerolog"
    "os"
)

// Setup a logger (required by vfsrw)
out := zerolog.ConsoleWriter{Out: os.Stderr}
zlogger := zerolog.New(out)
var _zlogger zLogger.ZLogger = &zlogger

// Create a new VFS instance
cfg := vfsrw.Config{}
vfs, err := vfsrw.NewFS(cfg, _zlogger)
if err != nil {
    // handle error
}
defer vfs.Close()

// Register the local filesystem
if err := vfsrw.AddLocal(vfs, nil); err != nil {
    // handle error
}

// Create a sub-filesystem for your OCFL storage root
// appendfs is required for write operations
fsys, err := appendfs.Sub(vfs, "C:/path/to/ocfl/root")
if err != nil {
    // handle error
}
```

## Storage Root Initialization

Initialize a new OCFL Storage Root in a given filesystem.

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
    "github.com/je4/utils/v2/pkg/checksum"
)

// ... setup logger and filesystem (fsys) using vfsrw ...

ctx := context.Background()
ocflVer := version.Version1_1

// Create factories
extFactory, _ := extensionimpl.NewFactory[storageroot.ExtensionManager](nil, logger)
srFactory := ocfl.NewFactoryStorageRoot(ocflVer, extFactory, logger)

// Initialize Storage Root
sr := srFactory.NewStorageRoot(ctx).
	WithWriteFS(fsys).
	WithDigestAlgorithm(checksum.DigestSHA512)

initializer := sr.GetInitializer()
defer initializer.Close()
err := initializer.Init()
if err != nil {
    // handle error
}
```

## Creating an Object

Create and initialize a new OCFL Object within the storage root.

```go
// objFS is an appendfs.FS pointing to the object's folder
obj := objectFactory.NewObject(ctx)
obj.WithWriteFS(objFS)
initializer := obj.GetInitializer()
defer initializer.Close()

err := initializer.Init("my-object-id", checksum.DigestSHA512, nil)
if err != nil {
    // handle error
}
```

## Adding and Updating Content

Add files to an object or update existing ones by creating a new OCFL version.

```go
// Start an update to create a new version
vw, err := obj.StartUpdate("Adding initial files", "User Name", "user@example.com", false)
if err != nil {
    // handle error
}
defer vw.Close()

// Add data as a file
err = vw.AddData([]byte("Hello OCFL"), "folder/hello.txt", false, "", false, false)
if err != nil {
    // handle error
}

// Finalize the version and write the inventory
err = vw.Close()
if err != nil {
    // handle error
}
```

## Extracting Objects

Extract the content of a specific version (or the head) to a destination filesystem.

```go
// destFS is where the files will be extracted
extractor := obj.GetExtractor().WithDestFS(destFS)
defer extractor.Close()

// Extract the head version
err := extractor.Extract(nil, false, "")
if err != nil {
    // handle error
}
```
