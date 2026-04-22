# OCFL Extension Package

The `extension` package provides the core interfaces and management logic for OCFL extensions. Extensions are the official way to add functionality to OCFL objects and storage roots, as defined in the [OCFL Specification](https://ocfl.io/extensions/).

This package is divided into four primary components:

1. [**Extension Interface**](docs/EXTENSION.md): Functional units for specific OCFL extensions.
2. [**Manager Interface**](docs/MANAGER.md): Coordination of multiple extensions.
3. [**Initial Extension Spec**](../../../../docs/initial.md): Identification of the primary manager.
4. [**Factory Interface**](docs/FACTORY.md): Registry-based creation of extensions and managers.

- **Specification**: [OCFL 1.1: 5. Extensions](../../../data/specs/ocfl_1.1.md#5-extensions)
- **External Docs**: [OCFL Extensions](https://ocfl.io/extensions/)

## Directory Structure

- `pkg/ocfl/extension`: Core interfaces (`extension.go`, `manager.go`, `factory.go`).
- `pkg/ocfl/extension/extensionimpl`: Default implementation of the extension factory and related utilities.
- `pkg/ocfl/extension/docs`: Detailed documentation for each component.

## Integration

The extension system interacts closely with other core components:

- **Objects**: The [Object](../object/docs/OBJECT.md) interface uses a Manager to apply extensions that affect object structure or metadata.
- **Inventories**: Extensions can influence how [Inventories](../inventory/README.md) are handled, particularly regarding fixity or custom metadata fields.
- **Storage Roots**: Extensions like `0002-flat-direct-storage-layout` or `0003-hash-and-id-direct-storage-layout` define how objects are organized within a [Storage Root](../storageroot/README.md).
- **Implementations**: Concrete extension implementations are located in [`pkg/extension`](../../../../gocfl-extensions/extension). See the [Extensions Directory](../../../docs/README.md#extensions) (or individual files in [`/docs`](../../../docs)) for their documentation.

---
- [Back to Object Documentation](../object/docs/OBJECT.md)
- [Back to Storage Root Documentation](../storageroot/README.md)
- [Back to Inventory Documentation](../inventory/README.md)
