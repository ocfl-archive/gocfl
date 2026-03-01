# The Object Interface

The `Object` interface (`pkg/ocfl/object/object.go`) is the central abstraction for an OCFL object in the `gocfl` library. It provides access to object-level metadata and serves as a factory for operational interfaces like loaders, initializers, and extractors.

- **Specification**: [OCFL 1.1 Object Specification](../../../../data/specs/ocfl_1.1.md#3-object-structure)

## Main Interface

The `Object` interface includes the following key methods:

### Operational Accessors
- `GetInitializer(objectFS appendfs.FS) Initializer`: Returns an [Initializer](INITIALIZER.md) to create a new object.
- `GetLoader(sourceFS fs.FS, extensionFactory extension.Factory) Loader`: Returns a [Loader](LOADER.md) to load an existing object.
- `GetExtractor(objectFS fs.FS) Extractor`: Returns an [Extractor](EXTRACTOR.md) to retrieve files and versions from the object.
- `GetChecker(sourceFS fs.FS) Checker`: Returns a [Checker](CHECKER.md) to validate the object's integrity and OCFL compliance.
- `StartUpdate(...) (VersionWriter, error)`: Begins a new version update using a [VersionWriter](VERSION_WRITER.md).

### Metadata and Info
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

Extensions allow the library to support additional functionality like custom checksums, storage layouts, or metadata handling, as defined in the [OCFL Extensions Specification](https://ocfl.io/extensions/). Die Standardimplementierung des `ExtensionManager` ist der [`GOCFLExtensionManager`](../../../extension/NNNN-gocfl-extension-manager.go) (identifiziert als `NNNN-gocfl-extension-manager`).

---
- [Back to Object Overview](../README.md)
- [Function Modules](MODULES.md)
