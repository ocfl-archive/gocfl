# Object Load Sequence

This document describes the sequence of operations performed during the loading of an existing OCFL Object, starting from the high-level entry point.

## Sequence Diagram

The following diagram illustrates the steps taken by `initocfl.LoadObject()` and the internal `Loader.Load()` method.

```mermaid
sequenceDiagram
    participant App as Application
    participant Init as initocfl
    participant Util as util (Version Discovery)
    participant FS as ReadFS (Filesystem)
    participant Obj as Object
    participant Factory as FactoryObject
    participant Loader as Loader
    participant EM as ExtensionManager
    participant Inv as Inventory

    App->>Init: LoadObject(ctx, FS, ...)
    
    rect rgb(220, 240, 220)
    Note over Init, Util: Version Detection
    Init->>Util: GetObjectVersion(FS)
    Util->>FS: Read Namaste file (0=ocfl_...)
    Util-->>Init: OCFL Version
    Init->>Init: logger.WithVersion(Version)
    end

    rect rgb(240, 240, 240)
    Note over Init, Factory: Setup & Instantiation
    Init->>Init: NewExtensionFactory()
    Init->>Init: NewFactoryObject(Version, ExtensionFactory)
    Init->>Factory: NewObject(ctx)
    Factory-->>Init: Object Instance
    Init->>Obj: WithReadFS(FS)
    opt FS is appendfs.FS
        Init->>Obj: WithWriteFS(FS)
    end
    end

    rect rgb(220, 220, 240)
    Note over Init, Loader: Internal Loading
    Init->>Obj: GetLoader()
    Obj-->>Loader: Loader Instance
    Init->>Loader: Load()
    
    rect rgb(240, 240, 240)
    Note over Loader, EM: Extension Setup
    Loader->>Loader: loadExtensionManager()
    Loader->>FS: Sub("extensions")
    Loader->>EM: LoadExtensionManager(SubFS)
    EM->>FS: Read extension configs
    Loader->>Obj: WithExtensionManager(EM)
    end

    rect rgb(220, 240, 220)
    Note over Loader: Inventory Discovery
    Loader->>Loader: loadInventory()
    Loader->>Loader: findInventoryFile()
    alt OCFL 1.0 or 1.1
        Note over Loader: Must be in root
    else OCFL 2.0 or higher
        Loader->>FS: ReadDir(".")
        Note over Loader: Find in root or latest "vN" folder
    end
    end

    rect rgb(220, 220, 240)
    Note over Loader, Inv: Inventory Loading
    Loader->>Loader: loadInventoryFile(path)
    Loader->>FS: ReadFile(inventory.json)
    Loader->>Loader: unmarshalInventoryData()
    Note over Loader: Validation
    Loader->>FS: ReadFile(inventory.json.[digest])
    Loader->>Loader: Verify checksum
    Loader->>Obj: WithInventory(Inv)
    end

    Loader-->>Init: Success
    end

    Init-->>App: Return Object instance
```

## Description of Steps

1.  **Version Detection**: The process begins by identifying the OCFL version of the object. This is done by reading the "Namaste" file (e.g., `0=ocfl_1.1`) in the object directory via `util.GetObjectVersion()`. The logger is also updated with this version.
2.  **Setup & Instantiation**: 
    *   An `ExtensionFactory` is created.
    *   The `FactoryObject` for the detected OCFL version is initialized and used to create a new `Object` instance.
    *   The `Object` is configured with the read-only filesystem and, if available, the writable filesystem.
3.  **Internal Loading**:
    *   The `initocfl` package calls `Load()` on the `Object`'s `Loader`.
    *   **Extension Setup**: The `Loader` initializes the `ExtensionManager` by reading configuration files from the `extensions/` directory.
    *   **Inventory Discovery**: The `Loader` identifies the location of the `inventory.json`. For OCFL 1.0 and 1.1, this file must be in the object root. For later versions, it may also search in the latest version directory.
    *   **Inventory Loading**: The `inventory.json` is read, unmarshaled, and its integrity is verified using the sidecar checksum file. Finally, the loaded `Inventory` is associated with the `Object`.
4.  **Completion**: The fully initialized and loaded `Object` instance is returned to the application.
