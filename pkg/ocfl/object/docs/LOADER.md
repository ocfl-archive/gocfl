# Loader Module

The `Loader` module is responsible for reading and parsing an existing OCFL object from a filesystem. It handles the initial discovery of the object's structure and its version history.

- **Interface**: `Loader` (`pkg/ocfl/object/object.go`)

## Main Methods

The `Loader` interface includes several key methods for configuring and executing the loading process:

- `Load() error`: Performs the actual reading of the object structure from the filesystem, including parsing the [Inventory](../../inventory/README.md) and discovering available extensions.
- `WithObject(o Object) Loader`: Associates the loader with an [Object](OBJECT.md) instance.
- `SetExtensionFactory(factory extension.Factory[ExtensionManager]) Loader`: Configures the [Extension Factory](../../extension/README.md) to use for instantiating extensions during the load process.
- `GetFS() fs.FS`: Returns the filesystem associated with the loader.

> [!NOTE]
> Parameters like the filesystem are now passed to the `Object` via `WithReadFS(fsys)` before calling `GetLoader()`.

## Usage Example

Typically, the loader is accessed via the [Object](OBJECT.md) interface:

```go
// Configure the object with a filesystem
obj.WithReadFS(sourceFS)

// Get the loader and load the object
loader := obj.GetLoader()
if err := loader.Load(); err != nil {
    // handle error
}
```

The high-level function `initocfl.LoadObject` in [pkg/ocfl/initocfl](../../initocfl/README.md) is the recommended way to load existing objects as it handles version detection and extension setup automatically.

---
- [Back to Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
