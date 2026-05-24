# Initializer Module

The `Initializer` module sets up a new OCFL object on a target filesystem. It handles the creation of the initial folder structure and the base `inventory.json`.

- **Interface**: `Initializer` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1 Object Root](https://ocfl.io/1.1/spec/#object-root)

## Key Methods

- `Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error`: Initializes the object, setting the identifier, primary digest algorithm (e.g., `sha512`), and any additional [Fixity](../../inventory/docs/FIXITY.md) algorithms. See [Creation Sequence](create_sequence.md) for details.
- `WithObject(o Object) Initializer`: Associates the initializer with an [Object](OBJECT.md) instance.

> [!NOTE]
> The target filesystem is associated with the `Object` via `WithWriteFS(fsys)` before retrieving the initializer.

## Usage Example

The initializer is accessed via the [Object](OBJECT.md) interface:

```go
// Configure the object with a writable filesystem
obj.WithWriteFS(fsys)

// Initialize the object
initializer := obj.GetInitializer()
if err := initializer.Init(id, digest, fixity); err != nil {
    // handle error
}
```

The high-level function `ocfl.InitObject` is the recommended way to initialize objects as it automates factory setup and filesystem configuration.

---
- [Object Overview](../README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
