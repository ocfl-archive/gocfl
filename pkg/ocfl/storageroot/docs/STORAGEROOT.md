# The StorageRoot Interface

The `StorageRoot` interface (`pkg/ocfl/storageroot/storageroot.go`) is the central abstraction for an OCFL Storage Root. It manages the filesystem, [Extension Manager](#extension-manager), and provides factory methods for its functional modules.

- **Specification**: [OCFL 1.1 Storage Root](../../../../data/specs/ocfl_1.1.md#4-storage-root)

## Main Interface

The `StorageRoot` interface includes the following key areas:

### Functional Modules
- `GetLoader(extensionFactory extension.Factory) Loader`: Returns a [Loader](LOADER.md) to load an existing storage root.
- `GetInitializer() Initializer`: Returns an [Initializer](INITIALIZER.md) to create a new storage root.

### Filesystem and State
- `WithReadFS(sourceFS fs.FS) StorageRoot` / `GetReadFS() fs.FS`: Handles the read-only filesystem where the storage root is located.
 - `WithWriteFS(appendFS appendfs.FS) StorageRoot` / `GetWriteFS() appendfs.FS`: Handles the writable filesystem for updates and creations.
- `IsModified() bool` / `SetModified()`: Tracks whether the storage root state has changed.

### Object Management
- `GetObjectFolders() ([]string, error)`: Lists the folders that contain OCFL objects within the storage root.
- `ObjectExists(id string) (bool, error)`: Checks if an object with the given ID exists in the storage root.
- `IdToFolder(id string) (folder string, err error)`: Translates an object identifier to its storage folder using the active layout extension.
- `Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error`: Provides statistical information about the storage root or objects within it.

### Metadata
- `GetOCFLVersion() version.OCFLVersion`: Returns the OCFL specification version the storage root adheres to.
- `GetDigest() checksum.DigestAlgorithm` / `SetDigest(digest checksum.DigestAlgorithm)`: Manages the default digest algorithm for the storage root.

---

## Extension Manager

The storage root uses a specialized `ExtensionManager` (`pkg/ocfl/storageroot/storagerootExtensionManager.go`) to coordinate extensions. Die Standardimplementierung ist der [`GOCFLExtensionManager`](../../../../../gocfl-extensions/extension/NNNN-gocfl-extension-manager.go) (identifiziert als `NNNN-gocfl-extension-manager`).

### ExtensionManager Interface

The `ExtensionManager` for storage roots extends the base [ManagerCore](../../extension/docs/MANAGER.md) and adds support for storage layouts:

- **Inheritance**: Embeds `extension.ManagerCore`.
- **Layout Support**: Embeds `ExtensionStorageRootPath`.

### ExtensionStorageRootPath

This interface (`pkg/ocfl/storageroot/storagerootExtension.go`) is critical for OCFL storage roots as it defines how object identifiers are mapped to storage paths.

- **Methods**:
  - `WriteLayout(fsys appendfs.FS) error`: Persists the layout configuration.
  - `BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)`: Implements the logic to map an ID to a path.
- **Specification**: [OCFL 1.1 Storage Layouts](../../../../data/specs/ocfl_1.1.md#42-storage-layout)

---
- [Back to Storage Root Overview](../README.md)
- [Loader Module](LOADER.md)
- [Initializer Module](INITIALIZER.md)
