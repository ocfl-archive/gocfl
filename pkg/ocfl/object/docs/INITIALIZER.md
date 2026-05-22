# Initializer Module

The `Initializer` module is used to set up a brand-new OCFL object on a target filesystem. It handles the initial folder structure creation and the generation of the base `inventory.json`.

- **Interface**: `Initializer` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1: 3.1 Object Root](../../version/ocfl_spec_1.1.md#31-object-root)

## Main Methods

The `Initializer` interface includes methods for configuring and performing the initialization:

- `Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error`: Performs the initialization, including setting the object identifier, primary digest algorithm (e.g., `sha512`), and any additional [Fixity](../../inventory/docs/FIXITY.md) algorithms. See the [Creation Sequence Diagram](create_sequence.md) for details.
- `WithObject(o Object) Initializer`: Associates the initializer with an [Object](OBJECT.md) instance.

> [!NOTE]
> The target filesystem is no longer set directly on the `Initializer`. Instead, use `obj.WithWriteFS(fsys)` on the `Object` before calling `GetInitializer()`.

## Usage Example

The initializer is accessed via the [Object](OBJECT.md) interface:

```go
// Set the writable filesystem on the object
obj.WithWriteFS(fsys)

// Get the initializer and run it
initializer := obj.GetInitializer()
if err := initializer.Init(id, digest, fixity); err != nil {
    // handle error
}
```

The high-level function `initocfl.InitObject` in [pkg/ocfl/initocfl](../../initocfl/README.md) is the recommended way to initialize objects as it handles factory setup and filesystem configuration automatically.

---
- [Back to Object Overview](../README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
