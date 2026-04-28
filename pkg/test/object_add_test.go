package test

import (
	"os"
	"testing"

	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestObjectAdd(t *testing.T) {
	ctx := t.Context()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

	// 1. Initialisierung des Zieldateisystems (filesystem.vfsrw mit afero memfs)
	cfg := vfsrw.Config{
		"testmem": &vfsrw.VFS{
			Name: "testmem",
			Type: "afero",
			Afero: &vfsrw.Afero{
				BaseDir: "mem://",
			},
		},
	}

	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	assert.NoError(t, err)
	defer vfs.Close()

	destFS, ok := any(vfs).(appendfs.FS)
	assert.True(t, ok, "vfs should implement appendfs.FS")

	// Erstelle Sub-FS für den Storage Root, damit Pfade relativ sind
	srFS, err := appendfs.Sub(destFS, "vfs://testmem/")
	assert.NoError(t, err)

	// 2. Initialisierung der OCFL Komponenten
	extFactory, err := extensionimpl.NewFactory(nil, nil, logger)
	assert.NoError(t, err)

	ocflVer := version.Version1_1
	fact := factoryimpl.NewFactory(ocflVer, extFactory, logger)

	extManager0, err := extFactory.LoadExtensionManager(nil)
	assert.NoError(t, err)
	extManager, ok := extManager0.(storageroot.ExtensionManager)
	assert.True(t, ok, "extension manager should implement storageroot.ExtensionManager")

	sr := storagerootimpl.NewStorageRootBase(ctx, fact, ocflVer, extFactory, logger).
		WithWriteFS(srFS).
		WithDigestAlgorithm(checksum.DigestSHA512).
		WithExtensionManager(extManager)

	// 3. Storage Root Initialisierung
	initializer := sr.GetInitializer().
		WithFS(srFS)
	assert.NotNil(t, initializer)

	err = initializer.Init()
	assert.NoError(t, err)

	// 4. Objekt erstellen
	objID := "test-object"
	objFolder, err := sr.IdToFolder(objID)
	assert.NoError(t, err)

	objFS, err := appendfs.Sub(srFS, objFolder)
	assert.NoError(t, err)

	// Objekt-Extension Manager laden
	objExtManager0, err := extFactory.LoadExtensionManager(nil)
	assert.NoError(t, err)
	objExtManager, ok := objExtManager0.(object.ExtensionManager)
	assert.True(t, ok)

	obj, err := sr.CreateObject(objID, checksum.DigestSHA512, []checksum.DigestAlgorithm{}, objExtManager)
	assert.NoError(t, err)

	// Objekt initialisieren
	objInit := obj.GetInitializer(objFS)
	err = objInit.Init(objID, checksum.DigestSHA512, []checksum.DigestAlgorithm{})
	assert.NoError(t, err)

	// 5. Version hinzufügen
	vw, err := obj.StartUpdate(objFS, "initial version", "Junie", "junie@jetbrains.com", false)
	assert.NoError(t, err)

	testFileName := "test.txt"
	testContent := []byte("Hello OCFL")
	err = vw.AddData(testContent, testFileName, false, "", false, false)
	assert.NoError(t, err)

	err = vw.Close()
	assert.NoError(t, err)

	// 6. Verifizierung
	// Überprüfe ob Objekt-Namaste existiert
	objNamaste := "0=ocfl_object_" + string(ocflVer)
	_, err = vfs.Stat("vfs://testmem/" + objFolder + "/" + objNamaste)
	assert.NoError(t, err, "Object Namaste file should exist")

	// Überprüfe ob Inventory existiert
	_, err = vfs.Stat("vfs://testmem/" + objFolder + "/inventory.json")
	assert.NoError(t, err, "Inventory file should exist")

	// Überprüfe ob Inhaltsdatei existiert (standardmäßig in v1/content/)
	_, err = vfs.Stat("vfs://testmem/" + objFolder + "/v1/content/" + testFileName)
	assert.NoError(t, err, "Content file should exist")

	// 7. Objekt laden und Inventory prüfen
	// Erstelle ein neues leeres Objekt-Objekt
	loadedObj := fact.NewObject(ctx)

	// Loader erstellen und konfigurieren
	loader := fact.NewLoader(ctx).
		WithObject(loadedObj).
		WithFS(objFS).
		WithExtensionFactory(extFactory)

	// Objekt laden
	err = loader.Load()
	assert.NoError(t, err)

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
