# Initializer Module

The `Initializer` module is used to set up a brand-new OCFL object on a target filesystem. It handles the initial folder structure creation and the generation of the base `inventory.json`.

- **Interface**: `Initializer` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1: 3.1 Object Root](../../../../data/specs/ocfl_1.1.md#31-object-root)

## Main Methods

The `Initializer` interface includes methods for configuring and performing the initialization:

- `Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error`: Performs the initialization, including setting the object identifier, primary digest algorithm (e.g., `sha512`), and any additional [Fixity](../../inventory/docs/FIXITY.md) algorithms.
- `SetObject(o Object) Initializer`: Associates the initializer with an [Object](OBJECT.md) instance.
- `SetFS(objectFS appendfs.FS) Initializer`: Sets the target [append-capable filesystem](../../appendfs/README.md) for initialization.

## Usage Example

The initializer is accessed via the [Object](OBJECT.md) interface:

```go
initializer := obj.GetInitializer(fsys)
if err := initializer.Init(id, digest, fixity); err != nil {
    // handle error
}
```

Higher-level orchestration functions in [pkg/ocfl/functions](../../functions/README.md) like `CreateObject` simplify this process.

---
- [Back to Object Overview](../README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
