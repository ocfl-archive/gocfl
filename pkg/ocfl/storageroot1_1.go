package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"strconv"
)

const Version1_1 OCFLVersion = "1.1"

type StorageRootV1_1 struct {
	*StorageRootBase
}

func NewStorageRootV1_1(ctx context.Context, fs OCFLFS, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootV1_1, error) {
	srb, err := NewStorageRootBase(ctx, fs, Version1_1, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_1)
	}

	sr := &StorageRootV1_1{StorageRootBase: srb}
	return sr, nil
}

func (osr *StorageRootV1_1) OpenObject(id string) (Object, error) {
	folder, err := osr.extensionManager.BuildStoragerootPath(osr, id)
	version, err := getVersion(osr.ctx, osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		subfs, err := osr.fs.SubFS(folder)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create subfs of %v for %s", osr.fs, folder)
		}
		return NewObject(osr.ctx, subfs, osr.version, id, osr, osr.logger)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s for [%s]", folder, id)
	}
	versionFloat, err := strconv.ParseFloat(string(version), 64)
	if err != nil {
		return nil, errors.WithStack(GetValidationError(Version1_1, E004))
	}
	rootVersionFloat, err := strconv.ParseFloat(string(osr.version), 64)
	if err != nil {
		return nil, errors.WithStack(GetValidationError(Version1_1, E075))
	}
	// TODO: check. could not find this rule in standard
	if versionFloat > rootVersionFloat {
		return nil, errors.Errorf("root OCFL version declaration (%s) smaller than highest object version declaration (%s)", osr.version, version)
	}
	subfs, err := osr.fs.SubFS(folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of %v for %s", osr.fs, folder)
	}

	return NewObject(osr.ctx, subfs, version, id, osr, osr.logger)
}
