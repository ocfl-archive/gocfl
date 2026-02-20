package streamfs

import (
	"io/fs"

	"github.com/je4/filesystem/v3/pkg/writefs"
)

type FS interface {
	fs.FS
	writefs.MkDirFS
	writefs.CreateFS
}
