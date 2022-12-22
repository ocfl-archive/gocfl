package baseFS

import (
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io/fs"
)

type ocflFS2FS struct {
	ocfl.OCFLFS
}

func (ofs2fs *ocflFS2FS) Open(name string) (fs.File, error) {
	return ofs2fs.OCFLFS.Open(name)
}
