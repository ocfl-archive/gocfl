package test

import (
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/vfsrw"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	DestFS      vfsrw.VFSRW
	SourceFS    appendfs.FS
	ReadSRFS    fs.FS
	StorageRoot storageroot.StorageRoot
	Logger      zLogger.ZLogger
	OCFLLogger  ocfllogger.OCFLLogger
	Closer      io.Closer
}

func SetupTestEnv(t *testing.T, ocflVer version.OCFLVersion) *TestEnv {
	ctx := t.Context()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfl.NewOCFLLogger(ctx, &zlogger, nil, ocflVer, nil)

	cfg := vfsrw.Config{
		"testmem": &vfsrw.VFS{
			Name: "testmem",
			Type: "afero",
			Afero: &vfsrw.Afero{
				BaseDir: "mem://",
			},
			/*
				ZipAsFolder: &vfsrw.ZipAsFolder{
					Enabled:   true,
					Digests:   []checksum.DigestAlgorithm{checksum.DigestSHA512},
					CacheSize: 10,
					Compress:  false,
					ReadOnly:  false,
				},
			*/
		},
	}

	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = vfs.Close()
	})

	destFS := appendfs.FS(vfs)

	srFS, closer0, err := appendfs.Sub(destFS, "vfs://testmem/")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = closer0.Close()
	})

	_, err = ocfl.InitStorageRoot(ctx, srFS, srFS, ocflVer, checksum.DigestSHA512, nil, logger)
	require.NoError(t, err)

	// Lese-Dateisystem für Storage Root (als fs.FS) neu laden
	// Wir nutzen hier das ursprüngliche destFS, da es für In-Memory VFS okay ist.
	readSRFS := fs.FS(srFS)

	// Den Storage Root neu laden
	sr, err := ocfl.LoadStorageRoot(ctx, readSRFS, nil, nil, logger)
	require.NoError(t, err)
	// Auch das Schreib-FS wieder mitgeben für spätere Updates in den Tests
	sr = sr.WithWriteFS(srFS)

	return &TestEnv{
		DestFS:      vfs,
		SourceFS:    srFS,
		ReadSRFS:    readSRFS,
		StorageRoot: sr,
		Logger:      _zlogger,
		OCFLLogger:  logger,
		Closer:      sr,
	}
}

func CreateTestObject(t *testing.T, env *TestEnv, objID string, ocflVer version.OCFLVersion) (object.Object, appendfs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	objFS, closer0, err := appendfs.Sub(env.SourceFS, objFolder)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = closer0.Close()
	})

	obj, err := ocfl.InitObject(t.Context(), objFS, nil, ocflVer, objID, checksum.DigestSHA512, nil, env.OCFLLogger)
	require.NoError(t, err)

	return obj, objFS
}

func ReloadObject(t *testing.T, env *TestEnv, objID string) (object.Object, fs.FS, io.Closer) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	// Wir nutzen env.ReadSRFS als Lese-Dateisystem
	objFS, err := fs.Sub(env.ReadSRFS, objFolder)
	require.NoError(t, err)

	loadedObj, err := ocfl.LoadObject(t.Context(), objFS, nil, env.OCFLLogger)
	require.NoError(t, err)

	return loadedObj, objFS, loadedObj
}
