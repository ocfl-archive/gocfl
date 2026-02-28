package streamfs

import (
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
)

func Sub(fsys FS, path string) (FS, error) {
	if fsys == nil {
		return nil, errors.New("fsys is nil")
	}
	newFS, err := writefs.SubFSCreate(fsys, path)
	if err != nil {
		return nil, errors.Wrapf(err, "subfs %v/%s", fsys, path)
	}
	newStreamFS, ok := newFS.(FS)
	if !ok {
		return nil, errors.Errorf("subfs %v/%s is not a FS", fsys, path)
	}
	return newStreamFS, nil
}

func EnsureFS(fsys fs.FS) (FS, error) {
	if fsys == nil {
		return nil, errors.New("fsys is nil")
	}
	if sfs, ok := fsys.(FS); ok {
		return sfs, nil
	}
	return nil, errors.Errorf("filesystem %T does not implement streamfs.FS (must support Create and Mkdir)", fsys)
}
