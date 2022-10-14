package ocfl

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/object"
	"io"
	"io/fs"
	"path/filepath"
)

type Object interface {
	LoadInventory() (Inventory, error)
	StoreInventory() error
	StoreExtensions() error
	New(id string, path object.Path) error
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

func GetObjectVersion(ofs OCFLFS) (version OCFLVersion, err error) {
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
			t1 := fmt.Sprintf("ocfl_object_%s\n", version)
			t2 := fmt.Sprintf("ocfl_object_%s\r\n", version)
			if string(cnt) != t1 && string(cnt) != t2 {
				// todo: which error version should be used???
				return "", GetValidationError(Version1_0, E007)
			}
		}
	}
	if version == "" {
		return "", GetValidationError(Version1_0, E003)
	}
	return version, nil
}

func NewObject(fsys OCFLFS, version OCFLVersion, id string, logger *logging.Logger) (Object, error) {
	var err error
	if version == "" {
		version, err = GetObjectVersion(fsys)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get version of object")
		}
	}
	switch version {
	case Version1_0:
		o, err := NewObjectV1_0(fsys, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	case Version1_1:
		o, err := NewObjectV1_1(fsys, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		return nil, errors.New(fmt.Sprintf("Object Version %s not supported", version))
	}
}
