package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/vfsrw"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

func main() {
	// --- Schritt 1: Parsen der Befehlszeilenparameter ---
	pathPtr := flag.String("path", "", "Pfad zum OCFL Storage Root")
	idPtr := flag.String("id", "my-object-id", "OCFL Objekt-ID")
	flag.Parse()

	if *pathPtr == "" {
		fmt.Println("Verwendung: go run main.go -path <storage_root_path> [-id <object_id>]")
		os.Exit(1)
	}

	// --- Schritt 2: Einrichten der Logging-Infrastruktur ---
	ctx := context.Background()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := initocfl.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

	// --- Schritt 3: Konfiguration des Virtuellen Dateisystems (VFS) ---
	cfg := vfsrw.Config{}
	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	if err != nil {
		log.Fatalf("VFS konnte nicht erstellt werden: %v", err)
	}
	defer vfs.Close()

	if err := vfsrw.AddLocal(vfs, &vfsrw.ZipAsFolder{
		Enabled:   true,
		Digests:   []checksum.DigestAlgorithm{checksum.DigestSHA512},
		CacheSize: 3,
		Compress:  false,
		ReadOnly:  false,
	}); err != nil {
		logger.Fatal().Err(err).Msg("Lokales Dateisystem konnte nicht hinzugefügt werden")
	}

	srPath := writefs.RealPath(vfs, *pathPtr)
	storageRootFS, closer, err := appendfs.Sub(vfs, srPath)
	if err != nil {
		log.Fatalf("Sub-FS für Storage Root '%s' konnte nicht erstellt werden: %v", srPath, err)
	}
	defer closer.Close()

	// --- Schritt 4: Initialisierung des OCFL Storage Roots ---
	ocflVer := version.Version1_1
	sr, err := initocfl.InitStorageRoot(ctx, storageRootFS, nil, ocflVer, checksum.DigestSHA512, nil, logger)
	if err != nil {
		log.Fatalf("Storage Root bei '%s' konnte nicht initialisiert werden: %v", srPath, err)
	}
	fmt.Printf("OCFL Storage Root erfolgreich bei '%s' initialisiert.\n", srPath)

	// --- Schritt 5: Bestimmung des Objektpfads innerhalb des Storage Roots ---
	objID := *idPtr
	objFolder, err := sr.IdToFolder(objID)
	if err != nil {
		log.Fatalf("Ordner für ID '%s' konnte nicht ermittelt werden: %v", objID, err)
	}

	// --- Schritt 6: Initialisierung des OCFL-Objekts ---
	objFS, closer, err := appendfs.Sub(storageRootFS, objFolder)
	if err != nil {
		log.Fatalf("Sub-FS für Objektordner '%s' konnte nicht erstellt werden: %v", objFolder, err)
	}
	defer closer.Close()

	obj, err := initocfl.InitObject(ctx, objFS, nil, ocflVer, objID, checksum.DigestSHA512, nil, logger)
	if err != nil {
		log.Fatalf("Objekt '%s' bei '%s' konnte nicht initialisiert werden: %v", objID, objFolder, err)
	}
	fmt.Printf("OCFL-Objekt '%s' erfolgreich im Ordner '%s' erstellt.\n", objID, objFolder)

	// --- Schritt 7: Dateien zum Objekt hinzufügen ---
	vw, err := obj.StartUpdate("Erste Version mit Beispieldateien", "GOCFL-Example", "mailto:gocfl@example.com", false)
	if err != nil {
		log.Fatalf("Update für Objekt '%s' konnte nicht gestartet werden: %v", objID, err)
	}

	filesToAdd := map[string]string{
		"README.md":     "# Mein OCFL Objekt\nDies ist ein automatisch erstelltes Objekt.",
		"data/info.txt": "Beispielinhalt für die Datendatei.",
	}

	for path, content := range filesToAdd {
		err = vw.AddData([]byte(content), path, true, "", false, false)
		if err != nil {
			log.Fatalf("Datei '%s' konnte nicht zum Objekt hinzugefügt werden: %v", path, err)
		}
		fmt.Printf("Datei hinzugefügt: %s\n", path)
	}

	// Änderungen übernehmen und Version abschließen
	if err := vw.Close(); err != nil {
		log.Fatalf("VersionWriter konnte nicht geschlossen werden: %v", err)
	}

	// Validierungsfehler prüfen
	vErrors := logger.ValidationErrors()
	if len(vErrors) > 0 {
		fmt.Printf("\nValidierungsprobleme für Objekt '%s' gefunden:\n", objID)
		for _, vErr := range vErrors {
			fmt.Printf("- %s\n", vErr.Error())
		}
	}

	fmt.Printf("Erfolgreich 2 Dateien zum OCFL-Objekt '%s' hinzugefügt.\n", objID)
}
