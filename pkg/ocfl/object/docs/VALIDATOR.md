# Validator Module

The `Validator` module performs comprehensive validation and integrity checks on OCFL objects. It ensures compliance with the OCFL specification and verifies that internal structures and data remain consistent.

- **Interface**: `Validator` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1 Object Validation](https://ocfl.io/1.1/spec/#object-validation)

## Key Methods

- `Validate() error`: Executes the validation process. It verifies the presence of required files, validates the [Inventory](../../inventory/README.md) against its digests, and ensures the physical file layout matches the inventory state.
- `WithObject(obj Object) Validator`: Associates the validator with an [Object](OBJECT.md) instance.

> [!NOTE]
> The target filesystem is associated with the `Object` via `WithReadFS(fsys)` before retrieving the validator.

## Usage Example

The validator is typically accessed via the [Object](OBJECT.md) interface:

```go
// Configure the object with a filesystem
obj.WithReadFS(objectFS)

// Execute the validation
validator := obj.GetValidator()
if err := validator.Validate(); err != nil {
    // handle validation errors
}
```

The high-level function `ocfl.ValidateObject` is the recommended way to validate objects as it provides standardized setup and error reporting.

---
- [Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
