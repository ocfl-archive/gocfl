# Checker Module

The `Checker` module is responsible for performing comprehensive validation and integrity checks on an existing OCFL object. It ensures that the object complies with the OCFL specification and that its internal structure and data are consistent.

- **Interface**: `Checker` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1: 3.5.3.3 Object Validation](../../../../data/specs/ocfl_1.1.md#3533-object-validation)

## Main Methods

The `Checker` interface includes several key methods for configuring and executing validation:

- `Check() error`: Executes the validation process. It checks the presence of required files, validates the [Inventory](../../inventory/README.md) against its digests, and ensures that the physical file layout matches the inventory.
- `SetObject(obj Object) Checker`: Associates the checker with an [Object](OBJECT.md) instance.
- `SetFS(objectFS fs.FS) Checker`: Sets the filesystem where the OCFL object to be checked is located.

## Usage Example

The checker is typically accessed via the [Object](OBJECT.md) interface:

```go
checker := obj.GetChecker(sourceFS)
if err := checker.Check(); err != nil {
    // handle validation errors
}
```

Higher-level orchestration functions in [pkg/ocfl/functions](../../functions/README.md) like `CheckObject` provide a convenient wrapper for validating objects.

---
- [Back to Object Overview](../README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
