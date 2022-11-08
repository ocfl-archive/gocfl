package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"io/fs"
	"path/filepath"
)

type Object interface {
	LoadInventory() (Inventory, error)
	StoreInventory() error
	StoreExtensions() error
	New(id string) error
	Load() error
	StartUpdate(msg string, UserName string, UserAddress string) error
	AddFolder(fsys fs.FS) error
	AddFile(virtualFilename string, reader io.Reader, digest string) error
	DeleteFile(virtualFilename string, reader io.Reader, digest string) error
	GetID() string
	GetVersion() OCFLVersion
	Check() error
	Close() error
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
			r, err := ofs.Open(filepath.Join(file.Name()))
			if err != nil {
				return "", errors.Wrapf(err, "cannot open %s", file.Name())
			}
			cnt, err := io.ReadAll(r)
			if err != nil {
				r.Close()
				return "", errors.Wrapf(err, "cannot read %s", file.Name())
			}
			r.Close()
			t := fmt.Sprintf("ocfl_object_%s", version)
			if string(cnt) != t+"\n" && string(cnt) != t+"\r\n" {
				// todo: which error version should be used???
				addValidationErrors(ctx, GetValidationError(Version1_0, E007).AppendDescription("%s: %s != %s", file.Name(), cnt, t+"\\n"))
			}
		}
	}
	if version == "" {
		addValidationErrors(ctx, GetValidationError(Version1_0, E003).AppendDescription("no version file found in \"%s\"", ofs.String()))
		return "", nil
	}
	return version, nil
}

func NewObject(ctx context.Context, fsys OCFLFS, version OCFLVersion, id string, storageroot StorageRoot, logger *logging.Logger) (Object, error) {
	var err error
	if version == "" {
		version, err = GetObjectVersion(ctx, fsys)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get version of object")
		}
	}
	switch version {
	case Version1_1:
		o, err := NewObjectV1_1(ctx, fsys, id, storageroot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		o, err := NewObjectV1_0(ctx, fsys, id, storageroot, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
		//		return nil, errors.New(fmt.Sprintf("Object Version %s not supported", version))
	}
}
