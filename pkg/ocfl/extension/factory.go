package extension

import (
	"encoding/json"
	"io/fs"
)

// BuilderFunc is a function type that creates a new instance of an Extension.
type BuilderFunc func() (Extension, error)

// ExternalParamFunc is a function type that returns a list of ExternalParam for an extension.
type ExternalParamFunc func() ([]*ExternalParam, error)

// Factory is an interface for managing and loading OCFL extensions.
// It allows registering extension builders and creators, and provides methods to load extensions from files or raw data.
type Factory[T ManagerCore[T]] interface {
	// AddCreator registers a CreatorFunc for a given extension name.
	AddCreator(name string, creator CreatorFunc, documentation *string)
	// RegisterExtension registers a BuilderFunc for a given extension name.
	RegisterExtension(name string, builder BuilderFunc, documentation *string)
	// AddStorageRootDefaultExtension adds a default extension to be used for storage roots.
	AddStorageRootDefaultExtension(ext Extension)
	// AddObjectDefaultExtension adds a default extension to be used for objects.
	AddObjectDefaultExtension(ext Extension)

	// LoadExtensionFile loads an extension from a file within the provided filesystem.
	LoadExtensionFile(fsys fs.FS) (Extension, error)
	// LoadExtensionData loads an extension from raw JSON data and an optional filesystem.
	LoadExtensionData(data json.RawMessage, extFS fs.FS) (Extension, error)

	// LoadExtensionManager loads an extension manager from the provided filesystem.
	LoadExtensionManager(fsys fs.FS) (T, error)
	// GetExtensionDocs returns a map of extension names to their documentation strings.
	GetExtensionDocs() map[string]*string
}
