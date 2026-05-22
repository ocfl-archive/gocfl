# Storage Root Load Sequence

This document describes the sequence of operations performed during the loading of an existing OCFL Storage Root, starting from the high-level entry point.

## Sequence Diagram

The following diagram illustrates the steps taken by `initocfl.LoadStorageRoot()` and the internal `Loader.Load()` method.

```mermaid
sequenceDiagram
    participant App as Application
    participant Init as initocfl
    participant Util as util (Version Discovery)
    participant FS as ReadFS (Filesystem)
    participant SR as StorageRoot
    participant Factory as FactoryStorageRoot
    participant Loader as Loader
    participant EM as ExtensionManager

    App->>Init: LoadStorageRoot(ctx, FS, ...)
    
    rect rgb(220, 240, 220)
    Note over Init, Util: Version Detection
    Init->>Util: GetStorageRootVersion(FS)
    Util->>FS: Read Namaste file (0=ocfl_...)
    Util-->>Init: OCFL Version
    end

    rect rgb(240, 240, 240)
    Note over Init, Factory: Setup & Instantiation
    Init->>Init: NewExtensionFactory()
    Init->>Init: NewFactoryStorageRoot(Version, ExtensionFactory)
    Init->>Factory: NewStorageRoot(ctx)
    Factory-->>Init: StorageRoot Instance
    Init->>SR: WithReadFS(FS)
    Init->>SR: WithDigestAlgorithm(...)
    end

    rect rgb(220, 220, 240)
    Note over Init, EM: Internal Loading
    Init->>SR: GetLoader()
    SR-->>Loader: Loader Instance
    Init->>Loader: Load()
    Loader->>Loader: loadExtensionManager()
    Loader->>FS: Sub("extensions")
    Loader->>EM: LoadExtensionManager(SubFS)
    EM->>FS: Read extension configs
    Loader-->>Init: Success
    end

    Init-->>App: Return StorageRoot instance
```

## Description of Steps

1.  **Version Detection**: The process begins by identifying the OCFL version of the Storage Root. This is done by reading the "Namaste" file (e.g., `0=ocfl_1.1`) in the root directory via `util.GetStorageRootVersion()`.
2.  **Setup & Instantiation**: 
    *   An `ExtensionFactory` is created.
    *   The `FactoryStorageRoot` for the detected OCFL version is initialized and used to create a new `StorageRoot` instance.
    *   The `StorageRoot` is configured with the filesystem and default digest algorithm.
3.  **Internal Loading**: 
    *   The `initocfl` package calls `Load()` on the `StorageRoot`'s `Loader`.
    *   The `Loader` identifies the `extensions/` directory and uses the `ExtensionFactory` to load and initialize the `ExtensionManager` with the configuration files found on disk.
4.  **Completion**: The fully initialized and loaded `StorageRoot` instance is returned to the application.
