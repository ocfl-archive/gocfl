# Factory Initialization and Extension Registration Sequence Diagram

This diagram shows how extensions register themselves and how the `Factory` is initialized, collecting these registrations.

```mermaid
sequenceDiagram
    participant App as Application
    participant Ext as Extension (pkg/extensions/...)
    participant Reg as Registration (pkg/ocfl/extension/registration.go)
    participant FactImpl as Factory (pkg/ocfl/extension/extensionimpl/factory.go)

    Note over Ext: init() function
    Ext->>Reg: RegisterExtensionStorageRoot/Object(name, builder, ...)
    Reg-->>Reg: Add to global registration map

    App->>FactImpl: NewFactory[T](params, logger)
    activate FactImpl
    FactImpl->>Reg: RegisterWithFactory(factory, logger)
    activate Reg
    loop for each entry in registration map
        Reg->>Reg: Check if type matches T
        Reg->>FactImpl: RegisterExtension(name, builder, ...)
        FactImpl-->>FactImpl: Add to factory.extension map
    end
    deactivate Reg
    FactImpl-->>App: return factory
    deactivate FactImpl
```

## Description

1.  **Self-Registration**: Each extension in `pkg/extensions/...` has an `init()` function that calls `extension.RegisterExtensionStorageRoot` or `extension.RegisterExtensionObject`. These functions add the extension's builder to a global map in the `registration.go` file.
2.  **Factory Creation**: When the application creates a new factory using `extensionimpl.NewFactory[T]`, the factory calls `extension.RegisterWithFactory`.
3.  **Registration Transfer**: `RegisterWithFactory` iterates over the global registration map and calls `RegisterExtension` on the factory instance for each extension that matches the requested manager type `T`.
4.  **Ready for Use**: The factory instance now contains all relevant extensions and can be used to load extensions from configurations or files.
