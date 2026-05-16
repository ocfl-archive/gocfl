package test

import (
	"io/fs"
	"testing"

	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectExtract(t *testing.T) {
	env := SetupTestEnv(t)
	objID := "test-extract-object"
	obj, _ := CreateTestObject(t, env, objID)

	// 1. Version hinzufügen (v1)
	vw1, err := obj.StartUpdate("initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName1 := "test1.txt"
	testContent1 := []byte("Hello OCFL Extraction v1")
	err = vw1.AddData(testContent1, testFileName1, false, "", false, false)
	require.NoError(t, err)

	err = vw1.Close()
	require.NoError(t, err)

	// 2. Version hinzufügen (v2): Add & Rename
	vw2, err := obj.StartUpdate("v2: add and rename", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName2 := "test2.txt"
	testContent2 := []byte("Hello OCFL Extraction v2")
	err = vw2.AddData(testContent2, testFileName2, false, "", false, false)
	require.NoError(t, err)

	testFileName1Renamed := "test1_renamed.txt"
	err = vw2.RenameFile(testFileName1, testFileName1Renamed, "")
	require.NoError(t, err)

	err = vw2.Close()
	require.NoError(t, err)

	// 3. Version hinzufügen (v3): Add & Delete
	vw3, err := obj.StartUpdate("v3: add and delete", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName3 := "test3.txt"
	testContent3 := []byte("Hello OCFL Extraction v3")
	err = vw3.AddData(testContent3, testFileName3, false, "", false, false)
	require.NoError(t, err)

	err = vw3.DeleteFile(testFileName2, "")
	require.NoError(t, err)

	err = vw3.Close()
	require.NoError(t, err)

	// 4. Extraktion für alle Versionen durchführen und prüfen
	loadedObj, _, loadedObjCloser := ReloadObject(t, env, objID)
	defer loadedObjCloser.Close()

	inv := loadedObj.GetInventory()
	require.NotNil(t, inv)

	vfs := env.DestFS

	for vNum := range inv.GetVersions().GetVersionNumbers() {
		t.Run(vNum.String(), func(t *testing.T) {
			extractPath := "vfs://testmem/extract_" + vNum.String()
			destFS, err := appendfs.Sub(appendfs.FS(vfs), extractPath)
			require.NoError(t, err)

			extractor := loadedObj.GetExtractor().WithDestFS(destFS)
			err = extractor.Extract(vNum, true, "")
			require.NoError(t, err)

			// Verifizierung basierend auf der Versionsnummer
			switch vNum.String() {
			case "v1":
				// testFileName1
				_, err = fs.Stat(destFS, testFileName1)
				assert.NoError(t, err)
				content, err := fs.ReadFile(destFS, testFileName1)
				assert.NoError(t, err)
				assert.Equal(t, testContent1, content)

				// testFileName2 (should not exist)
				_, err = fs.Stat(destFS, testFileName2)
				assert.Error(t, err)

				// testFileName1Renamed (should not exist)
				_, err = fs.Stat(destFS, testFileName1Renamed)
				assert.Error(t, err)

			case "v2":
				// testFileName1Renamed
				_, err = fs.Stat(destFS, testFileName1Renamed)
				assert.NoError(t, err)
				content1, err := fs.ReadFile(destFS, testFileName1Renamed)
				assert.NoError(t, err)
				assert.Equal(t, testContent1, content1)

				// testFileName2
				_, err = fs.Stat(destFS, testFileName2)
				assert.NoError(t, err)
				content2, err := fs.ReadFile(destFS, testFileName2)
				assert.NoError(t, err)
				assert.Equal(t, testContent2, content2)

				// testFileName1 (should not exist)
				_, err = fs.Stat(destFS, testFileName1)
				assert.Error(t, err)

			case "v3":
				// testFileName1Renamed
				_, err = fs.Stat(destFS, testFileName1Renamed)
				assert.NoError(t, err)

				// testFileName3
				_, err = fs.Stat(destFS, testFileName3)
				assert.NoError(t, err)
				content3, err := fs.ReadFile(destFS, testFileName3)
				assert.NoError(t, err)
				assert.Equal(t, testContent3, content3)

				// testFileName2 (should not exist)
				_, err = fs.Stat(destFS, testFileName2)
				assert.Error(t, err)
			}

			// Verifizierung des Manifests
			_, err = fs.Stat(destFS, "manifest.sha512")
			assert.NoError(t, err)
		})
	}
}
