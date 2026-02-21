package objectimpl

import (
	"slices"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

func (objectBase *ObjectBase) Init(fsys streamfs.FS, id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensionManager extension.ExtensionManagerCore) error {
	objectBase.logger.Debug().Msgf("%s", id)
	objectBase.extensionManager = extensionManager.(object.ExtensionManager)

	objectConformanceDeclaration := "ocfl_object_" + string(objectBase.version)
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration
	/*


		// first check whether object is not empty
		fp, err := objectBase.fsys.Open(objectConformanceDeclarationFile)
		if err == nil {
			// not empty, close it and return error
			if err := fp.Close(); err != nil {
				return errors.Wrapf(err, "cannot close '%s'", objectConformanceDeclarationFile)
			}
			return fmt.Errorf("cannot create object '%s'. '%v/%s' already exists", id, objectBase.fsys, objectConformanceDeclarationFile)
		}
		cnt, err := fs.ReadDir(objectBase.fsys, ".")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return errors.Wrapf(err, "cannot read '%v/%s'", objectBase.fsys, ".")
		}
		if len(cnt) > 0 {
			return fmt.Errorf("'%v/%s' is not empty", ".", objectBase.fsys)
		}
	*/
	if _, err := writefs.WriteFile(fsys, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", fsys, objectConformanceDeclarationFile)
	}

	if err := writefs.MkDir(fsys, "extensions"); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", fsys, "extensions")
	}
	subFS, err := streamfs.Sub(fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", fsys, "extensions")
	}
	if err := objectBase.extensionManager.WriteConfig(subFS); err != nil {
		return errors.Wrapf(err, "cannot write extension config to %v", subFS)
	}

	// enforce sha512/sha256
	algs := []checksum.DigestAlgorithm{
		checksum.DigestSHA512,
		checksum.DigestSHA256,
	}
	algs = append(algs, objectBase.extensionManager.GetFixityDigests()...)
	slices.Sort(algs)
	algs = slices.Compact(algs)
	if !util.SliceContains(algs, fixity) {
		return errors.Errorf("forbidden digest algorithm for fixity %v. Supported algorithms are %v. (to fix try to use extension 0001-digest-algorithms)", fixity, algs)
	}

	objectBase.i = objectBase.factory.NewInventory(objectBase.ctx).
		WithWriteable().
		WithID(id).
		WithDigestAlgorithm(digest).
		WithFixity(objectBase.factory.NewFixity(objectBase.ctx).WithAlgorithms(fixity...))
	return nil
}
