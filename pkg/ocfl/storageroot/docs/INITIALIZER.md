# Initializer Module

The `Initializer` module is responsible for creating a new OCFL storage root on a filesystem. It sets up the required directory structure, OCFL version markers, and initial extensions (like the storage layout).

- **Interface**: `Initializer` (`pkg/ocfl/storageroot/storageroot.go`)
- **Specification**: [OCFL 1.1 Storage Root Initialization](../../version/ocfl_spec_1.1.md#4-storage-root)

## Main Methods

The `Initializer` interface provides several key methods for configuring and executing the creation process:

- `Init() error`: Performs the actual initialization, creating the `0=ocfl_1.1` marker, the `extensions` directory, and the storage layout configuration.
- `SetStorageRoot(sr StorageRoot) Initializer`: Associates the initializer with a [StorageRoot](STORAGEROOT.md) instance.
- `Close() error`: Finalizes the initialization process.

## Usage Example

Typically, the initializer is accessed via the [StorageRoot](STORAGEROOT.md) interface:

```go
initializer := sr.WithWriteFS(appendFS).GetInitializer()
if err := initializer.Init(); err != nil {
    // handle error
}
defer initializer.Close()
```

The high-level function `initocfl.InitStorageRoot` in [pkg/ocfl/initocfl](../../initocfl/README.md) is the recommended way to initialize storage roots.

---
- [Back to Storage Root Overview](../README.md)
- [The StorageRoot Interface](STORAGEROOT.md)
