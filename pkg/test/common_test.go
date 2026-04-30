package test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type TestEnv struct {
	DestFS           vfsrw.VFSRW
	SourceFS         appendfs.FS
	ReadSRFS         fs.FS
	ExtFactory       extension.Factory
	OCFLFactory      factory.Factory
	StorageRoot      storageroot.StorageRoot
	Logger           zLogger.ZLogger
	OCFLLogger       ocfllogger.OCFLLogger
	ExtensionManager storageroot.ExtensionManager
}

func SetupTestEnv(t *testing.T) *TestEnv {
	ctx := t.Context()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

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
	t.Cleanup(func() {
		_ = vfs.Close()
	})

	destFS := appendfs.FS(vfs)

	srFS, err := appendfs.Sub(destFS, "vfs://testmem/")
	assert.NoError(t, err)

	extFactory, err := extensionimpl.NewFactory(nil, nil, logger)
	assert.NoError(t, err)

	ocflVer := version.Version1_1
	fact := factoryimpl.NewFactory(ocflVer, extFactory, logger)

	sr := storagerootimpl.NewStorageRootBase(ctx, fact, ocflVer, extFactory, logger)
	sr.WithWriteFS(srFS)
	sr.WithDigestAlgorithm(checksum.DigestSHA512)

	extManager0, err := extFactory.LoadExtensionManager(nil)
	assert.NoError(t, err)
	extManager, ok := extManager0.(storageroot.ExtensionManager)
	assert.True(t, ok, "extension manager should implement storageroot.ExtensionManager")
	sr.WithExtensionManager(extManager)

	initializer := sr.GetInitializer()
	assert.NotNil(t, initializer)
	initializer.WithFS(srFS)

	err = initializer.Init()
	assert.NoError(t, err)

	// Lese-Dateisystem für Storage Root (als fs.FS) neu laden
	// Wir nutzen hier das ursprüngliche destFS, da es für In-Memory VFS okay ist.
	// In echten Szenarien (ZIP) müsste das Dateisystem neu geöffnet werden.
	readSRFS_append, err := appendfs.Sub(destFS, "vfs://testmem/")
	assert.NoError(t, err)
	readSRFS := fs.FS(readSRFS_append)

	// Den alten Storage Root verwerfen und einen neuen erstellen, der geladen wird.
	// Wichtig: Nach der Initialisierung (Schreiben) muss der Storage Root neu geladen werden,
	// da das Schreib-Dateisystem (appendfs) evtl. keine Lese-Operationen unterstützt (z.B. ZIP).
	sr = storagerootimpl.NewStorageRootBase(ctx, fact, ocflVer, extFactory, logger)
	sr.WithReadFS(readSRFS)
	sr.WithWriteFS(srFS) // Auch das Schreib-FS wieder mitgeben für spätere Updates in den Tests

	// Jetzt den Loader verwenden, um die Erweiterungen etc. aus dem Lese-FS zu laden
	loader := sr.GetLoader(extFactory)
	err = loader.Load()
	assert.NoError(t, err)

	return &TestEnv{
		DestFS:           vfs,
		SourceFS:         srFS,
		ReadSRFS:         readSRFS,
		ExtFactory:       extFactory,
		OCFLFactory:      fact,
		StorageRoot:      sr,
		Logger:           _zlogger,
		OCFLLogger:       logger,
		ExtensionManager: sr.GetExtensionManager(),
	}
}

func CreateTestObject(t *testing.T, env *TestEnv, objID string) (object.Object, appendfs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	assert.NoError(t, err)

	objFS, err := appendfs.Sub(env.SourceFS, objFolder)
	assert.NoError(t, err)

	objExtManager0, err := env.ExtFactory.LoadExtensionManager(nil)
	assert.NoError(t, err)
	objExtManager, ok := objExtManager0.(object.ExtensionManager)
	assert.True(t, ok)

	obj, err := env.StorageRoot.CreateObject(objID, checksum.DigestSHA512, []checksum.DigestAlgorithm{}, objExtManager)
	assert.NoError(t, err)

	objInit := obj.GetInitializer(objFS)
	err = objInit.Init(objID, checksum.DigestSHA512, []checksum.DigestAlgorithm{})
	assert.NoError(t, err)

	return obj, objFS
}

func ReloadObject(t *testing.T, env *TestEnv, objID string) (object.Object, fs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	assert.NoError(t, err)

	// Wir nutzen env.ReadSRFS als Lese-Dateisystem
	objFS, err := fs.Sub(env.ReadSRFS, objFolder)
	assert.NoError(t, err)

	loadedObj := env.OCFLFactory.NewObject(t.Context())
	loader := loadedObj.GetLoader(objFS, env.ExtFactory)
	err = loader.Load()
	assert.NoError(t, err)

	return loadedObj, objFS
}
