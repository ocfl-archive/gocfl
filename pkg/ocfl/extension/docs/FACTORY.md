# Extension Factory

The `Factory` (`factory.go`) is the creation engine for extensions and managers. It provides a registry-based approach to instantiate extensions by name and load them from filesystem data.

A default implementation is provided in the `extensionimpl` subpackage (`pkg/ocfl/extension/extensionimpl/factory.go`).

## Factory Interface

Key methods of the `Factory` interface:
- `AddCreator(name string, creator CreatorFunc)`: Registers a new extension implementation.
- `LoadExtensionFile(fsys fs.FS)`: Loads an extension configuration from a `config.json` file.
- `LoadExtensionManager(fsys fs.FS, ver OCFLVersion)`: Instantiates a manager and populates it with extensions found in the filesystem.
- **Defaults**: Allows adding default extensions for both storage roots and objects via `AddStorageRootDefaultExtension` and `AddObjectDefaultExtension`.

The `Factory` is used by the [Object Loader](../../object/docs/LOADER.md) and the [Storage Root Loader](../../storageroot/docs/LOADER.md) to reconstruct extension state from their respective `extensions/` directories.

## Implementation Details: `extensionimpl.Factory`

The default implementation in `extensionimpl` (see [`pkg/ocfl/extension/extensionimpl`](../extensionimpl)) includes several advanced features:
- Uses a map of `CreatorFunc` to instantiate specific extensions by their `extensionName`.
- Supports the `initial` extension specifically when loading from a filesystem.
- Validates that extension names match their folder names (as per OCFL recommendations).
- Integrates with the `ocfllogger` to report validation errors or warnings during the loading process.
- Handles both object-level and storage-root-level extensions.

---
- [Back to Extension Overview](../README.md)
- [Extension Interface](EXTENSION.md)
- [Manager Interface](MANAGER.md)
