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
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
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

	ocflVer := version.Version1_1
	_, err = initocfl.InitStorageRoot(ctx, srFS, srFS, ocflVer, checksum.DigestSHA512, nil, logger)
	require.NoError(t, err)

	// Lese-Dateisystem für Storage Root (als fs.FS) neu laden
	// Wir nutzen hier das ursprüngliche destFS, da es für In-Memory VFS okay ist.
	readsrfsAppend, err := appendfs.Sub(destFS, "vfs://testmem/")
	require.NoError(t, err)
	readSRFS := fs.FS(readsrfsAppend)

	// Den Storage Root neu laden
	sr, srCloser, err := initocfl.LoadStorageRoot(ctx, readSRFS, nil, nil, logger)
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
		Closer:      srCloser,
	}
}

func CreateTestObject(t *testing.T, env *TestEnv, objID string) (object.Object, appendfs.FS) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	objFS, err := appendfs.Sub(env.SourceFS, objFolder)
	require.NoError(t, err)

	obj, err := initocfl.InitObject(t.Context(), objFS, nil, env.StorageRoot.GetOCFLVersion(), objID, checksum.DigestSHA512, nil, env.OCFLLogger)
	require.NoError(t, err)

	return obj, objFS
}

func ReloadObject(t *testing.T, env *TestEnv, objID string) (object.Object, fs.FS, io.Closer) {
	objFolder, err := env.StorageRoot.IdToFolder(objID)
	require.NoError(t, err)

	// Wir nutzen env.ReadSRFS als Lese-Dateisystem
	objFS, err := fs.Sub(env.ReadSRFS, objFolder)
	require.NoError(t, err)

	loadedObj, objCloser, err := initocfl.LoadObject(t.Context(), objFS, nil, env.OCFLLogger)
	require.NoError(t, err)

	return loadedObj, objFS, objCloser
}
