# The Object Interface

The `Object` interface (`pkg/ocfl/object/object.go`) is the central abstraction for OCFL objects. It provides access to metadata and serves as a factory for operational modules like loaders, initializers, and extractors.

- **Specification**: [OCFL 1.1 Object Specification](https://ocfl.io/1.1/spec/#object-structure)

## Interface Overview

### Operational Modules
- `GetInitializer() Initializer`: Returns an [Initializer](INITIALIZER.md).
- `GetLoader() Loader`: Returns a [Loader](LOADER.md).
- `GetExtractor() Extractor`: Returns an [Extractor](EXTRACTOR.md).
- `GetValidator() Validator`: Returns a [Validator](VALIDATOR.md).
- `StartUpdate(msg, name, address string, echo bool) (VersionWriter, error)`: Begins a new version update.

### Filesystem and State
- `WithReadFS(fsys fs.FS) Object`: Configures a read-only filesystem.
- `WithWriteFS(fsys appendfs.FS) Object`: Configures a writable filesystem (required for initialization and updates).
- `GetID() string`: Returns the object identifier.
- `GetInventory() inventory.Inventory`: Accesses the underlying [Inventory](../../inventory/README.md).
- `GetMetadata() (*inventory.Metadata, error)`: Returns standardized high-level metadata.
- `GetOCFLVersion() version.OCFLVersion`: Returns the OCFL version (e.g., `1.0`, `1.1`, or `2.0`).

## Implementations

Implementations are version-specific and managed via the [Factory](../../factory/README.md) pattern, allowing transparent support for different OCFL specification versions.

## Extensions

The `Object` manages its associated [Extension Manager](MODULES.md#extension-manager-integration) via:
- `WithExtensionManager(manager ExtensionManager) Object`
- `GetExtensionManager() ExtensionManager`

Extensions facilitate custom checksums, storage layouts, and metadata handling. See [Object Extension Hooks](HOOKS.md) for details on how extensions interact with object operations.

---
- [Object Documentation Overview](README.md)
- [Functional Modules](MODULES.md)
