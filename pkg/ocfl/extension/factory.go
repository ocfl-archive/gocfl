package extension

import (
	"encoding/json"
	"io/fs"
)

type BuilderFunc func() (Extension, error)
type ExternalParamFunc func() ([]*ExternalParam, error)

type Factory[T ManagerCore[T]] interface {
	AddCreator(name string, creator CreatorFunc, documentation *string)
	RegisterExtension(name string, builder BuilderFunc, documentation *string)
	AddStorageRootDefaultExtension(ext Extension)
	AddObjectDefaultExtension(ext Extension)

	LoadExtensionFile(fsys fs.FS) (Extension, error)
	LoadExtensionData(data json.RawMessage, extFS fs.FS) (Extension, error)

	LoadExtensionManager(fsys fs.FS) (T, error)
	GetExtensionDocs() map[string]*string
}
