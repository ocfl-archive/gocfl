# Factory Object Initialization Sequence Diagram

Dieses Dokument beschreibt den Initialisierungsprozess der Unified Factory für Objekte.

## NewFactoryObject Sequence Diagram

Das folgende Diagramm zeigt den Ablauf beim Aufruf von `ocfl.NewFactoryObject` für alle unterstützten Versionen.

```mermaid
sequenceDiagram
    participant App as Application
    participant Init as ocfl (ocfl/factory.go)
    
    box "factoryimpl"
    participant Impl10 as factoryObject10
    participant Impl11 as factoryObject11
    participant Impl20 as factoryObject20
    end
    
    App->>Init: NewFactoryObject(ver, extensionFactory, logger)
    activate Init
    
    alt ver == Version1_0
        Init->>Impl10: NewFactoryObject10(extFact, logger)
        activate Impl10
        Note over Impl10: Initialisiert Basis-Komponente
        Impl10-->>Init: return factoryObject10
        deactivate Impl10
    else ver == Version1_1 (Default)
        Init->>Impl11: NewFactoryObject11(extFact, logger)
        activate Impl11
        Note over Impl11: Initialisiert Basis-Komponente
        Impl11-->>Init: return factoryObject11
        deactivate Impl11
    else ver == Version2_0
        Init->>Impl20: NewFactoryObject20(extFact, logger)
        activate Impl20
        Note over Impl20: Initialisiert Basis-Komponente
        Impl20-->>Init: return factoryObject20
        deactivate Impl20
    end
    
    Init-->>App: return FactoryObject
    deactivate Init
```

## Beschreibung

1.  **Entry Point**: Die Anwendung ruft `ocfl.NewFactoryObject` auf. Dabei werden die gewünschte OCFL-Version, eine `extension.Factory` für Objekte und ein Logger übergeben.
2.  **Version Dispatch**: Basierend auf der übergebenen Version delegiert `ocfl` den Aufruf an eine spezifische Implementierungsfunktion im Paket `factoryimpl` (z. B. `NewFactoryObject11` für Version 1.1).
3.  **Base Initialization**: Die spezifische Implementierung initialisiert die Basis-Komponente (via `NewFactoryBaseObject`).
4.  **Composition**: Die Basis-Komponente wird in ein versionsspezifisches Struct eingebettet, das das Unified Interface `FactoryObject` implementiert.
