# Object Creation Sequence

This document describes the sequence of operations performed during the creation and initialization of a new OCFL Object. Detailed structural information can be found in the [**Object Architecture Documentation**](architecture.md).

## Sequence Diagram

The following diagram illustrates the steps taken by `initocfl.InitObject()` and the internal `Initializer.Init()` method, followed by adding the first files as seen in the examples.

```mermaid
sequenceDiagram
    participant App as Application
    participant Init as initocfl
    participant FS as AppendFS (Filesystem)
    participant Obj as Object
    participant Factory as FactoryObject
    participant EM as ExtensionManager
    participant Initializer as Initializer
    participant Inv as Inventory
    participant VW as VersionWriter

    App->>Init: InitObject(ctx, FS, ...)
    
    rect rgb(240, 240, 240)
    Note over Init, EM: Setup & Instantiation
    Init->>Init: SetupExtensionManager()
    Init->>Init: NewFactoryObject(Version, ...)
    Init->>Factory: NewObject(ctx)
    Factory-->>Init: Object Instance
    Init->>Obj: WithWriteFS(FS)
    Init->>Obj: WithExtensionManager(EM)
    end

    rect rgb(220, 220, 240)
    Note over Init, Initializer: Initialization
    Init->>Obj: GetInitializer()
    Obj-->>Initializer: Initializer Instance
    Init->>Initializer: Init(ID, Digest, ...)
    
    Initializer->>FS: IsEmpty()
    
    rect rgb(240, 240, 240)
    Note over Initializer, FS: Write OCFL Structure
    Initializer->>FS: WriteFile("0=ocfl_object_...")
    Initializer->>FS: MkDir("extensions")
    Initializer->>FS: Sub("extensions")
    Initializer->>EM: WriteConfig(SubFS)
    end
    
    rect rgb(220, 240, 220)
    Note over Initializer, Inv: Create Inventory
    Initializer->>Factory: NewInventory(ctx)
    Factory-->>Inv: Inventory Instance
    Initializer->>Inv: WithID(ID)
    Initializer->>Inv: WithDigestAlgorithm(Digest)
    Initializer->>Obj: WithInventory(Inv)
    end
    
    Initializer-->>Init: Success
    end
    
    Init-->>App: Return Object instance

    App->>Obj: StartUpdate(msg, ...)
    Obj-->>VW: VersionWriter Instance
    loop for each file/folder
        App->>VW: AddData/AddFile/AddFolder(...)
        VW->>FS: Write files to content directory (vN/content/...)
    end
    App->>VW: Close()
    VW->>Inv: Update state (manifest/state)
    VW->>FS: Write inventory.json & sidecar (once per version)
    VW-->>App: Success
```

## Description of Steps

1.  **Setup & Instantiation**: 
    *   The `initocfl` package sets up the `ExtensionManager` based on provided parameters.
    *   A `FactoryObject` for the desired OCFL version is used to create a new `Object` instance.
    *   The `Object` is configured with the writable filesystem and the `ExtensionManager`.
2.  **Initialization**:
    *   The `initocfl` package calls `Init()` on the `Object`'s `Initializer`.
    *   **Validation**: The `Initializer` verifies that the target directory is empty.
    *   **Write OCFL Structure**: It writes the OCFL object Namaste file (e.g., `0=ocfl_object_1.1`) and initializes the `extensions/` directory with configurations from the `ExtensionManager`.
    *   **Create Inventory**: A new, empty `Inventory` is created with the provided Object ID and digest algorithm, and associated with the `Object`.
3.  **Adding Files (Post-Initialization)**:
    *   To add content, the application calls `StartUpdate()`, which returns a `VersionWriter`.
    *   **File Upload Loop**: Multiple files or folders can be added iteratively using `AddData()`, `AddFile()`, or `AddFolder()`.
    *   Files are written directly to their destination in the version's content directory during the addition process.
    *   Closing the `VersionWriter` finalizes the version, updates the inventory (manifest and state), and writes the `inventory.json` and its sidecar file.

## See Also
- [**Object Architecture**](architecture.md)
- [**Object Load Sequence**](load_sequence.md)
- [**Object Initializer Documentation**](INITIALIZER.md)
- [**Version Writer Documentation**](VERSION_WRITER.md)
