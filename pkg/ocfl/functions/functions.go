package functions

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func CreateObject(ctx context.Context, id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, _ []checksum.DigestAlgorithm, _ extension.Factory[object.ExtensionManager], _ object.ExtensionManager, fsys appendfs.FS, logger ocfllogger.OCFLLogger) (object.Object, error) {
	obj, err := initocfl.InitObject(ctx, fsys, ver, id, digest, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && obj.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, obj.GetID())
	}

	return obj, nil
}

func LoadObjectByID(sr storageroot.StorageRoot, _ extension.Factory[object.ExtensionManager], id string, logger ocfllogger.OCFLLogger) (object.Object, error) {
	folder, err := sr.IdToFolder(id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	fsys, err := writefs.Sub(sr.GetReadFS(), folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs for %v / %s", sr.GetReadFS(), folder)
	}
	obj, err := initocfl.LoadObject(context.Background(), fsys, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	return obj, nil
}

func LoadObjectFS(ctx context.Context, objectFS fs.FS, _ extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) (object.Object, error) {
	return initocfl.LoadObject(ctx, objectFS, logger)
}

func CheckObject(ctx context.Context, objectFS fs.FS, _ extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) error {
	fmt.Printf("object folder '%v'\n", objectFS)
	obj, err := initocfl.LoadObject(ctx, objectFS, logger)
	if err != nil {
		logger.ValidationError(validation.E001, "invalid fsys '%v': %v", objectFS, err)
		return errors.Wrapf(err, "cannot load object from folder '%v'", objectFS)
	}
	checker := obj.GetChecker()
	if err := checker.Check(); err != nil {
		return errors.Wrapf(err, "cannot check object from folder '%v'", objectFS)
	}
	return nil
}

func Extract(ctx context.Context, objectFS fs.FS, destFS appendfs.FS, path string, version *inventory.VersionNumber, withManifest bool, area string, _ extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) error {
	if !version.IsValid() {
		version = inventory.NewVersionNumber().WithLatest()
	}

	logger.Debug().Msgf("Extracting object '%s' with version '%s'", path, version)
	var o object.Object
	var err error
	objFsys, err := writefs.Sub(objectFS, path)
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs  '%v' / %s", objectFS, path)
	}
	o, err = initocfl.LoadObject(ctx, objFsys, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	extractor := o.GetExtractor().WithDestFS(destFS)
	if err := extractor.Extract(version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

func ExtractMeta(ctx context.Context, fsys fs.FS, path string, _ extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) (*inventory.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	obj, err := initocfl.LoadObject(ctx, objFsys, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	defer logger.Debug().Msgf("extraction done")
	extractor := obj.GetExtractor().WithDestFS(nil)
	return extractor.GetMetadata()
}
