# The Object Interface

The `Object` interface (`pkg/ocfl/object/object.go`) is the central abstraction for an OCFL object in the `gocfl` library. It provides access to object-level metadata and serves as a factory for operational interfaces like loaders, initializers, and extractors.

- **Specification**: [OCFL 1.1 Object Specification](../../version/ocfl_spec_1.1.md#3-object-structure)

## Main Interface

The `Object` interface includes the following key methods:

### Operational Accessors
- `GetInitializer() Initializer`: Returns an [Initializer](INITIALIZER.md) to create a new object.
- `GetLoader() Loader`: Returns a [Loader](LOADER.md) to load an existing object.
- `GetExtractor() Extractor`: Returns an [Extractor](EXTRACTOR.md) to retrieve files and versions from the object.
- `GetChecker() Checker`: Returns a [Checker](CHECKER.md) to validate the object's integrity and OCFL compliance.
- `StartUpdate(msg string, name string, address string, echo bool) (VersionWriter, error)`: Begins a new version update.

### Filesystem and State
- `WithReadFS(fsys fs.FS) Object`: Configures the object with a read-only filesystem.
- `WithWriteFS(fsys appendfs.FS) Object`: Configures the object with a writable filesystem (required for `Initializer` and `VersionWriter`).
- `GetReadFS() fs.FS`: Returns the current read filesystem.
- `GetWriteFS() appendfs.FS`: Returns the current write filesystem.
- `GetID() string`: Returns the object identifier.
- `GetInventory() inventory.Inventory`: Provides access to the underlying [Inventory](../../inventory/README.md) object.
- `GetMetadata() (*inventory.Metadata, error)`: Returns high-level object metadata in a standardized format.
- `GetOCFLVersion() version.OCFLVersion`: Returns the OCFL specification version the object adheres to (e.g., `1.0`, `1.1`, or `2.0`).
- `Stat(w io.Writer, statInfo []StatInfo) error`: Writes statistical information about the object to the provided writer.

## Implementations

Actual implementations of this interface (like `ObjectImpl`) are typically version-specific and are managed via the [Factory](../../factory/README.md) pattern. This allows the library to support [OCFL 1.0](https://ocfl.io/1.0/spec/), [OCFL 1.1](https://ocfl.io/1.1/spec/), and [OCFL 2.0](https://ocfl.io/draft/spec/) transparently.

## Interaction with Extensions

The `Object` interface provides methods to manage and retrieve the [Extension Manager](MODULES.md#integration-with-extension-manager):
- `WithExtensionManager(manager ExtensionManager) Object`
- `GetExtensionManager() ExtensionManager`

Extensions allow the library to support additional functionality like custom checksums, storage layouts, or metadata handling, as defined in the [OCFL Extensions Specification](https://github.com/OCFL/extensions/). 

A detailed description of the available hooks for objects can be found under [Object Extension Hooks](HOOKS.md).

The default implementation of the `ExtensionManager` is the [`GOCFLExtensionManager`](../../../../../gocfl-extensions/extension/NNNN-gocfl-extension-manager.go) (identified as `NNNN-gocfl-extension-manager`).

---
- [Back to Object Documentation Overview](README.md)
- [Function Modules](MODULES.md)
