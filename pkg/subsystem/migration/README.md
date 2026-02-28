# Migration Subsystem

[Zurück zur Übersicht](../README.md)

Das Migrations-Subsystem (`pkg/subsystem/migration`) bietet Funktionalitäten für die Migration von Dateiformaten innerhalb eines OCFL-Objekts. Es ermöglicht die Ausführung externer Befehle, um Dateien von einem Format in ein anderes zu konvertieren oder Transformationen anzuwenden.

## Funktionsweise

Das Subsystem verwaltet eine Liste von `Functions`, wobei jede Funktion einem externen Befehl (z.B. `convert`, `ffmpeg`) entspricht. Diese Funktionen werden basierend auf Dateieigenschaften (wie Dateiendungen, reguläre Ausdrücke für Dateinamen oder PRONOM-IDs) ausgewählt.

### Strategien

Es stehen verschiedene Strategien für die Migration zur Verfügung:

- **`replace`**: Die Originaldatei wird im Ziel-OCFL-Objekt durch die migrierte Version ersetzt.
- **`add`**: Die migrierte Version wird zusätzlich zur Originaldatei hinzugefügt.
- **`folder`**: Die migrierte Version wird in einem speziellen Unterordner abgelegt (oft verwendet für versionierte Migrationen).

## Konfiguration

In der GOCFL-Konfiguration können Migrationsfunktionen definiert werden:

```toml
[migration]
  [migration.function.imagemagick]
    title = "ImageMagick Convert"
    id = "imagemagick"
    command = "convert {source} {destination}"
    strategy = "replace"
    filenameRegexp = '\.(jpg|png)$'
    filenameReplacement = '.tiff'
    timeout = 60000000000 # 60 Sekunden
    pronoms = ["fmt/41", "fmt/42"]
```

### Platzhalter

Innerhalb des `command`-Strings können Platzhalter verwendet werden, die zur Laufzeit ersetzt werden:

- `{source}`: Pfad zur Quelldatei.
- `{destination}`: Pfad zur Zieldatei, in die die Ausgabe geschrieben werden soll.

## Verwendung in Extensions

Extensions können das Subsystem nutzen, um eine Migration durchzuführen. Typischerweise wird `DoMigrate` verwendet, welches die temporäre Dateiverwaltung und den Aufruf der Migration übernimmt.
