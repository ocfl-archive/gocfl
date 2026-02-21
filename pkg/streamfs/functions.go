package streamfs

import (
	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
)

func Sub(fsys FS, path string) (FS, error) {
	if fsys == nil {
		return nil, errors.New("fsys is nil")
	}
	newFS, err := writefs.Sub(fsys, path)
	if err != nil {
		return nil, errors.Wrapf(err, "subfs %v/%s", fsys, path)
	}
	newStreamFS, ok := newFS.(FS)
	if !ok {
		return nil, errors.Errorf("subfs %v/%s is not a FS", fsys, path)
	}
	return newStreamFS, nil
}
