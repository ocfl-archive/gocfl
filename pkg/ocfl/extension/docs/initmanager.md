# Extension Manager Initialization Sequence Diagram

This diagram shows how the `Factory` loads an `ExtensionManager` from a filesystem (typically the `extensions` folder in an OCFL object or storage root), including the iteration over the filesystem and the resolution of the manager and its extensions.

```mermaid
sequenceDiagram
    participant ExtFact as extensionFactory (extension.Factory[T])
    participant Fact as FactoryImpl (extensionimpl/factory.go)
    participant FS as fs.FS (Extensions Folder)
    participant Ext as Extension Instance
    participant Initial as Initial Extension
    participant Manager as Manager Instance (T)

    ExtFact->>Fact: LoadExtensionManager(fsys)
    activate Fact

    Note over Fact: 1. Load all extension folders
    Fact->>Fact: loadExtensions(fsys)
    activate Fact
    Fact->>FS: ReadDir(".")
    loop for each directory
        Fact->>Fact: LoadExtensionFile(subfs)
        Fact->>Ext: Load(config.json, subfs)
        Fact-->>Fact: Append to extensions list
    end
    deactivate Fact

    Note over Fact: 2. Extract "initial" extension
    Fact->>Fact: extractInitial(extensions)
    alt initial found
        Fact-->>Fact: use found initial
    else initial not found
        Fact->>Fact: loadDefaultInitial()
        Fact-->>Fact: use default initial
    end

    Note over Fact: 3. Extract Manager extension
    Fact->>Initial: GetExtension()
    Note right of Initial: Returns the name of the<br/>manager extension
    Initial-->>Fact: managerName
    Note over Fact: Search managerName in extensions list
    Fact->>Fact: extractManager(managerName, extensions)
    alt manager found in list
        Fact-->>Fact: use found manager
    else manager not found in list
        Fact->>Fact: loadDefaultManager(managerName)
        Fact-->>Fact: use default manager
    end

    Note over Fact: 4. Add remaining extensions to Manager
    loop for each remaining extension
        Fact->>Manager: Add(ext)
    end

    Fact->>Manager: Finalize()
    Fact->>Manager: SetInitial(initial)
    Fact-->>ExtFact: return manager
    deactivate Fact
```

## Description

0.  **Entry Point**: Der Aufrufer (z.B. ein `Loader` für ein OCFL Objekt) ruft `extensionFactory.LoadExtensionManager` mit dem `extensions` Unterdateisystem auf.
1.  **Iterate Filesystem**: Die Methode `loadExtensions` liest das bereitgestellte Dateisystem (normalerweise das `extensions` Verzeichnis). Für jedes Unterverzeichnis wird versucht, eine Extension mittels `LoadExtensionFile` zu laden, welche die `config.json` in diesem Verzeichnis liest.
2.  **Initial Extension**: The system looks for an extension named `initial`. This extension (defined in OCFL) specifies which manager should be used (it contains the extension name of the manager). If not present, a default initial configuration is used.
3.  **Manager Selection**: Based on the name provided by the `initial` extension (via `initial.GetExtension()`), the factory searches for this name within the list of already loaded extensions (`extractManager`). If a matching extension is found, it is used as the "Manager". This manager will hold and coordinate all other extensions. If no such extension is found in the filesystem, a default manager of that name is created.
4.  **Assembly**: All other loaded extensions are added to the manager via its `Add` method. Finally, the manager is finalized and returned to the application.
