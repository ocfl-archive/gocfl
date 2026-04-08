package objectimpl

import (
	"context"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/rs/zerolog"

	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"

	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
)

type inventoryMock struct {
	inventory.Inventory
}

func (mock inventoryMock) WithWriteable() inventory.Inventory {
	return mock
}

func (mock inventoryMock) WithID(s string) inventory.Inventory {
	return mock
}

func (mock inventoryMock) WithDigestAlgorithm(c checksum.DigestAlgorithm) inventory.Inventory {
	return mock
}

func (mock inventoryMock) WithFixity(i inventory.Fixity) inventory.Inventory {
	return mock
}

type fixityMock struct {
	inventory.Fixity
}

func (mock fixityMock) WithAllowedAlgorithms(...checksum.DigestAlgorithm) inventory.Fixity {
	return mock
}

func (mock fixityMock) WithAlgorithms(...checksum.DigestAlgorithm) inventory.Fixity {
	return mock
}

// factoryImpl cannot be imported due to a circular import.
type factoryMock struct {
	factory.Factory
}

func (mock factoryMock) GetVersion() version.OCFLVersion {
	return "value"
}

func (mock factoryMock) NewInventory(ctx context.Context) inventory.Inventory {
	return inventoryMock{}
}

func (mock factoryMock) NewFixity(ctx context.Context) inventory.Fixity {
	return fixityMock{}
}

type fsMock struct {
	fs.FS
}

type writeCloser struct {
	io.WriteCloser
}

func (mock fsMock) Create(s string) (writefs.FileWrite, error) {
	fileObj, _ := os.Create(s)
	wc := writeCloser{
		fileObj,
	}
	return wc, nil
}

func (mock fsMock) MkDir(s string) error {
	os.MkdirTemp("", s)
	return nil
}

type extManagerMock struct {
	object.ExtensionManager
}

func (mock extManagerMock) WriteConfig(appendfs.FS) error {
	return nil
}

func (mock extManagerMock) GetFixityDigests() []checksum.DigestAlgorithm {
	return []checksum.DigestAlgorithm{}
}

func getTestObject() object.Object {
	ctx := context.TODO()
	zerologger := zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger()
	ublogger := ublogger.Logger{
		Logger: &zerologger,
	}
	var logger = ocfllogger.NewOCFLLogger(ctx, &ublogger, nil, version.Default, nil)
	params := make(map[string]string)
	params["test"] = "test"
	ext, _ := extensionimpl.NewFactory(params, logger)
	obj := NewObjectBase(
		ctx,
		factoryMock{},
		"2.0",
		ext,
		logger,
	)
	extMgr := extManagerMock{}
	obj.WithExtensionManager(extMgr)
	return obj
}

func TestSomthing(t *testing.T) {
	ctx := context.TODO()
	zerologger := zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger()
	ublogger := ublogger.Logger{
		Logger: &zerologger,
	}
	var logger = ocfllogger.NewOCFLLogger(ctx, &ublogger, nil, version.Default, nil)
	dname, err := os.MkdirTemp("", "gocfl-test-initializer")
	fsys := os.DirFS(dname)
	obj := getTestObject()
	init := initializer{
		Object:   obj,
		logger:   logger,
		objectFS: fsMock{fsys},
		factory:  factoryMock{},
	}
	err = init.Init(
		"ID:123",
		checksum.DigestSHA512,
		[]checksum.DigestAlgorithm{},
	)
	if err != nil {
		t.Errorf("cannot initialize object: %s", err)
	}

	os.RemoveAll(dname)
}
