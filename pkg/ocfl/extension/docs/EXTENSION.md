# Extension Interface

An `Extension` (`extension.go`) is a single functional unit that implements a specific OCFL extension (e.g., storage layout, fixity, or metadata).

- **Specification**: [OCFL 1.1: 5. Extensions](../../version/ocfl_spec_1.1.md#5-extensions)
- **External Docs**: [OCFL Extensions](https://ocfl.io/extensions/)

## Interface Definition

Key methods of the `Extension` interface:
- `WithLogger(logger ocfllogger.OCFLLogger) Extension`: Sets the logger for the extension.
- `GetName() string`: Returns the unique name of the extension (e.g., `0002-flat-direct-storage-layout`).
- `Load(data json.RawMessage, extFS fs.FS) error`: Loads the extension's configuration from JSON data and an optional filesystem.
- `SetParams(params map[string]string) error`: Configures the extension via key-value parameters.
- `WriteConfig(fsys appendfs.FS) error`: Persists the extension's configuration to the OCFL object or storage root.
- `GetConfig() any`: Returns the extension's configuration object.
- `IsRegistered() bool`: Returns whether the extension is registered in the factory.
- `Terminate() error`: Performs cleanup when the extension is no longer needed.

## Creating a Custom Extension

To create your own extension, follow these steps:

### 1. Implementing the Interface
Create a struct that implements the `extension.Extension` interface. Optionally, you can embed `extension.ExtensionConfig` to manage the extension's name by default.

```go
type MyExtension struct {
    *extension.ExtensionConfig
    // Custom configuration fields
}

func (m *MyExtension) GetName() string { return "my-custom-extension" }
// ... implementation of other methods ...
```

### 2. Registration
Extensions must be registered when the application starts. This is typically done in an `init()` function within the extension's package. There are two types of registration:

- `extension.RegisterExtensionStorageRoot`: For extensions active in an OCFL Storage Root.
- `extension.RegisterExtensionObject`: For extensions active within an OCFL Object.

```go
func init() {
    extension.RegisterExtensionStorageRoot("my-custom-extension", NewMyExtension, GetMyExtensionParams, &MyExtensionDoc)
    extension.RegisterExtensionObject("my-custom-extension", NewMyExtension, GetMyExtensionParams, &MyExtensionDoc)
}
```

### 3. Extension Hooks (Interaction)
The actual functionality of an extension is realized via **Hooks**. The `ExtensionManager` uses type assertion to check if an extension implements additional interfaces.

#### Important Hooks for Storage Roots:
- `storageroot.ExtensionStorageRootPath`: Influences how object IDs are mapped to filesystem paths (`BuildStorageRootPath`).

#### Important Hooks for Objects:
Detailed descriptions of these hooks can be found under [Object Extension Hooks](../../object/docs/HOOKS.md).

- `object.ExtensionObjectContentPath`: Transforms paths in the manifest (`BuildObjectManifestPath`).
- `object.ExtensionContentChange`: Allows actions before/after adding, updating, or deleting files (`AddFileBefore`, `AddFileAfter`, etc.).
- `object.ExtensionMetadata`: Adds its own information to the object metadata output (`GetMetadata`).
- `object.ExtensionFixityDigest`: Registers additional checksum algorithms.

To use a hook, simply implement the corresponding interface in your extension struct. A comprehensive guide with code examples can be found at [docs/custom_extensions.md](../../../docs/custom_extensions.md).

## Implementations & Examples

Each extension should implement this interface. Common implementations in this library (located in [`pkg/extensions/`](../../extensions/README.md)) include:
- **Storage layout extensions**: (e.g., [`0002-flat-direct-storage-layout`](https://ocfl.io/extensions/0002-flat-direct-storage-layout.html)). These are used within [Storage Roots](../../storageroot/README.md) to determine object placement.
- **Fixity extensions**: For additional checksums (e.g., [`0001-digest-algorithms`](https://ocfl.io/extensions/0001-digest-algorithms.html)).
- **Metadata extensions**: For custom descriptive or technical metadata.
- **Extension Manager**: The [`GOCFLExtensionManager`](../../extensions/ext_NNNN_gocfl_extension_manager/NNNN-gocfl-extension-manager.md) is itself an extension that coordinates others.

## Usage

In most cases, you don't instantiate extensions directly. Instead, you use a [Factory](FACTORY.md) to load them from a filesystem or JSON data:

```go
// 1. Create a factory (typically done by the Object/StorageRoot loader)
// T is the manager type, e.g. extension.ManagerCore[extension.Extension]
factory, _ := extensionimpl.NewFactory[extension.ManagerCore[extension.Extension]](params, logger)

// 2. Load an extension from a directory (containing config.json)
ext, err := factory.LoadExtensionFile(extensionFS)
if err != nil {
    log.Fatal(err)
}

// 3. Use the extension
name := ext.GetName()
config := ext.GetConfig()
```

Extensions are typically managed by a [Manager](MANAGER.md) (either for [Objects](../../object/docs/OBJECT.md) or [Storage Roots](../../storageroot/docs/STORAGEROOT.md)).

---
- [Back to Extension Overview](README.md)
- [OCFL Object Structure](../../object/docs/OBJECT.md)
- [OCFL Storage Root](../../storageroot/docs/STORAGEROOT.md)
