package objectimpl

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewInitializer(ctx context.Context, factory factory.Factory, logger ocfllogger.OCFLLogger) object.Initializer {
	return &initializer{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "initializer"),
	}
}

type initializer struct {
	object.Object
	objectFS appendfs.FS
	logger   ocfllogger.OCFLLogger
	ctx      context.Context
	factory  factory.Factory
}

func (initializer *initializer) WithObject(o object.Object) object.Initializer {
	initializer.Object = o
	return initializer
}

func (initializer *initializer) WithFS(objectFS appendfs.FS) object.Initializer {
	initializer.objectFS = objectFS
	return initializer
}

func (initializer *initializer) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error {
	initializer.logger.Debug().Msgf("%s", id)

	objectConformanceDeclaration := "ocfl_object_" + string(initializer.factory.GetVersion())
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration
	/*


		// first check whether object is not empty
		fp, err := initializer.initializer.destFS.Open(objectConformanceDeclarationFile)
		if err == nil {
			// not empty, close it and return error
			if err := fp.Close(); err != nil {
				return errors.Wrapf(err, "cannot close '%s'", objectConformanceDeclarationFile)
			}
			return fmt.Errorf("cannot create object '%s'. '%v/%s' already exists", id, initializer.initializer.destFS, objectConformanceDeclarationFile)
		}
		cnt, err := fs.ReadDir(initializer.initializer.destFS, ".")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errors.Wrapf(err, "cannot read '%v/%s'", initializer.initializer.destFS, ".")
		}
		if len(cnt) > 0 {
			return fmt.Errorf("'%v/%s' is not empty", ".", initializer.initializer.destFS)
		}
	*/
	if _, err := writefs.WriteFile(initializer.objectFS, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", initializer.objectFS, objectConformanceDeclarationFile)
	}

	if err := writefs.MkDir(initializer.objectFS, "extensions"); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", initializer.objectFS, "extensions")
	}
	subFS, err := appendfs.Sub(initializer.objectFS, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", initializer.objectFS, "extensions")
	}
	if err := initializer.GetExtensionManager().WriteConfig(subFS); err != nil {
		return errors.Wrapf(err, "cannot write extension config to %v", subFS)
	}

	// enforce sha512/sha256
	allowedAlgorithms := []checksum.DigestAlgorithm{
		checksum.DigestSHA512,
		checksum.DigestSHA256,
	}
	allowedAlgorithms = append(allowedAlgorithms, initializer.GetExtensionManager().GetFixityDigests()...)
	//slices.Sort(allowedAlgorithms)
	//allowedAlgorithms = slices.Compact(allowedAlgorithms)
	if !util.SliceContains(allowedAlgorithms, fixity) {
		return errors.Errorf("forbidden digest algorithm for fixity %v. Supported algorithms are %v. (to fix try to use extension 0001-digest-algorithms)", fixity, allowedAlgorithms)
	}

	newInventory := initializer.factory.NewInventory(initializer.ctx).
		WithWriteable().
		WithID(id).
		WithDigestAlgorithm(digest).
		WithFixity(initializer.factory.NewFixity(initializer.ctx).WithAllowedAlgorithms(allowedAlgorithms...).WithAlgorithms(fixity...))
	initializer.WithInventory(newInventory)
	return nil

}

var _ object.Initializer = (*initializer)(nil)
