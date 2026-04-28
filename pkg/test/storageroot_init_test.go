package test

import (
	"context"
	"os"
	"testing"

	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestStorageRootInit(t *testing.T) {
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(context.Background(), &zlogger, nil, version.Version1_1, nil)

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

	sr := storagerootimpl.NewStorageRootBase(context.Background(), fact, ocflVer, extFactory, logger)
	sr.WithWriteFS(srFS)
	sr.WithDigestAlgorithm(checksum.DigestSHA512)

	extManager0, err := extFactory.LoadExtensionManager(nil)
	assert.NoError(t, err)
	extManager, ok := extManager0.(storageroot.ExtensionManager)
	assert.True(t, ok, "extension manager should implement storageroot.ExtensionManager")
	sr.WithExtensionManager(extManager)

	// 3. Storage Root Initialisierung
	initializer := sr.GetInitializer()
	assert.NotNil(t, initializer)
	initializer.WithFS(srFS)

	err = initializer.Init()
	assert.NoError(t, err)

	// 4. Verifizierung
	// Überprüfe ob Namaste-Datei existiert
	namaste := "0=ocfl_" + string(ocflVer)
	_, err = vfs.Stat("vfs://testmem/" + namaste)
	assert.NoError(t, err, "Namaste file should exist")

	// Überprüfe ob OCFL Spezifikation kopiert wurde (aus gocfl/docs)
	_, err = vfs.Stat("vfs://testmem/ocfl_spec_1.1.md")
	assert.NoError(t, err, "OCFL specification doc should exist")

	// Überprüfe ob Extensions-Konfigurationen vorhanden sind
	_, err = vfs.Stat("vfs://testmem/extensions/initial/config.json")
	assert.NoError(t, err, "initial extension config should exist")

	_, err = vfs.Stat("vfs://testmem/extensions/NNNN-gocfl-extension-manager/config.json")
	assert.NoError(t, err, "extension manager config should exist")
}
