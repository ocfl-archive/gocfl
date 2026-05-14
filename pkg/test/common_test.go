package test

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/je4/filesystem/v4/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	DestFS                vfsrw.VFSRW
	SourceFS              appendfs.FS
	ReadSRFS              fs.FS
	ObjectExtFactory      extension.Factory[object.ExtensionManager]
	StorageRootExtFactory extension.Factory[storageroot.ExtensionManager]
	StorageRootFactory    factory.FactoryStorageRoot
	ObjectFactory         factory.FactoryObject
	StorageRoot           storageroot.StorageRoot
	Logger                zLogger.ZLogger
	OCFLLogger            ocfllogger.OCFLLogger
	ExtensionManager      storageroot.ExtensionManager
	ObjectExtManager      object.ExtensionManager
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
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = vfs.Close()
	})

	destFS := appendfs.FS(vfs)

	srFS, err := appendfs.Sub(destFS, "vfs://testmem/")
	require.NoError(t, err)

	storageRootExtFactory, err := extensionimpl.NewFactory[storageroot.ExtensionManager](nil, logger)
	require.NoError(t, err)
	storageRootExtManager, err := storageRootExtFactory.LoadExtensionManager(nil)
	require.NoError(t, err)

	objectExtFactory, err := extensionimpl.NewFactory[object.ExtensionManager](nil, logger)
	require.NoError(t, err)
	objectExtManager, err := objectExtFactory.LoadExtensionManager(nil)
	require.NoError(t, err)

	ocflVer := version.Version1_1
	storageRootFact := initocfl.NewFactoryStorageRoot(ocflVer, storageRootExtFactory, logger)
	objectFact := initocfl.NewFactoryObject(ocflVer, objectExtFactory, logger)

	sr := storagerootimpl.NewStorageRootBase(ctx, storageRootFact, ocflVer, storageRootExtFactory, logger)
	sr.WithWriteFS(srFS)
	sr.WithDigestAlgorithm(checksum.DigestSHA512)

	sr.WithExtensionManager(storageRootExtManager)

	initializer := sr.GetInitializer()
	require.NotNil(t, initializer)

	err = initializer.Init()
	require.NoError(t, err)

	// Lese-Dateisystem für Storage Root (als fs.FS) neu laden
	// Wir nutzen hier das ursprüngliche destFS, da es für In-Memory VFS okay ist.
	// In echten Szenarien (ZIP) müsste das Dateisystem neu geöffnet werden.
	readSRFS_append, err := appendfs.Sub(destFS, "vfs://testmem/")
	require.NoError(t, err)
	readSRFS := fs.FS(readSRFS_append)

	// Den alten Storage Root verwerfen und einen neuen erstellen, der geladen wird.
	// Wichtig: Nach der Initialisierung (Schreiben) muss der Storage Root neu geladen werden,
	// da das Schreib-Dateisystem (appendfs) evtl. keine Lese-Operationen unterstützt (z.B. ZIP).
	sr = storagerootimpl.NewStorageRootBase(ctx, storageRootFact, ocflVer, storageRootExtFactory, logger)
	sr.WithReadFS(readSRFS)
	sr.WithWriteFS(srFS) // Auch das Schreib-FS wieder mitgeben für spätere Updates in den Tests

	// Jetzt den Loader verwenden, um die Erweiterungen etc. aus dem Lese-FS zu laden
	loader := sr.GetLoader()
	err = loader.Load()
	require.NoError(t, err)

	return &TestEnv{
		DestFS:                vfs,
		SourceFS:              srFS,
		ReadSRFS:              readSRFS,
		StorageRootFactory:    storageRootFact,
		StorageRootExtFactory: storageRootExtFactory,
		ObjectFactory:         objectFact,
		ObjectExtFactory:      objectExtFactory,
		ObjectExtManager:      objectExtManager,
		StorageRoot:           sr,
		Logger:                _zlogger,
		OCFLLogger:            logger,
		ExtensionManager:      storageRootExtManager,
	}
}

func CreateTestObject(t *testing.T, env *TestEnv, objID string) (object.Object, appendfs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	objFS, err := appendfs.Sub(env.SourceFS, objFolder)
	require.NoError(t, err)

	obj := env.ObjectFactory.NewObject(context.TODO()).WithExtensionManager(env.ObjectExtManager).WithWriteFS(objFS)
	objInit := obj.GetInitializer()

	err = objInit.Init(objID, checksum.DigestSHA512, []checksum.DigestAlgorithm{})
	require.NoError(t, err)

	return obj, objFS
}

func ReloadObject(t *testing.T, env *TestEnv, objID string) (object.Object, fs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	// Wir nutzen env.ReadSRFS als Lese-Dateisystem
	objFS, err := fs.Sub(env.ReadSRFS, objFolder)
	require.NoError(t, err)

	loadedObj := env.ObjectFactory.NewObject(t.Context()).WithReadFS(objFS)
	loader := loadedObj.GetLoader().WithFS(objFS)
	err = loader.Load()
	require.NoError(t, err)

	return loadedObj, objFS
}
