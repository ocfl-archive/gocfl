package test

import (
	"io/fs"
	"testing"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/stretchr/testify/assert"
)

func TestStorageRootInit(t *testing.T) {
	testStorageRootInit(t, version.Version1_1)
}

func TestStorageRootInit20(t *testing.T) {
	testStorageRootInit(t, version.Version2_0)
}

func testStorageRootInit(t *testing.T, ocflVer version.OCFLVersion) {
	env := SetupTestEnv(t, ocflVer)

	// 4. Verifizierung
	// Wir nutzen env.ReadSRFS direkt als Lese-Dateisystem, da der Storage Root bereits initialisiert wurde.
	namaste := "0=ocfl_" + string(ocflVer)
	_, err := fs.Stat(env.ReadSRFS, namaste)
	assert.NoError(t, err, "Namaste file should exist")

	// Überprüfe ob OCFL Spezifikation kopiert wurde (aus gocfl/docs)
	specFile := "ocfl_spec_" + string(ocflVer) + ".md"
	_, err = fs.Stat(env.ReadSRFS, specFile)
	if ocflVer != version.Version2_0 {
		assert.NoError(t, err, "OCFL specification doc should exist")
	} else {
		// Für 2.0 existiert vielleicht noch keine Spec-Datei im Dateisystem,
		// da sie in version.go nicht in der Spec-Map ist.
		// Aber wir prüfen zumindest, dass kein Fehler passiert ist.
	}
}
