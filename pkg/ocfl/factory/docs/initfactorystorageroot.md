# Factory Storage Root Initialization Sequence Diagram

Dieses Dokument beschreibt den Initialisierungsprozess der Unified Factory für Storage Roots.

## NewFactoryStorageRoot Sequence Diagram

Das folgende Diagramm zeigt den Ablauf beim Aufruf von `initocfl.NewFactoryStorageRoot` für alle unterstützten Versionen.

```mermaid
sequenceDiagram
    participant App as Application
    participant Init as initocfl (initocfl/factory.go)

    box "factoryimpl"
    participant Impl10 as factoryStorageRoot10
    participant Impl11 as factoryStorageRoot11
    participant Impl20 as factoryStorageRoot20
    end

    App->>Init: NewFactoryStorageRoot(ver, extensionFactory, logger)
    activate Init

    alt ver == Version1_0
        Init->>Impl10: NewFactoryStorageRoot10(extFact, logger)
        activate Impl10
        Note over Impl10: Initialisiert Basis-Komponente
        Impl10-->>Init: return factoryStorageRoot10
        deactivate Impl10
    else ver == Version1_1 (Default)
        Init->>Impl11: NewFactoryStorageRoot11(extFact, logger)
        activate Impl11
        Note over Impl11: Initialisiert Basis-Komponente
        Impl11-->>Init: return factoryStorageRoot11
        deactivate Impl11
    else ver == Version2_0
        Init->>Impl20: NewFactoryStorageRoot20(extFact, logger)
        activate Impl20
        Note over Impl20: Initialisiert Basis-Komponente
        Impl20-->>Init: return factoryStorageRoot20
        deactivate Impl20
    end

    Init-->>App: return FactoryStorageRoot
    deactivate Init
```

## Beschreibung

1.  **Entry Point**: Die Anwendung ruft `initocfl.NewFactoryStorageRoot` auf. Dabei werden die gewünschte OCFL-Version, eine `extension.Factory` für Storage Roots und ein Logger übergeben.
2.  **Version Dispatch**: Basierend auf der übergebenen Version delegiert `initocfl` den Aufruf an eine spezifische Implementierungsfunktion im Paket `factoryimpl` (z. B. `NewFactoryStorageRoot11` für Version 1.1).
3.  **Base Initialization**: Die spezifische Implementierung initialisiert die Basis-Komponente (via `NewFactoryBaseStorageRoot`).
4.  **Composition**: Die Basis-Komponente wird in ein versionsspezifisches Struct eingebettet, das das Unified Interface `FactoryStorageRoot` implementiert.
