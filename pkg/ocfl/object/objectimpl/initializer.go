package objectimpl

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
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
	objectFS streamfs.FS
	logger   ocfllogger.OCFLLogger
	ctx      context.Context
	factory  factory.Factory
}

func (initializer *initializer) WithObject(o object.Object) object.Initializer {
	initializer.Object = o
	return initializer
}

func (initializer *initializer) WithFS(objectFS streamfs.FS) object.Initializer {
	initializer.objectFS = objectFS
	return initializer
}

func (initializer *initializer) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error {
	initializer.logger.Debug().Msgf("%s", id)

	objectConformanceDeclaration := "ocfl_object_" + string(initializer.factory.GetVersion())
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration
	/*


		// first check whether object is not empty
		fp, err := initializer.initializer.objectFS.Open(objectConformanceDeclarationFile)
		if err == nil {
			// not empty, close it and return error
			if err := fp.Close(); err != nil {
				return errors.Wrapf(err, "cannot close '%s'", objectConformanceDeclarationFile)
			}
			return fmt.Errorf("cannot create object '%s'. '%v/%s' already exists", id, initializer.initializer.objectFS, objectConformanceDeclarationFile)
		}
		cnt, err := fs.ReadDir(initializer.initializer.objectFS, ".")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errors.Wrapf(err, "cannot read '%v/%s'", initializer.initializer.objectFS, ".")
		}
		if len(cnt) > 0 {
			return fmt.Errorf("'%v/%s' is not empty", ".", initializer.initializer.objectFS)
		}
	*/
	if _, err := writefs.WriteFile(initializer.objectFS, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", initializer.objectFS, objectConformanceDeclarationFile)
	}

	if err := writefs.MkDir(initializer.objectFS, "extensions"); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", initializer.objectFS, "extensions")
	}
	subFS, err := streamfs.Sub(initializer.objectFS, "extensions")
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
		WithFixity(initializer.factory.NewFixity(initializer.ctx).WithAlgorithms(fixity...))
	initializer.WithInventory(newInventory)
	return nil

}

var _ object.Initializer = (*initializer)(nil)
