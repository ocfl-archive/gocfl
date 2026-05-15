# Checker Module

The `Checker` module is responsible for performing comprehensive validation and integrity checks on an existing OCFL object. It ensures that the object complies with the OCFL specification and that its internal structure and data are consistent.

- **Interface**: `Checker` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1: 3.5.3.3 Object Validation](../../version/ocfl_spec_1.1.md#3533-object-validation)

## Main Methods

The `Checker` interface includes several key methods for configuring and executing validation:

- `Check() error`: Executes the validation process. It checks the presence of required files, validates the [Inventory](../../inventory/README.md) against its digests, and ensures that the physical file layout matches the inventory.
- `WithObject(obj Object) Checker`: Associates the checker with an [Object](OBJECT.md) instance.

> [!NOTE]
> The target filesystem is now passed to the `Object` via `WithReadFS(fsys)` before calling `GetChecker()`.

## Usage Example

The checker is typically accessed via the [Object](OBJECT.md) interface:

```go
// Set the filesystem on the object
obj.WithReadFS(objectFS)

// Run the check
checker := obj.GetChecker()
if err := checker.Check(); err != nil {
    // handle validation errors
}
```

The high-level function `ocflactions.CheckObject` in [pkg/ocfl/ocflactions](../actions.go) is the recommended way to validate objects as it handles common setup and error reporting.

---
- [Back to Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
