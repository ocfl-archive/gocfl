package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/je4/filesystem/v4/pkg/vfsrw"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
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
	loadedObj, _ := ReloadObject(t, env, objID)
	inv := loadedObj.GetInventory()
	require.NotNil(t, inv)

	vfs, err := vfsrw.NewFS(vfsrw.Config{}, env.Logger)
	require.NoError(t, err)
	err = vfsrw.AddLocal(vfs, nil)
	require.NoError(t, err)

	for vNum := range inv.GetVersions().GetVersionNumbers() {
		t.Run(vNum.String(), func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "gocfl_extract_test_"+vNum.String())
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			destFS, err := appendfs.Sub(appendfs.FS(vfs), filepath.ToSlash(tempDir))
			require.NoError(t, err)

			extractor := loadedObj.GetExtractor().WithDestFS(destFS)
			err = extractor.Extract(vNum, true, "")
			require.NoError(t, err)

			// Verifizierung basierend auf der Versionsnummer
			switch vNum.String() {
			case "v1":
				require.FileExists(t, filepath.Join(tempDir, testFileName1))
				content, _ := os.ReadFile(filepath.Join(tempDir, testFileName1))
				require.Equal(t, testContent1, content)
				require.NoFileExists(t, filepath.Join(tempDir, testFileName2))
				require.NoFileExists(t, filepath.Join(tempDir, testFileName1Renamed))
			case "v2":
				require.FileExists(t, filepath.Join(tempDir, testFileName1Renamed))
				content1, _ := os.ReadFile(filepath.Join(tempDir, testFileName1Renamed))
				require.Equal(t, testContent1, content1)

				require.FileExists(t, filepath.Join(tempDir, testFileName2))
				content2, _ := os.ReadFile(filepath.Join(tempDir, testFileName2))
				require.Equal(t, testContent2, content2)

				require.NoFileExists(t, filepath.Join(tempDir, testFileName1))
			case "v3":
				require.FileExists(t, filepath.Join(tempDir, testFileName1Renamed))
				require.FileExists(t, filepath.Join(tempDir, testFileName3))
				content3, _ := os.ReadFile(filepath.Join(tempDir, testFileName3))
				require.Equal(t, testContent3, content3)

				require.NoFileExists(t, filepath.Join(tempDir, testFileName2))
			}

			// Verifizierung des Manifests
			require.FileExists(t, filepath.Join(tempDir, "manifest.sha512"))
		})
	}
}
