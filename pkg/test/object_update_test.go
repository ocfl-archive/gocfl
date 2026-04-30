package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectUpdate(t *testing.T) {
	env := SetupTestEnv(t)
	objID := "test-object"
	obj, objFS := CreateTestObject(t, env, objID)

	// 5. Erste Version hinzufügen (v1)
	vw1, err := obj.StartUpdate(objFS, "initial version", "Junie", "junie@jetbrains.com", false)
	assert.NoError(t, err)

	testFileName1 := "test1.txt"
	testContent1 := []byte("Hello OCFL v1")
	err = vw1.AddData(testContent1, testFileName1, false, "", false, false)
	assert.NoError(t, err)

	err = vw1.Close()
	assert.NoError(t, err)

	// 6. Zweite Version hinzufügen (v2)
	vw2, err := obj.StartUpdate(objFS, "second version", "Junie", "junie@jetbrains.com", false)
	assert.NoError(t, err)

	// Neue Datei hinzufügen
	testFileName2 := "test2.txt"
	testContent2 := []byte("Hello OCFL v2")
	err = vw2.AddData(testContent2, testFileName2, false, "", false, false)
	assert.NoError(t, err)

	// Bestehende Datei ändern (OCFL erlaubt das Ersetzen durch einfaches AddData mit gleichem Pfad)
	testContent1Updated := []byte("Hello OCFL v1 (updated in v2)")
	err = vw2.AddData(testContent1Updated, testFileName1, false, "", false, false)
	assert.NoError(t, err)

	err = vw2.Close()
	assert.NoError(t, err)

	// 7. Verifizierung
	// Wir laden das Objekt neu über ein Lese-Dateisystem, da das Schreib-Dateisystem (appendfs)
	// nach dem Schreiben nicht für Stat() o.ä. verwendet werden darf (z.B. bei ZIP).
	loadedObj, _ := ReloadObject(t, env, objID)

	// Inventory abrufen
	inv := loadedObj.GetInventory()
	assert.NotNil(t, inv)

	// Prüfe ID
	assert.Equal(t, objID, inv.GetID())

	// Prüfe Head (sollte v2 sein)
	assert.Equal(t, "v2", inv.GetHead().String())

	// Prüfe Dateien im State von v2
	foundTest1 := false
	foundTest2 := false
	err = inv.IterateFiles(inv.GetHead(), func(internal []string, external []string, digest string) error {
		for _, ext := range external {
			if ext == testFileName1 {
				foundTest1 = true
			}
			if ext == testFileName2 {
				foundTest2 = true
			}
		}
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, foundTest1, "test1.txt should be in v2 state")
	assert.True(t, foundTest2, "test2.txt should be in v2 state")
}
