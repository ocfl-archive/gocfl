package test

import (
	"io/fs"
	"testing"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/stretchr/testify/assert"
)

func TestStorageRootInit(t *testing.T) {
	env := SetupTestEnv(t)

	// 4. Verifizierung
	// Wir nutzen env.ReadSRFS direkt als Lese-Dateisystem, da der Storage Root bereits initialisiert wurde.
	ocflVer := version.Version1_1
	namaste := "0=ocfl_" + string(ocflVer)
	_, err := fs.Stat(env.ReadSRFS, namaste)
	assert.NoError(t, err, "Namaste file should exist")

	// Überprüfe ob OCFL Spezifikation kopiert wurde (aus gocfl/docs)
	_, err = fs.Stat(env.ReadSRFS, "ocfl_spec_1.1.md")
	assert.NoError(t, err, "OCFL specification doc should exist")

}
