package test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/vfsrw"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/filesystem/pkg/zipfs"
	"github.com/ocfl-archive/filesystem/pkg/zipfsw"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectCreateZipContainer(t *testing.T) {
	ctx := t.Context()
	ocflVer := version.Version1_1
	objID := "test-zip-object"

	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfl.NewOCFLLogger(ctx, &zlogger, nil, ocflVer, nil)

	// 1. VFS konfigurieren
	cfg := vfsrw.Config{
		"testzip": &vfsrw.VFS{
			Name: "testzip",
			Type: "afero",
			Afero: &vfsrw.Afero{
				BaseDir: "mem://",
			},
		},
	}

	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	require.NoError(t, err)
	defer vfs.Close()

	// Wir erstellen die ZIP-Datei über zipfsw.NewFS
	// Wir nutzen Sub um an den testzip mount zu kommen
	baseSRFS, err := fs.Sub(vfs, "vfs:/testzip")
	require.NoError(t, err)

	zipFile, err := writefs.Create(baseSRFS, "container.zip")
	require.NoError(t, err)

	zFS, err := zipfsw.NewFS(zipFile, true, true, "container", []checksum.DigestAlgorithm{checksum.DigestSHA512}, nil, _zlogger)
	require.NoError(t, err)

	srFS := zFS.(appendfs.FS)

	// 2. StorageRoot im ZIP erzeugen
	var sr storageroot.StorageRoot
	sr, err = ocfl.InitStorageRoot(ctx, srFS, srFS, ocflVer, checksum.DigestSHA512, nil, logger)
	require.NoError(t, err)

	// 3. Objekt erzeugen
	objFolder, err := sr.IdToFolder(objID)
	require.NoError(t, err)

	objFS, err := writefs.Sub(srFS, objFolder)
	require.NoError(t, err)

	obj, err := ocfl.InitObject(ctx, objFS.(appendfs.FS), nil, ocflVer, objID, checksum.DigestSHA512, nil, logger)
	require.NoError(t, err)

	// 4. Daten hinzufügen
	vw, err := obj.StartUpdate("initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
	require.NoError(t, err)

	testFileName := "test_data.txt"
	testContent := []byte("This data should be inside a zip container")
	err = vw.AddData(testContent, testFileName, false, "", false, false)
	assert.NoError(t, err)

	err = vw.Close()
	assert.NoError(t, err)

	zFS.(io.Closer).Close()

	// 5. Verifizierung
	// Wir prüfen, ob im Basis-Filesystem (mem://) die container.zip existiert
	// Da wir nicht einfach an das zugrundeliegende FS kommen, ohne das Interface zu casten,
	// und wir wissen dass vfsrw.VFSRW ein vfsrw.FS (meistens) ist,
	// versuchen wir es direkt über Stat auf dem vfs mit dem vfs:// Prefix.
	_, err = fs.Stat(baseSRFS, "container.zip")
	assert.NoError(t, err, "container.zip should exist in the base filesystem")

	// 6. ZIP wieder öffnen und validieren
	// Wir brauchen Stat für die Größe
	fi, err := fs.Stat(baseSRFS, "container.zip")
	require.NoError(t, err)

	zipFileRead, err := baseSRFS.Open("container.zip")
	require.NoError(t, err)
	defer zipFileRead.Close()

	readerAt, ok := zipFileRead.(io.ReaderAt)
	if !ok {
		// Falls das FS kein ReaderAt liefert (unwahrscheinlich bei afero memfs für Dateien),
		// müssten wir es in den Speicher lesen.
		content, err := io.ReadAll(zipFileRead)
		require.NoError(t, err)
		readerAt = bytes.NewReader(content)
	}

	// Wir verwenden zipfs.NewFS zum Öffnen für den Lesezugriff
	// Signatur: NewFS(r io.ReaderAt, size int64, name string, logger zLogger.ZLogger)
	zFSRead, err := zipfs.NewFS(readerAt, fi.Size(), "container.zip", _zlogger)
	require.NoError(t, err)
	defer zFSRead.Close()

	srRead, err := ocfl.LoadStorageRoot(ctx, zFSRead, nil, nil, logger)
	require.NoError(t, err)

	objFolderRead, err := srRead.IdToFolder(objID)
	require.NoError(t, err)

	// Wir behelfen uns hier mit fs.Sub, da wir nur lesen wollen
	objFSRead, err := fs.Sub(zFSRead, objFolderRead)
	require.NoError(t, err)

	objRead, err := ocfl.LoadObject(ctx, objFSRead, nil, logger)
	require.NoError(t, err)
	defer objRead.Close()

	err = objRead.GetValidator().Validate()
	assert.NoError(t, err)
	validationErrors := logger.ValidationErrors()
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

	// Inhalt prüfen
	testFileNameRead := "test_data.txt"
	// Wir nutzen GetExtractor() um an die Dateien zu kommen,
	// da GetReadFS() das Basis-Verzeichnis des Objekts liefert (inkl. v1/, inventory.json etc.)
	// und nicht den logischen Stand einer Version.
	ext := objRead.GetExtractor()
	fp, _, _, err := ext.GetFileReader("v1/content/" + testFileNameRead)
	require.NoError(t, err)
	defer fp.Close()
	contentRead, err := io.ReadAll(fp)
	require.NoError(t, err)
	assert.Equal(t, testContent, contentRead)
}
