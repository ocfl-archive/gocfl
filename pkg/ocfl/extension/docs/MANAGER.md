# Extension Manager

The `ManagerCore` (`manager.go`) is responsible for coordinating multiple extensions within an OCFL object or storage root. It manages the order of execution, dependencies, and persistence of extension configurations.

- **Specification**: [OCFL 1.1 Object Extensions](../../../../data/specs/ocfl_1.1.md#311-extensions-directory)
- **Specification**: [OCFL 1.1 Storage Root Extensions](../../../../data/specs/ocfl_1.1.md#43-extensions-directory)

## ManagerCore Interface

Key methods of the `ManagerCore` interface:
- **Aggregation**: Maintains a list of active extensions via `GetExtensions()` and allows adding new ones via `Add(ext Extension)`.
- **Configuration Management**: Handles the storage and retrieval of configuration files for all managed extensions via `GetConfigName()`.
- **Finalization**: `Finalize()` method for completing manager setup.
- **Root Layout Support**: Specific methods for handling storage root layouts via `StoreRootLayout(fsys appendfs.FS)`.
- **Initialization**: Support for an `Initial` extension via `SetInitial(initial Initial)`.

## Difference between Manager and Factory

For a high-level comparison, see the [Extension Documentation README](README.md#core-concepts-manager-vs-factory).

While both components deal with extensions, they have distinct responsibilities:

| Feature | Extension Manager (`ManagerCore`) | Extension Factory (`Factory`) |
| :--- | :--- | :--- |
| **Role** | **Runtime Coordinator** | **Registry & Creator** |
| **Lifecycle** | Manages extensions *during* their use in an object/root. | Handles the *instantiation* of extensions. |
| **State** | Holds instances of active extensions. | Holds a registry of available extension types. |
| **Interaction** | Decides which extension to call for a specific task. | Knows how to build an extension from JSON/Filesystem. |
| **Persistence** | Responsible for writing configs to the `extensions/` folder. | Responsible for reading from the `extensions/` folder. |

In short: The **Factory** *creates* the extensions (and the manager itself), and the **Manager** *uses* them to perform OCFL operations.

## Initial Extension

The `Initial` interface is used for the "initial" extension as defined in the [OCFL Community Extension `initial`](../../../../docs/initial.md). Its purpose is to indicate that a specific extension takes precedence and must be applied before all other extensions.

Key features of the `Initial` extension in this library:
- **Precedence**: It points to the primary extension that coordinates others (typically the extension manager).
- **Configuration**: The target extension is specified via the `extension` parameter in the `initial` extension's configuration.
- **Factory Integration**: The [Factory](FACTORY.md) uses the `initial` extension to determine which `ManagerCore` implementation to instantiate.

For more details, see the [**Initial Extension Specification**](../../../../docs/initial.md). In the context of this library, the extension referenced by `initial` typically implements the `ManagerCore` interface (or more specifically `ExtensionManager` in the `object` package) to handle the coordination of all other extensions.

## Integration

The Manager is typically integrated into the [Object](../../object/docs/OBJECT.md) to handle object-level extensions and into the [Storage Root](../../storageroot/docs/STORAGEROOT.md) for root-level layouts and configurations.

### Object-Specific Manager

In the `object` package, a more specialized interface `ExtensionManager` is defined in [`pkg/ocfl/object/extensionManager.go`](../../object/extensionManager.go). It embeds `ManagerCore` and adds multiple object-specific extension interfaces:

- **Path Management**: `ExtensionObjectContentPath`, `ExtensionObjectStatePath`, `ExtensionObjectExtractPath`.
- **Change Hooks**: `ExtensionContentChange`, `ExtensionObjectChange`.
- **Fixity & Metadata**: `ExtensionFixityDigest`, `ExtensionMetadata`.
- **System Hooks**: `ExtensionArea`, `ExtensionStream`, `ExtensionNewVersion`.

This allows the [Object](../../object/docs/OBJECT.md) to interact with extensions at various stages of their lifecycle (e.g., during file addition, version creation, or metadata extraction).

Die primäre Implementierung dieses erweiterten Interface ist [`GOCFLExtensionManager`](../../../../../gocfl-extensions/extension/NNNN-gocfl-extension-manager.go) (identifiziert als `NNNN-gocfl-extension-manager`, siehe [Doc](../../../../docs/NNNN-gocfl-extension-manager.md)).

### Storage Root-Specific Manager

The `storageroot` package also defines a specialized `ExtensionManager` that coordinates [Storage Layouts](../../storageroot/docs/STORAGEROOT.md#extensionstoragerootpath) alongside standard extensions. For more details, see the [Storage Root Extension Manager](../../storageroot/docs/STORAGEROOT.md#extension-manager).

---
- [Back to Extension Overview](README.md)
- [Extension Interface](EXTENSION.md)
- [Initial Specification](../../../../docs/initial.md)
- [Factory Interface](FACTORY.md)
- [Storage Root Documentation](../../storageroot/README.md)
