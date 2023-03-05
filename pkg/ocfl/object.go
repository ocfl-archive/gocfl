package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/op/go-logging"
	"io"
)

type Object interface {
	LoadInventory(folder string) (Inventory, error)
	CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	StoreInventory() error
	GetInventory() Inventory
	StoreExtensions() error
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensions []Extension) error
	Load() error
	StartUpdate(msg string, UserName string, UserAddress string, echo bool) error
	BeginArea(area string)
	EndArea() error
	AddFolder(fsys OCFLFSRead, checkDuplicate bool, area string) error
	AddFile(fsys OCFLFSRead, path string, checkDuplicate bool, area string) error
	AddReader(r io.ReadCloser, internalFilename string, area string) error
	DeleteFile(virtualFilename string, reader io.Reader, digest string) error
	GetID() string
	GetVersion() OCFLVersion
	Check() error
	Close() error
	GetFS() OCFLFSRead
	GetFSRW() OCFLFS
	IsModified() bool
	Stat(w io.Writer, statInfo []StatInfo) error
	Extract(fs OCFLFS, version string, manifest bool) error
	GetMetadata() (*ObjectMetadata, error)
	GetAreaPath(area string) (string, error)
}

func GetObjectVersion(ctx context.Context, ofs OCFLFSRead) (version OCFLVersion, err error) {
	files, err := ofs.ReadDir(".")
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
			cnt, err := ofs.ReadFile(file.Name())
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
		addValidationErrors(ctx, GetValidationError(Version1_0, E003).AppendDescription("no version file found in '%s'", ofs.String()).AppendContext("object folder '%s'", ofs))
		return "", nil
	}
	return version, nil
}

func newObject(ctx context.Context, fsys OCFLFSRead, version OCFLVersion, storageRoot StorageRoot, logger *logging.Logger) (Object, error) {
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
