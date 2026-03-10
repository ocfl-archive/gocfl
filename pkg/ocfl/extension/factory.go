package extension

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type BuilderFunc func() (Extension, error)
type ExternalParamFunc func() ([]*ExternalParam, error)

type Factory interface {
	AddCreator(name string, creator CreatorFunc)
	RegisterExtension(name string, builder BuilderFunc)
	AddStorageRootDefaultExtension(ext Extension)
	AddObjectDefaultExtension(ext Extension)

	LoadExtensionFile(fsys fs.FS) (Extension, error)
	LoadExtensionData(fsys fs.FS, data []byte) (Extension, error)

	LoadExtensionManager(fsys fs.FS, ver version.OCFLVersion) (ManagerCore, error)
}
