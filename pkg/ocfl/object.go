package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"io"
	"io/fs"
)

type Object interface {
	LoadInventory(folder string) (Inventory, error)
	CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	StoreInventory() error
	StoreExtensions() error
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensions []Extension) error
	Load() error
	StartUpdate(msg string, UserName string, UserAddress string, echo bool) error
	AddFolder(fsys fs.FS, checkDuplicate bool) error
	AddFile(fsys fs.FS, sourceFilename string, internalFilename string, checkDuplicate bool) error
	DeleteFile(virtualFilename string, reader io.Reader, digest string) error
	GetID() string
	GetVersion() OCFLVersion
	Check() error
	Close() error
	getFS() OCFLFS
	IsModified() bool
}

func GetObjectVersion(ctx context.Context, ofs OCFLFS) (version OCFLVersion, err error) {
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
		addValidationErrors(ctx, GetValidationError(Version1_0, E003).AppendDescription("no version file found in '%s'", ofs.String()).AppendContext("object folder '%s'", ofs))
		return "", nil
	}
	return version, nil
}

func newObject(ctx context.Context, fsys OCFLFS, version OCFLVersion, storageroot StorageRoot, logger *logging.Logger) (Object, error) {
	var err error
	if version == "" {
		version, err = GetObjectVersion(ctx, fsys)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get version of object")
		}
	}
	switch version {
	case Version1_1:
		o, err := newObjectV1_1(ctx, fsys, storageroot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		o, err := newObjectV1_0(ctx, fsys, storageroot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
		//		return nil, errors.Finalize(fmt.Sprintf("Object Version %s not supported", version))
	}
}
