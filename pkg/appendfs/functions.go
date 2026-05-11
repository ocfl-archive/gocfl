package appendfs

import (
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/writefs"
)

func Sub(fsys FS, path string) (FS, error) {
	if fsys == nil {
		return nil, errors.New("fsys is nil")
	}
	newFS, err := writefs.SubCreate(fsys, path)
	if err != nil {
		return nil, errors.Wrapf(err, "subfs %v/%s", fsys, path)
	}
	newAppendFS, ok := newFS.(FS)
	if !ok {
		return nil, errors.Errorf("subfs %v/%s is not a FS", fsys, path)
	}
	return newAppendFS, nil
}

func EnsureFS(fsys fs.FS) (FS, error) {
	if fsys == nil {
		return nil, errors.New("fsys is nil")
	}
	if sfs, ok := fsys.(FS); ok {
		return sfs, nil
	}
	return nil, errors.Errorf("filesystem %T does not implement appendfs.FS (must support Create and Mkdir)", fsys)
}
