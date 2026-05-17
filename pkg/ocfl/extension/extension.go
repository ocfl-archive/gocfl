// Package extension provides the basic interfaces and structures for OCFL extensions.
// It defines how extensions are loaded, configured, and managed within the gocfl ecosystem.
package extension

import (
	"encoding/json"
	"io/fs"

	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

// ExtensionConfig represents the base configuration for any OCFL extension,
// containing at least the extension name.
type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

// Extension is the core interface that all OCFL extensions must implement.
// It provides methods for lifecycle management, configuration, and logging.
type Extension interface {
	// WithLogger returns a new instance of the extension with the provided logger.
	WithLogger(logger ocfllogger.OCFLLogger) Extension
	// GetName returns the unique name of the extension.
	GetName() string
	// Load initializes the extension with the provided configuration data and filesystem access.
	Load(data json.RawMessage, extFS fs.FS) error
	// SetParams allows setting external parameters for the extension.
	SetParams(params map[string]string) error
	// WriteConfig persists the extension's configuration to the provided filesystem.
	WriteConfig(fsys appendfs.FS) error
	// GetConfig returns the current configuration of the extension.
	GetConfig() any
	// IsRegistered returns true if the extension is properly registered in the system.
	IsRegistered() bool
	// Terminate performs any necessary cleanup when the extension is no longer needed.
	Terminate() error
}
