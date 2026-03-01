# Initializer Module

The `Initializer` module is responsible for creating a new OCFL storage root on a filesystem. It sets up the required directory structure, OCFL version markers, and initial extensions (like the storage layout).

- **Interface**: `Initializer` (`pkg/ocfl/storageroot/storageroot.go`)
- **Specification**: [OCFL 1.1 Storage Root Initialization](../../../../data/specs/ocfl_1.1.md#4-storage-root)

## Main Methods

The `Initializer` interface provides several key methods for configuring and executing the creation process:

- `Init() error`: Performs the actual initialization, creating the `0=ocfl_1.1` marker, the `extensions` directory, and the storage layout configuration.
- `WithStorageRoot(sr StorageRoot) Initializer`: Associates the initializer with a [StorageRoot](STORAGEROOT.md) instance.
 - `WithFS(objectFS appendfs.FS) Initializer`: Sets the destination filesystem where the storage root will be created.
- `Close() error`: Finalizes the initialization process.

## Usage Example

Typically, the initializer is accessed via the [StorageRoot](STORAGEROOT.md) interface:

```go
initializer := sr.GetInitializer().WithFS(appendFS)
if err := initializer.Init(); err != nil {
    // handle error
}
defer initializer.Close()
```

The [Factory](FACTORY.md) and high-level functions also provide convenient ways to instantiate initializers.

---
- [Back to Storage Root Overview](../README.md)
- [The StorageRoot Interface](STORAGEROOT.md)
