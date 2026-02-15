package functions

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func CreateObject(ctx context.Context, id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, fact factory.Factory, fsys fs.FS, logger zLogger.ZLogger) (object.Object, error) {
	obj := fact.NewObject(ctx).WithFS(fsys)
	if obj == nil {
		return nil, errors.New("cannot instantiate object")
	}

	// create initial filesystem structure for new object
	if err := obj.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && obj.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, obj.GetID())
	}

	return obj, nil
}

func LoadObject(ctx context.Context, fsys fs.FS, fact factory.Factory, logger zLogger.ZLogger) (object.Object, error) {
	obj := fact.NewObject(ctx).WithFS(fsys)
	// load the object
	if err := obj.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from fsys '%v'", fsys)
	}

	return obj, nil
}

func CheckObject(ctx context.Context, fsys fs.FS, fact factory.Factory, logger zLogger.ZLogger) error {
	fmt.Printf("object folder '%v'\n", fsys)
	validator, err := validation.NewValidator(ctx, version.Version1_0, fmt.Sprintf("%v", fsys), logger)
	if err != nil {
		return errors.Wrapf(err, "cannot create validator for '%v'", fsys)
	}
	obj, err := LoadObject(ctx, fsys, fact, logger)
	if err != nil {
		if err := validator.AddValidationError(validation.E001, "invalid fsys '%v': %v", fsys, err); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", validation.E001)
		}
		//			return errors.Wrapf(err, "cannot load object from folder '%s'", objectFolder)
	} else {
		if err := obj.Check(); err != nil {
			return errors.Wrapf(err, "check of '%s' failed", obj.GetID())
		}
	}
	return nil
}

func Extract(
	ctx context.Context,
	destFS,
	fsys fs.FS,
	path string,
	version *inventory.VersionNumber,
	withManifest bool,
	area string,
	fact factory.Factory,
	logger zLogger.ZLogger,
) error {
	if !version.IsValid() {
		version = inventory.NewVersionNumber().WithLatest()
	}

	logger.Debug().Msgf("Extracting object '%s' with version '%s'", path, version)
	var o object.Object
	var err error
	objFsys, err := writefs.Sub(fsys, path)
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs  '%v' / %s", fsys, path)
	}
	o, err = LoadObject(ctx, objFsys, fact, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	if err := o.Extract(destFS, version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

func ExtractMeta(
	ctx context.Context,
	fsys fs.FS,
	path string,
	fact factory.Factory,
	logger zLogger.ZLogger,
) (*inventory.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	o, err := LoadObject(ctx, objFsys, fact, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	logger.Debug().Msgf("extraction done")
	return o.GetMetadata()
}
