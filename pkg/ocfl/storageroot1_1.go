package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"io/fs"
)

const Version1_1 OCFLVersion = "1.1"

type StorageRootV1_1 struct {
	*StorageRootBase
}

func NewStorageRootV1_1(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootV1_1, error) {
	srb, err := NewStorageRootBase(ctx, fsys, Version1_1, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_1)
	}

	sr := &StorageRootV1_1{StorageRootBase: srb}
	return sr, nil
}

func (osr *StorageRootV1_1) Init(version OCFLVersion, digest checksum.DigestAlgorithm, extensions []Extension) error {
	/*
		specFile := "ocfl_1.1.md"
		spec, err := writefs.Create(osr.fsys, specFile)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", specFile)
		}
		if _, err := spec.Write(specs.OCFL1_1); err != nil {
			_ = spec.Close()
			return errors.Wrapf(err, "cannot write into '%s'", specFile)
		}
		if err := spec.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%s'", specFile)
		}

	*/
	return osr.StorageRootBase.Init(version, digest, extensions)
}
