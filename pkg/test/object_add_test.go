package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectAdd(t *testing.T) {
	env := SetupTestEnv(t)
	objID := "test-object"
	obj, objFS := CreateTestObject(t, env, objID)

	// 5. Version hinzufügen
	vw, err := obj.StartUpdate(objFS, "initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
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
	loadedObj, _ := ReloadObject(t, env, objID)

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
