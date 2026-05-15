package extension

import (
	"encoding/json"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
)

// DefaultExtensionManagerName is the name of the default gocfl extension manager.
const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"

// DefaultExtensionInitialName is the name of the default initial extension.
const DefaultExtensionInitialName = "initial"

// ExtensionManagerTypeAssertionError is returned when a manager cannot be converted to the expected type.
var ExtensionManagerTypeAssertionError = errors.New("cannot convert manager to type")

// CreatorFunc is a function type that creates an Extension from raw data.
type CreatorFunc func(data json.RawMessage, extFS fs.FS) (Extension, error)

// Initial is an interface for extensions that handle initial object state.
type Initial interface {
	Extension
	// GetExtension returns the name of the initial extension.
	GetExtension() string
	// SetExtension sets the name of the initial extension.
	SetExtension(ext string)
}

// ManagerCore is the base interface for an extension manager.
// It manages a collection of extensions and provides methods for their lifecycle and configuration.
type ManagerCore[T any] interface {
	Extension
	// GetConfig returns the configuration of the manager.
	GetConfig() any
	// GetExtensions returns all extensions managed by this manager.
	GetExtensions() []Extension
	// Add adds an extension to the manager.
	Add(ext Extension) error
	// Finalize signals that no more extensions will be added and performs final setup.
	Finalize()
	// GetConfigName returns the configuration for a specific extension by name.
	GetConfigName(extName string) (any, error)
	// StoreRootLayout writes the storage root layout configuration.
	StoreRootLayout(fsys appendfs.FS) error
	// SetInitial sets the initial extension for the manager.
	SetInitial(initial Initial)
}

// ManagerConfig represents the configuration for an extension manager.
type ManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`      // Sorting rules for extensions.
	Exclusion map[string][][]string `json:"exclusion"` // Exclusion rules for extensions.
}
