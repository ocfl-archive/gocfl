package functions

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func GetObjectVersion(ctx context.Context, ofs fs.FS) (ver version.OCFLVersion, err error) {
	files, err := fs.ReadDir(ofs, ".")
	if err != nil {
		return "", errors.Wrap(err, "cannot get files")
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := object.ObjectVersionRegexp.FindStringSubmatch(file.Name())
		if matches != nil {
			if ver != "" {
				return "", ocflerrors.ErrVersionMultiple
			}
			ver = version.OCFLVersion(matches[1])
			cnt, err := fs.ReadFile(ofs, file.Name())
			if err != nil {
				return "", errors.Wrapf(err, "cannot read %s", file.Name())
			}
			t := fmt.Sprintf("ocfl_object_%s", ver)
			if string(cnt) != t+"\n" && string(cnt) != t+"\r\n" {
				// todo: which error version should be used???
				validation.AddValidationErrors(ctx, validation.GetValidationError(version.Version1_0, validation.E007).AppendDescription("%s: %s != %s", file.Name(), cnt, t+"\\n").AppendContext("object folder '%s'", ofs))
			}
		}
	}
	if ver == "" {
		validation.AddValidationErrors(ctx, validation.GetValidationError(version.Version1_0, validation.E003).AppendDescription("no version file found in '%v'", ofs).AppendContext("object folder '%s'", ofs))
		return "", nil
	}
	return ver, nil
}

func NewObject(ctx context.Context, factory types.Factory, fsys fs.FS, ver version.OCFLVersion, extensionFactory *extension.ExtensionFactory, extensionManager types.ExtensionManagerCore, logger zLogger.ZLogger) (types.Object, error) {
	var err error
	if ver == "" {
		ver, err = GetObjectVersion(ctx, fsys)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get version of object")
		}
	}
	o := factory.NewObject(ctx)
	return o, nil
}

func CreateObject(ctx context.Context, factory types.Factory, id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensionFactory *extension.ExtensionFactory, manager types.ExtensionManagerCore, fsys fs.FS, logger zLogger.ZLogger) (types.Object, error) {
	object, err := NewObject(ctx, factory, fsys, ver, extensionFactory, manager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate object")
	}

	// create initial filesystem structure for new object
	if err = object.Init(id, digest, fixity, manager); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && object.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, object.GetID())
	}

	return object, nil
}

func LoadObject(ctx context.Context, fsys fs.FS, extensionFactory *extension.ExtensionFactory, logger zLogger.ZLogger) (types.Object, error) {
	ver, err := util.GetVersion(ctx, fsys, "", "ocfl_object_")
	if errors.Is(err, ocflerrors.ErrVersionNone) {
		if err := validation.AddValidationError(ctx, version.Version1_0, validation.E003, "no version in fsys '%v'", fsys); err != nil {
			return nil, errors.Wrapf(err, "cannot add validation error %s", validation.E003)
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in '%v'", fsys)
	}
	extFSys, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of '%v' for '%s'", fsys, "extensions")
	}
	validator, err := validation.NewValidator(ctx, ver, fmt.Sprintf("fsys: %v", fsys), logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create validator for '%v'", fsys)
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, validator)
	//	extensionManager.SetFS(extFSys)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	f := factory.NewFactory(ver, extensionFactory, extensionManager, logger)
	object, err := NewObject(ctx, f, fsys, ver, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot instantiate object")
	}
	// load the object
	if err := object.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from fsys '%v'", fsys)
	}

	return object, nil
}

func CheckObject(ctx context.Context, factory types.Factory, fsys fs.FS, extensionFactory *extension.ExtensionFactory, logger zLogger.ZLogger) error {
	fmt.Printf("object folder '%v'\n", fsys)
	validator, err := validation.NewValidator(ctx, version.Version1_0, fmt.Sprintf("%v", fsys), logger)
	if err != nil {
		return errors.Wrapf(err, "cannot create validator for '%v'", fsys)
	}
	object, err := LoadObject(ctx, fsys, extensionFactory, logger)
	if err != nil {
		if err := validator.AddValidationError(validation.E001, "invalid fsys '%v': %v", fsys, err); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", validation.E001)
		}
		//			return errors.Wrapf(err, "cannot load object from folder '%s'", objectFolder)
	} else {
		if err := object.Check(); err != nil {
			return errors.Wrapf(err, "check of '%s' failed", object.GetID())
		}
	}
	return nil
}

func Extract(ctx context.Context, factory types.Factory, destFS, fsys fs.FS, path string, version *types.VersionNumber, withManifest bool, area string, extensionFactory *extension.ExtensionFactory, logger zLogger.ZLogger) error {
	if !version.IsValid() {
		version = types.NewVersionNumber().WithLatest()
	}

	logger.Debug().Msgf("Extracting object '%s' with version '%s'", path, version)
	var o types.Object
	var err error
	objFsys, err := writefs.Sub(fsys, path)
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs  '%v' / %s", fsys, path)
	}
	o, err = LoadObject(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	if err := o.Extract(destFS, version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

func ExtractMeta(ctx context.Context, factory types.Factory, fsys fs.FS, path string, extensionFactory *extension.ExtensionFactory, logger zLogger.ZLogger) (*types.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	o, err := LoadObject(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	logger.Debug().Msgf("extraction done")
	return o.GetMetadata()
}
