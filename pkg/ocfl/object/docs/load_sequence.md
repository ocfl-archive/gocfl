# Object Load Sequence

This document describes the sequence of operations for loading an existing OCFL Object. See the [Object Architecture](architecture.md) for structural details.

## Sequence Diagram

The following diagram illustrates the steps taken by `initocfl.LoadObject()` and the internal `Loader.Load()` method.

```mermaid
sequenceDiagram
    autonumber
    rect rgb(250, 250, 250)
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
    end
```

## Step Descriptions

1.  **Version Detection**: Identifies the OCFL version by reading the Namaste file (e.g., `0=ocfl_1.1`) via `util.GetObjectVersion()`.
2.  **Setup & Instantiation**: 
    *   An `ExtensionFactory` is created.
    *   A version-specific `FactoryObject` creates the `Object` instance.
    *   The `Object` is configured with the read-only and (if available) writable filesystems.
3.  **Internal Loading**:
    *   `Load()` is called on the `Loader`.
    *   **Extensions**: The `ExtensionManager` is initialized from the `extensions/` directory.
    *   **Inventory Discovery**: The `inventory.json` is located based on the OCFL version.
    *   **Inventory Loading**: The inventory is read, unmarshaled, and verified against its sidecar checksum.
4.  **Completion**: The fully loaded `Object` instance is returned.
