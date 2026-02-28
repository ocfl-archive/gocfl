# Thumbnail Subsystem

[Zurück zur Übersicht](../README.md)

Das Thumbnail-Subsystem (`pkg/subsystem/thumbnail`) dient zur automatischen Generierung von Vorschaubildern für archivierte Dateien. Es erlaubt die Definition von Verarbeitungsfunktionen, die externe Werkzeuge zur Bildskalierung und -konvertierung aufrufen.

## Funktionsweise

Das Subsystem verwaltet eine Liste von `Functions`, die für bestimmte MIME-Typen konfiguriert werden. Bei der Anforderung eines Thumbnails wird die passende Funktion ausgewählt und ausgeführt.

## Konfiguration

In der GOCFL-Konfiguration können Thumbnail-Funktionen definiert werden:

```toml
[thumbnail]
  background = "#FFFFFF" # Hintergrundfarbe für Transparenzen
  [thumbnail.function.imagemagick]
    title = "ImageMagick Thumbnail"
    id = "imagemagick"
    command = "convert {source}[0] -thumbnail {width}x{height} -background {background} -gravity center -extent {width}x{height} {destination}"
    timeout = 30000000000 # 30 Sekunden
    pronoms = ["fmt/41", "fmt/42"]
    mime = ['^image/.*$', '^application/pdf$']
```

### Platzhalter

Innerhalb des `command`-Strings können Platzhalter verwendet werden, die zur Laufzeit ersetzt werden:

- `{source}`: Pfad zur Quelldatei.
- `{destination}`: Pfad zur Zieldatei (Thumbnail-Ausgabe).
- `{width}`: Zielbreite des Thumbnails.
- `{height}`: Zielhöhe des Thumbnails.
- `{background}`: Die in der Konfiguration definierte Hintergrundfarbe.

## Verwendung

Das Thumbnail-Subsystem wird typischerweise von Extensions genutzt, die Thumbnails für OCFL-Objekte bereitstellen oder verwalten. Es übernimmt die Prozesssteuerung und das Timeout-Handling für die externen Aufrufe.
