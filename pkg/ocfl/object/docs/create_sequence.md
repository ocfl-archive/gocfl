# Object Creation Sequence

This document outlines the sequence of operations for initializing a new OCFL Object. See the [Object Architecture](architecture.md) for structural details.

## Sequence Diagram

The following diagram illustrates the steps taken by `ocfl.InitObject()` and the internal `Initializer.Init()` method.

```mermaid
%%{init: { 'theme': 'base', 'themeVariables': { 'background': '#ffffff', 'primaryTextColor': '#000000', 'lineColor': '#444444', 'secondaryColor': '#eeeeee', 'tertiaryColor': '#eeeeee', 'actorBkg': '#ffffff', 'actorBorder': '#666666', 'actorTextColor': '#000000', 'noteBkgColor': '#fff5ad', 'noteTextColor': '#000000', 'signalColor': '#444444', 'signalTextColor': '#000000', 'sequenceNumberColor': '#444444', 'loopTextColor': '#000000' } } }%%
sequenceDiagram
    autonumber
    rect rgb(255, 255, 255)
    participant App as Application
    participant Init as ocfl
    participant FS as AppendFS (Filesystem)
    participant Obj as Object
    participant Factory as FactoryObject
    participant EM as ExtensionManager
    participant Initializer as Initializer
    participant Inv as Inventory
    participant VW as VersionWriter
    
    App->>Init: InitObject(ctx, FS, ...)
    
    rect rgb(245, 245, 255)
    Note over Init, EM: Setup & Instantiation
    Init->>Init: SetupExtensionManager()
    Init->>Init: NewFactoryObject(Version, ...)
    Init->>Factory: NewObject(ctx)
    Factory-->>Init: Object Instance
    Init->>Obj: WithWriteFS(FS)
    Init->>Obj: WithExtensionManager(EM)
    end

    rect rgb(245, 255, 245)
    Note over Init, Initializer: Initialization
    Init->>Obj: GetInitializer()
    Obj-->>Initializer: Initializer Instance
    Init->>Initializer: Init(ID, Digest, ...)
    
    Initializer->>FS: IsEmpty()
    
    rect rgb(255, 255, 255)
    Note over Initializer, FS: Write OCFL Structure
    Initializer->>FS: WriteFile("0=ocfl_object_...")
    Initializer->>FS: MkDir("extensions")
    Initializer->>FS: Sub("extensions")
    Initializer->>EM: WriteConfig(SubFS)
    end
    
    rect rgb(255, 250, 240)
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
        VW->>FS: Write files to content directory (v1/content/...)
    end
    App->>VW: Close()
    VW->>Inv: Update state (manifest/state)
    VW->>FS: Write inventory.json & sidecar
    VW-->>App: Success
    end
```

## Step Descriptions

1.  **Setup & Instantiation**: 
    *   The `ocfl` package configures the `ExtensionManager`.
    *   A version-specific `FactoryObject` creates the `Object` instance.
    *   The `Object` is associated with the writable filesystem and the `ExtensionManager`.
2.  **Initialization**:
    *   `Init()` is called on the `Initializer`.
    *   **Validation**: Verifies the target directory is empty.
    *   **Structure**: Writes the OCFL Namaste file and initializes the `extensions/` directory.
    *   **Inventory**: A new `Inventory` is created with the provided ID and digest algorithm.
3.  **Adding Files**:
    *   `StartUpdate()` provides a `VersionWriter`.
    *   Files and folders are added iteratively.
    *   `Close()` finalizes the version, updates the inventory, and writes the `inventory.json`.

## See Also
- [Object Architecture](architecture.md)
- [Object Load Sequence](load_sequence.md)
- [Object Initializer](INITIALIZER.md)
- [Version Writer](VERSION_WRITER.md)
