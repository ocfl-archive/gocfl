package test

import (
	"testing"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/stretchr/testify/require"
)

func TestObjectUpdate(t *testing.T) {
	testObjectUpdate(t, version.Version1_1)
}

func TestObjectUpdate20(t *testing.T) {
	testObjectUpdate(t, version.Version2_0)
}

func testObjectUpdate(t *testing.T, ocflVer version.OCFLVersion) {
	env := SetupTestEnv(t, ocflVer)
	objID := "test-object"
	obj, _ := CreateTestObject(t, env, objID, ocflVer)

	// 5. Erste Version hinzufügen (v1)
	vw1, err := obj.StartUpdate("initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName1 := "test1.txt"
	testContent1 := []byte("Hello OCFL v1")
	err = vw1.AddData(testContent1, testFileName1, false, "", false, false)
	require.NoError(t, err)

	err = vw1.Close()
	require.NoError(t, err)

	// 6. Zweite Version hinzufügen (v2)
	vw2, err := obj.StartUpdate("second version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	// Neue Datei hinzufügen
	testFileName2 := "test2.txt"
	testContent2 := []byte("Hello OCFL v2")
	err = vw2.AddData(testContent2, testFileName2, false, "", false, false)
	require.NoError(t, err)

	// Bestehende Datei ändern (OCFL erlaubt das Ersetzen durch einfaches AddData mit gleichem Pfad)
	testContent1Updated := []byte("Hello OCFL v1 (updated in v2)")
	err = vw2.AddData(testContent1Updated, testFileName1, false, "", false, false)
	require.NoError(t, err)

	err = vw2.Close()
	require.NoError(t, err)

	// 7. Dritte Version hinzufügen (v3): Rename, Delete, Update
	vw3, err := obj.StartUpdate("third version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	// Rename: test2.txt -> test2_renamed.txt
	err = vw3.RenameFile(testFileName2, "test2_renamed.txt", "")
	require.NoError(t, err)

	// Delete: test1.txt
	err = vw3.DeleteFile(testFileName1, "")
	require.NoError(t, err)

	// Update (AddData): test3.txt hinzufügen
	testFileName3 := "test3.txt"
	testContent3 := []byte("Hello OCFL v3")
	err = vw3.AddData(testContent3, testFileName3, false, "", false, false)
	require.NoError(t, err)

	err = vw3.Close()
	require.NoError(t, err)

	// 8. Verifizierung
	// Wir laden das Objekt neu über ein Lese-Dateisystem
	loadedObj, _, loadedObjCloser := ReloadObject(t, env, objID)
	defer loadedObjCloser.Close()

	// Validierung des Objekts
	validator := loadedObj.GetValidator()
	err = validator.Validate()
	require.NoError(t, err, "Object validation should pass")

	validationErrors := env.OCFLLogger.ValidationErrors()
	for _, vErr := range validationErrors {
		if vErr.Code[0] == 'W' {
			t.Logf("Warning: %s", vErr.Error())
		} else {
			t.Errorf("Error: %s", vErr.Error())
		}
	}

	// Inventory abrufen
	inv := loadedObj.GetInventory()
	require.NotNil(t, inv)

	// Prüfe ID
	require.Equal(t, objID, inv.GetID())

	// Prüfe Head (sollte v3 sein)
	require.Equal(t, "v3", inv.GetHead().String())

	// Prüfe Dateien im State von v3
	foundTest1 := false
	foundTest2Original := false
	foundTest2Renamed := false
	foundTest3 := false
	err = inv.IterateFiles(inv.GetHead(), func(internal []string, external []string, digest string) error {
		for _, ext := range external {
			if ext == testFileName1 {
				foundTest1 = true
			}
			if ext == testFileName2 {
				foundTest2Original = true
			}
			if ext == "test2_renamed.txt" {
				foundTest2Renamed = true
			}
			if ext == testFileName3 {
				foundTest3 = true
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.False(t, foundTest1, "test1.txt should NOT be in v3 state (deleted)")
	require.False(t, foundTest2Original, "test2.txt should NOT be in v3 state under original name (renamed)")
	require.True(t, foundTest2Renamed, "test2_renamed.txt should be in v3 state")
	require.True(t, foundTest3, "test3.txt should be in v3 state")
}
