# Extension Factory

The `Factory` (`factory.go`) is the creation engine for extensions and managers. It provides a registry-based approach to instantiate extensions by name and load them from filesystem data.

A default implementation is provided in the `extensionimpl` subpackage (`pkg/ocfl/extension/extensionimpl/factory.go`).

## Factory Interface

Key methods of the `Factory` interface:
- `AddCreator(name string, creator CreatorFunc, documentation *string)`: Registers a new extension creator.
- `RegisterExtension(name string, builder BuilderFunc, documentation *string)`: Registers a new extension builder.
- `LoadExtensionFile(fsys fs.FS) (Extension, error)`: Loads an extension configuration from a `config.json` file in the provided filesystem.
- `LoadExtensionData(data json.RawMessage, extFS fs.FS) (Extension, error)`: Loads an extension from raw JSON data.
- `LoadExtensionManager(fsys fs.FS) (T, error)`: Instantiates a manager and populates it with extensions found in the filesystem.
- `GetExtensionDocs() map[string]*string`: Returns a map of registered extension names to their documentation.
- **Defaults**: Allows adding default extensions for both storage roots and objects via `AddStorageRootDefaultExtension(ext Extension)` and `AddObjectDefaultExtension(ext Extension)`.

## Comparison: Factory vs. Manager

For a high-level comparison between the Factory and the Manager, see the [Extension Documentation README](README.md#core-concepts-manager-vs-factory).

The **Factory** is the "Knowledge Base" and "Manufacturer" of the system:

- **Knowledge**: It knows which extension names (e.g. `0002-flat-direct-storage-layout`) belong to which Go implementations.
- **Creation**: It can turn a JSON configuration or a directory on disk into a live `Extension` object.
- **Bootstrapping**: It is used at the very beginning to either create a new `Manager` or to restore a `Manager` from an existing OCFL object/root.

In contrast, the **Manager** is the "Orchestrator" at runtime. Once the Factory has created the extensions, the Manager holds them and calls them whenever the OCFL Object or Storage Root needs to perform an action (like calculating a path or writing a file).

See [Extension Manager](MANAGER.md) for more details on the runtime behavior.

The `Factory` is used by the [Object Loader](../../object/docs/LOADER.md) and the [Storage Root Loader](../../storageroot/docs/LOADER.md) to reconstruct extension state from their respective `extensions/` directories.

## Implementation Details: `extensionimpl.Factory`

The default implementation in `extensionimpl` (see [`pkg/ocfl/extension/extensionimpl`](../extensionimpl)) includes several advanced features:
- Uses a map of `CreatorFunc` to instantiate specific extensions by their `extensionName`.
- Supports the `initial` extension specifically when loading from a filesystem.
- Validates that extension names match their folder names (as per OCFL recommendations).
- Integrates with the `ocfllogger` to report validation errors or warnings during the loading process.
- Handles both object-level and storage-root-level extensions.

---
- [Back to Extension Overview](README.md)
- [Extension Interface](EXTENSION.md)
- [Manager Interface](MANAGER.md)
