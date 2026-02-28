package extension

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type Factory interface {
	AddCreator(name string, creator CreatorFunc)

	AddStorageRootDefaultExtension(ext Extension)
	AddObjectDefaultExtension(ext Extension)

	LoadExtensionFile(fsys fs.FS) (Extension, error)
	LoadExtensionData(fsys fs.FS, data []byte) (Extension, error)

	LoadExtensionManager(fsys fs.FS, ver version.OCFLVersion) (ManagerCore, error)
	LoadExtensions(fsys fs.FS, ver version.OCFLVersion) (Extension, error)
}
