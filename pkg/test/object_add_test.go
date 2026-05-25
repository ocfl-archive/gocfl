package test

import (
	"testing"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/stretchr/testify/assert"
)

func TestObjectAdd(t *testing.T) {
	testObjectAdd(t, version.Version1_1)
}

func TestObjectAdd20(t *testing.T) {
	testObjectAdd(t, version.Version2_0)
}

func testObjectAdd(t *testing.T, ocflVer version.OCFLVersion) {
	env := SetupTestEnv(t, ocflVer)
	objID := "test-object"
	obj, _ := CreateTestObject(t, env, objID, ocflVer)

	// 5. Version hinzufügen
	vw, err := obj.StartUpdate("initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	assert.NoError(t, err)

	testFileName := "test.txt"
	testContent := []byte("Hello OCFL")
	err = vw.AddData(testContent, testFileName, false, "", false, false)
	assert.NoError(t, err)

	err = vw.Close()
	assert.NoError(t, err)

	// 6. Verifizierung
	// Wir laden das Objekt neu über ein Lese-Dateisystem, da das Schreib-Dateisystem (appendfs)
	// nach dem Schreiben nicht für Stat() o.ä. verwendet werden darf (z.B. bei ZIP).
	loadedObj, _, loadedObjeCloser := ReloadObject(t, env, objID)
	defer loadedObjeCloser.Close()

	// Objekt validieren
	v := loadedObj.GetValidator()
	validationErr := v.Validate()
	if ocflVer != version.Version2_0 {
		assert.NoError(t, validationErr)
	}

	validationErrors := env.OCFLLogger.ValidationErrors()
	var errCount, warnCount int
	for _, vErr := range validationErrors {
		if vErr.Code[0] == 'W' {
			t.Logf("Warning: %s", vErr.Error())
			warnCount++
		} else {
			t.Errorf("Error: %s", vErr.Error())
			errCount++
		}
	}
	t.Logf("Validation Summary: %d errors, %d warnings", errCount, warnCount)
	if errCount > 0 {
		t.Errorf("Validation failed with %d errors", errCount)
	}

	// Inventory abrufen
	inv := loadedObj.GetInventory()
	assert.NotNil(t, inv)

	// Prüfe ID
	assert.Equal(t, objID, inv.GetID())

	// Prüfe Head (sollte v1 sein)
	assert.Equal(t, "v1", inv.GetHead().String())

	// Prüfe Datei im State von v1
	foundTest := false
	err = inv.IterateFiles(inv.GetHead(), func(internal []string, external []string, digest string) error {
		for _, ext := range external {
			if ext == testFileName {
				foundTest = true
			}
		}
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, foundTest, "test.txt should be in v1 state")

}
