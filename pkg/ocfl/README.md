# OCFL Core Packages

Dieses Verzeichnis enthält die Kernpakete des `gocfl` Projekts, welche die [OCFL (Oxford Common File Layout) Spezifikation](https://ocfl.io/) implementieren.

## Paketübersicht

Die Funktionalität ist in mehrere spezialisierte Pakete unterteilt, die verschiedene Aspekte der OCFL-Spezifikation abdecken:

### [Extension](./extension/README.md)
Das `extension` Paket bietet Kernschnittstellen und Verwaltungslogik für OCFL-Erweiterungen. Erweiterungen sind der offizielle Weg, um OCFL-Objekte und Storage-Roots um zusätzliche Funktionalitäten zu ergänzen.
- [Extension Interface](./extension/docs/EXTENSION.md)
- [Manager Interface](./extension/docs/MANAGER.md)
- [Factory Interface](./extension/docs/FACTORY.md)

### [Factory](./factory/README.md)
Das `factory` Modul bietet einen vereinheitlichten Factory-Mechanismus, um Komponenten korrekt gemäß den verschiedenen OCFL-Spezifikationsversionen (z. B. 1.0, 1.1, 2.0) zu instanziieren.

### [Functions](./functions/)
Enthält High-Level-Orchestrierungsfunktionen für gängige Aufgaben wie das Laden, Erstellen, Überprüfen oder Extrahieren von OCFL-Objekten. Es dient als primäre API für viele Anwendungsfälle.

### [Inventory](./inventory/README.md)
Dieses Paket definiert die Strukturen und Schnittstellen für das OCFL Inventory (`inventory.json`). Es ist das Herzstück eines OCFL-Objekts und enthält alle Metadaten über Versionen, Dateien und Fixity-Informationen.
- [Inventory](./inventory/docs/INVENTORY.md)
- [Version](./inventory/docs/VERSION.md)
- [Manifest](./inventory/docs/MANIFEST.md)
- [Fixity](./inventory/docs/FIXITY.md)

### [Object](./object/README.md)
Verwaltet die High-Level-Operationen von OCFL-Objekten. Während das `inventory` Paket sich auf Datenstrukturen konzentriert, kümmert sich `object` um das Laden, Initialisieren, Aktualisieren und Validieren von Objekten.
- [Object Interface](./object/docs/OBJECT.md)
- [Funktionale Module](./object/docs/MODULES.md) (Loader, Initializer, etc.)

### [StorageRoot](./storageroot/README.md)
Verwaltet die OCFL Storage Root, die übergeordnete Struktur, die OCFL-Objekte enthält. Es kümmert sich um das Storage Layout und Root-Level-Erweiterungen.
- [StorageRoot Interface](./storageroot/docs/STORAGEROOT.md)

### [Validation](./validation/README.md)
Bietet Strukturen und Funktionen zur Handhabung von OCFL-Validierungsfehlern und -warnungen gemäß der Spezifikation.

---

## Hilfspakete

- **[ocflerrors](./ocflerrors/errors.go)**: Zentrale Definition von OCFL-spezifischen Fehlertypen.
- **[util](./util/helper.go)**: Hilfsfunktionen für OCFL-Operationen (z. B. Versionserkennung im Dateisystem).
- **[version](./version/version.go)**: Definition der unterstützten OCFL-Versionen.

---
- [Zurück zur übergeordneten Paketdokumentation](../README.md)
- [Zurück zur Projekthauptseite](../../README.md)
