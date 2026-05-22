package test

import (
	"io/fs"
	"testing"

	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflactions"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOCFLActions(t *testing.T) {
	testOCFLActions(t, version.Version1_1)
}

func TestOCFLActions20(t *testing.T) {
	testOCFLActions(t, version.Version2_0)
}

func testOCFLActions(t *testing.T, ocflVer version.OCFLVersion) {
	env := SetupTestEnv(t, ocflVer)
	objID := "test-actions-object"
	obj, _ := CreateTestObject(t, env, objID, ocflVer)

	// 1. Version hinzufügen (v1)
	vw, err := obj.StartUpdate("initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName := "test.txt"
	testContent := []byte("Hello OCFL Actions")
	err = vw.AddData(testContent, testFileName, false, "", false, false)
	require.NoError(t, err)

	err = vw.Close()
	require.NoError(t, err)

	// Holen wir uns den Pfad zum Objekt im Storage Root
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	// Wir brauchen ein fs.FS für das Objekt-Verzeichnis für CheckObject
	objFS, err := fs.Sub(env.ReadSRFS, objFolder)
	require.NoError(t, err)

	t.Run("CheckObject", func(t *testing.T) {
		err := ocflactions.CheckObject(t.Context(), objFS, env.OCFLLogger)
		assert.NoError(t, err)
	})

	t.Run("Extract", func(t *testing.T) {
		extractPath := "vfs://testmem/extract_actions"
		destFS, closer, err := appendfs.Sub(appendfs.FS(env.DestFS), extractPath)
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = closer.Close()
		})

		// Wir nutzen env.ReadSRFS als Basis und geben den relativen Pfad zum Objekt an
		err = ocflactions.Extract(t.Context(), env.ReadSRFS, destFS, objFolder, inventory.NewVersionNumber(), true, "", env.OCFLLogger)
		assert.NoError(t, err)

		// Verifizierung
		content, err := fs.ReadFile(destFS, testFileName)
		assert.NoError(t, err)
		assert.Equal(t, testContent, content)

		_, err = fs.Stat(destFS, "manifest.sha512")
		assert.NoError(t, err)
	})

	t.Run("ExtractMeta", func(t *testing.T) {
		meta, err := ocflactions.ExtractMeta(t.Context(), env.ReadSRFS, objFolder, env.OCFLLogger)
		assert.NoError(t, err)
		require.NotNil(t, meta)

		assert.Equal(t, objID, meta.ID)
		assert.Equal(t, "v1", meta.Head.String())
		// Prüfe ob die Datei in Files enthalten ist
		found := false
		for _, fm := range meta.Files {
			for _, names := range fm.VersionName {
				for _, name := range names {
					if name == testFileName {
						found = true
						break
					}
				}
			}
		}
		assert.True(t, found, "test file should be in metadata")
	})
}
