# Loader Module

The `Loader` module is responsible for reading and parsing an existing OCFL storage root from a filesystem. It handles the initial discovery of the storage root's structure, including identifying its layout and available extensions.

- **Interface**: `Loader` (`pkg/ocfl/storageroot/storageroot.go`)
- **Specification**: [OCFL 1.1 Storage Root Structure](../../../../data/specs/ocfl_1.1.md#4-storage-root)

## Main Methods

The `Loader` interface provides the following methods to configure and execute the loading process:

- `Load() error`: Performs the discovery and reading of the storage root structure from the filesystem.
- `WithStorageRoot(sr StorageRoot) Loader`: Associates the loader with a [StorageRoot](STORAGEROOT.md) instance.
- `WithFS(sourceFS fs.FS) Loader`: Sets the source filesystem where the storage root is located.
- `WithExtensionFactory(factory extension.Factory) Loader`: Configures the [Extension Factory](../../extension/docs/FACTORY.md) to use for instantiating extensions during the load process.
- `Close() error`: Finalizes the loading process and releases resources.

## Usage Example

Typically, the loader is accessed via the [StorageRoot](STORAGEROOT.md) interface:

```go
loader := sr.GetLoader(extensionFactory).WithFS(sourceFS)
if err := loader.Load(); err != nil {
    // handle error
}
defer loader.Close()
```

The [Factory](FACTORY.md) and high-level functions also provide convenient ways to instantiate loaders.

---
- [Back to Storage Root Overview](../README.md)
- [The StorageRoot Interface](STORAGEROOT.md)
