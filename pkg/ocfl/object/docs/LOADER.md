# Loader Module

The `Loader` module is responsible for reading and parsing an existing OCFL object from a filesystem. It handles the initial discovery of the object's structure and its version history.

- **Interface**: `Loader` (`pkg/ocfl/object/object.go`)

## Main Methods

The `Loader` interface includes several key methods for configuring and executing the loading process:

- `Load() error`: Performs the actual reading of the object structure from the filesystem, including parsing the [Inventory](../../inventory/README.md) and discovering available extensions.
- `SetObject(o Object) Loader`: Associates the loader with an [Object](OBJECT.md) instance.
- `SetFS(sourceFS fs.FS) Loader`: Sets the source filesystem where the OCFL object is located.
- `SetExtensionFactory(factory extension.Factory) Loader`: Configures the [Extension Factory](../../extension/README.md) to use for instantiating extensions during the load process.
- `GetFS() fs.FS`: Returns the filesystem associated with the loader.

## Usage Example

Typically, the loader is accessed via the [Object](OBJECT.md) interface:

```go
loader := obj.GetLoader(sourceFS, extensionFactory)
if err := loader.Load(); err != nil {
    // handle error
}
```

Higher-level orchestration functions in [pkg/ocfl/functions](../../functions/README.md) like `LoadObject` wrap this logic for common use cases.

---
- [Back to Object Overview](../README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
