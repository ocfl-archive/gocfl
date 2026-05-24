# Loader Module

The `Loader` module reads and parses an existing OCFL object from a filesystem. It discovers the object structure, version history, and associated extensions.

- **Interface**: `Loader` (`pkg/ocfl/object/object.go`)

## Key Methods

- `Load() error`: Reads the object structure, parses the [Inventory](../../inventory/README.md), and discovers extensions. See [Load Sequence](load_sequence.md) for details.
- `WithObject(o Object) Loader`: Associates the loader with an [Object](OBJECT.md) instance.
- `SetExtensionFactory(factory extension.Factory[ExtensionManager]) Loader`: Configures the [Extension Factory](../../extension/README.md) for instantiating extensions during load.
- `GetFS() fs.FS`: Returns the filesystem associated with the loader.

> [!NOTE]
> The target filesystem is associated with the `Object` via `WithReadFS(fsys)` before retrieving the loader.

## Usage Example

Typically, the loader is accessed via the [Object](OBJECT.md) interface:

```go
// Configure the object with a filesystem
obj.WithReadFS(sourceFS)

// Load the object
loader := obj.GetLoader()
if err := loader.Load(); err != nil {
    // handle error
}
```

The high-level function `ocfl.LoadObject` is the recommended way to load objects as it automates version detection and extension setup.

---
- [Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
