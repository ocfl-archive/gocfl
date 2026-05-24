# Object Update Sequence

This document describes the sequence of operations for updating an existing OCFL Object. See the [Object Architecture](architecture.md) for structural details.

## Sequence Diagram

The following diagram illustrates the update process using the `VersionWriter`.

```mermaid
%%{init: { 'theme': 'base', 'themeVariables': { 'background': '#ffffff', 'primaryTextColor': '#000000', 'lineColor': '#444444', 'secondaryColor': '#eeeeee', 'tertiaryColor': '#eeeeee', 'actorBkg': '#ffffff', 'actorBorder': '#666666', 'actorTextColor': '#000000', 'noteBkgColor': '#fff5ad', 'noteTextColor': '#000000', 'signalColor': '#444444', 'signalTextColor': '#000000', 'sequenceNumberColor': '#444444', 'loopTextColor': '#000000' } } }%%
sequenceDiagram
    autonumber
    rect rgb(255, 255, 255)
    participant App as Application
    participant Init as initocfl
    participant FS as AppendFS (Filesystem)
    participant Obj as Object
    participant EM as ExtensionManager
    participant VW as VersionWriter
    participant Inv as Inventory

    rect rgb(245, 245, 255)
    Note over App, Obj: 1. Loading existing Object
    App->>Init: LoadObject(ctx, FS, ...)
    Init-->>Obj: Load Sequence (see load_sequence.md)
    Init-->>App: Object Instance
    end

    rect rgb(245, 255, 245)
    Note over App, VW: 2. Starting Update
    App->>Obj: StartUpdate(message, user, ...)
    Obj->>Obj: Check for open writers
    Obj-->>VW: VersionWriter Instance
    end

    rect rgb(255, 255, 255)
    Note over App, FS: 3. Modifying Content (Loop)
    loop For each change
        alt Add/Update File
            App->>VW: AddData(content, path, ...)
            VW->>FS: Write file to vN/content/...
        else Rename File
            App->>VW: RenameFile(src, dest, ...)
            VW->>VW: Update internal state (manifest/state)
        else Delete File
            App->>VW: DeleteFile(path, ...)
            VW->>VW: Update internal state (remove from state)
        end
    end
    end

    rect rgb(255, 250, 240)
    Note over App, Inv: 4. Finalizing Version
    App->>VW: Close()
    VW->>Inv: Finalize new version state
    VW->>FS: Write inventory.json & sidecar
    VW-->>App: Success
    end
    end
```

## Step Descriptions

1.  **Loading existing Object**: Detects the OCFL version and initializes the `ExtensionManager` and `Inventory`. See [Load Sequence](load_sequence.md).
2.  **Starting Update**: `StartUpdate()` is called with commit metadata, returning a `VersionWriter`.
3.  **Modifying Content**: 
    *   **Add/Update**: Data is written directly to the new version's content directory (`vN/content/...`).
    *   **Rename**: Internal state is updated to map the new logical path to the existing content digest.
    *   **Delete**: The logical path is removed from the new version's state.
4.  **Finalizing Version**: `Close()` commits changes, updates the `Inventory`, and writes the `inventory.json` and sidecar files.

## See Also
- [Object Architecture](architecture.md)
- [Object Load Sequence](load_sequence.md)
- [Object Creation Sequence](create_sequence.md)
- [Version Writer](VERSION_WRITER.md)
