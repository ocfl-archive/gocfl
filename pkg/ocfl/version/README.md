# OCFL Version Package

Das Paket `version` verwaltet die unterstützten Versionen des Oxford Common File Layout (OCFL) Standards innerhalb der `gocfl` Bibliothek.

## Funktionen

- Definition von OCFL-Versionen (`1.0`, `1.1`, `2.0`).
- Validierung von Versionsstrings.
- Vergleich von Versionen (z.B. für Migrationspfade oder Kompatibilitätsprüfungen).
- Bereitstellung der OCFL-Spezifikationstexte (eingebettet als Markdown).
- Reguläre Ausdrücke zur Erkennung von OCFL-Strukturdateien (Namaste).

## Konstanten

| Konstante | Wert | Beschreibung |
|-----------|------|--------------|
| `Version1_0` | `"1.0"` | OCFL Standard Version 1.0 |
| `Version1_1` | `"1.1"` | OCFL Standard Version 1.1 |
| `Version2_0` | `"2.0"` | Zukünftige OCFL Version |
| `Default` | `"1.1"` | Standardversion für neue Objekte |

## Verwendung

### Version prüfen
```go
import "github.com/jeese/gocfl/pkg/ocfl/version"

if version.ValidVersion(version.OCFLVersion("1.1")) {
    // Version wird unterstützt
}
```

### Versionen vergleichen
```go
if version.Less(version.Version1_0, version.Version1_1) {
    // 1.0 ist älter als 1.1
}
```
