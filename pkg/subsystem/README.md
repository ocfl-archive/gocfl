# Subsystem

Das Paket `pkg/subsystem` enthält Implementierungen für spezialisierte Funktionalitäten, die von OCFL-Extensions verwendet werden können. Diese Funktionalitäten sind oft zu umfangreich oder komplex, um direkt in den Extensions implementiert zu werden, oder sie hängen von externen Tools ab.

## Zweck

Subsysteme bieten eine Abstraktionsschicht für externe Prozesse. Extensions können diese Subsysteme nutzen, um Aufgaben wie Dateikonvertierung oder Thumbnail-Erstellung an externe Programme zu delegieren, ohne die Details der Prozesssteuerung selbst implementieren zu müssen.

## Verfügbare Subsysteme

### [Migration](./migration/README.md)
Das Migrations-Subsystem ermöglicht die automatische Konvertierung von Dateien während der Ingest- oder Update-Phase. Es nutzt externe Tools (z.B. ImageMagick, FFmpeg, Pandoc), um Dateien basierend auf konfigurierbaren Strategien zu transformieren.

### [Thumbnail](./thumbnail/README.md)
Das Thumbnail-Subsystem wird zur automatischen Generierung von Vorschaubildern für archivierte Objekte verwendet. Wie das Migrations-Subsystem nutzt es externe Programme zur Bildverarbeitung.

## Konfiguration

Subsysteme werden in der Regel über die zentrale Konfiguration (`gocfl.toml` oder ähnliche) definiert. Sie ordnen bestimmte Datei-Typen (via Mime-Type, Pronom oder Dateiendung) spezifischen Befehlen zu.
