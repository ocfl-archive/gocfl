package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io/fs"
)

const Version2_0 OCFLVersion = "2.0"

type StorageRootV2_0 struct {
	*StorageRootBase
}

func NewStorageRootV2_0(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, logger zLogger.ZLogger) (*StorageRootV2_0, error) {
	srb, err := NewStorageRootBase(ctx, fsys, Version2_0, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version2_0)
	}

	sr := &StorageRootV2_0{StorageRootBase: srb}
	return sr, nil
}

func (osr *StorageRootV2_0) Init(version OCFLVersion, digest checksum.DigestAlgorithm, manager ExtensionManager) error {
	/*
		specFile := "ocfl_1.1.md"
		spec, err := writefs.Create(osr.fsys, specFile)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", specFile)
		}
		if _, err := spec.Write(specs.OCFL2_0); err != nil {
			_ = spec.Close()
			return errors.Wrapf(err, "cannot write into '%s'", specFile)
		}
		if err := spec.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%s'", specFile)
		}

	*/
	return osr.StorageRootBase.Init(version, digest, manager)
}
