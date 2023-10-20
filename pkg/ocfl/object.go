package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"io"
	"io/fs"
)

type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

type Object interface {
	LoadInventory(folder string) (Inventory, error)
	CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	StoreInventory(version bool, objectRoot bool) error
	GetInventory() Inventory
	StoreExtensions() error
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensions []Extension) error
	Load() error
	StartUpdate(msg string, UserName string, UserAddress string, echo bool) error
	EndUpdate() error
	BeginArea(area string)
	EndArea() error
	AddFolder(fsys fs.FS, checkDuplicate bool, area string) error
	AddFile(fsys fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) error
	DeleteFile(virtualFilename string, digest string) error
	RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error
	GetID() string
	GetVersion() OCFLVersion
	Check() error
	Close() error
	GetFS() fs.FS
	IsModified() bool
	Stat(w io.Writer, statInfo []StatInfo) error
	Extract(fs fs.FS, version string, manifest bool) error
	GetMetadata() (*ObjectMetadata, error)
	GetAreaPath(area string) (string, error)
	GetExtensionManager() *ExtensionManager
	BuildNames(files []string, area string) (*NamesStruct, error)
}

func GetObjectVersion(ctx context.Context, ofs fs.FS) (version OCFLVersion, err error) {
	files, err := fs.ReadDir(ofs, ".")
	if err != nil {
		return "", errors.Wrap(err, "cannot get files")
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := objectVersionRegexp.FindStringSubmatch(file.Name())
		if matches != nil {
			if version != "" {
				return "", errVersionMultiple
			}
			version = OCFLVersion(matches[1])
			cnt, err := fs.ReadFile(ofs, file.Name())
			if err != nil {
				return "", errors.Wrapf(err, "cannot read %s", file.Name())
			}
			t := fmt.Sprintf("ocfl_object_%s", version)
			if string(cnt) != t+"\n" && string(cnt) != t+"\r\n" {
				// todo: which error version should be used???
				addValidationErrors(ctx, GetValidationError(Version1_0, E007).AppendDescription("%s: %s != %s", file.Name(), cnt, t+"\\n").AppendContext("object folder '%s'", ofs))
			}
		}
	}
	if version == "" {
		addValidationErrors(ctx, GetValidationError(Version1_0, E003).AppendDescription("no version file found in '%v'", ofs).AppendContext("object folder '%s'", ofs))
		return "", nil
	}
	return version, nil
}

func newObject(ctx context.Context, fsys fs.FS, version OCFLVersion, storageRoot StorageRoot, logger *logging.Logger) (Object, error) {
	var err error
	if version == "" {
		version, err = GetObjectVersion(ctx, fsys)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get version of object")
		}
	}
	switch version {
	case Version1_1:
		o, err := newObjectV1_1(ctx, fsys, storageRoot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		o, err := newObjectV1_0(ctx, fsys, storageRoot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
		//		return nil, errors.Finalize(fmt.Sprintf("Object Version %s not supported", version))
	}
}
