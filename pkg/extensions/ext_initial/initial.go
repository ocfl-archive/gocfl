// Package ext_initial implements the OCFL "initial" extension.
// This extension allows indication that the semantics of a particular extension takes precedence over all other extensions.
package ext_initial

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

// InitialName is the unique name of the extension.
const InitialName = "initial"

// InitialDescription is a short description of the extension's purpose.
const InitialDescription = "initial extension defines the name of the extension manager"

// InitialDoc contains the documentation for the extension.
//
//go:embed initial.md
var InitialDoc string

// GetInitialParams returns the external parameters supported by this extension.
func GetInitialParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{}, nil
}

func init() {
	extension.RegisterExtensionStorageRoot(InitialName, NewInitial, GetInitialParams, &InitialDoc)
	extension.RegisterExtensionObject(InitialName, NewInitial, GetInitialParams, &InitialDoc)
}

// NewInitial creates a new instance of the Initial extension with default configuration.
func NewInitial() (extension.Extension, error) {
	var config = &InitialConfig{
		ExtensionConfig: &extension.ExtensionConfig{
			ExtensionName: InitialName,
		},
		Extension: extension.DefaultExtensionManagerName,
	}
	sl := &Initial{
		InitialConfig: config,
	}
	return sl, nil
}

// InitialConfig holds the configuration for the initial extension.
type InitialConfig struct {
	*extension.ExtensionConfig
	Extension string `json:"extension"`
}

// Initial represents the initial extension instance.
type Initial struct {
	*InitialConfig
	logger ocfllogger.OCFLLogger
}

// WithLogger sets the logger for the extension and returns the extension itself.
func (sl *Initial) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	sl.logger = logger.With("extension", InitialName)
	return sl
}

// Load unmarshals the configuration data and initializes the extension.
func (sl *Initial) Load(data json.RawMessage, _ fs.FS) error {
	if err := json.Unmarshal(data, sl.InitialConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal InitialConfig '%s'", string(data))
	}
	return nil
}

// Terminate cleans up any resources used by the extension.
func (sl *Initial) Terminate() error {
	return nil
}

// SetExtension sets the functional extension name.
func (sl *Initial) SetExtension(ext string) {
	sl.ExtensionName = ext
}

// GetExtension returns the functional extension name.
func (sl *Initial) GetExtension() string {
	return sl.InitialConfig.Extension
}

// GetConfig returns the extension's configuration.
func (sl *Initial) GetConfig() any {
	return sl.InitialConfig
}

// IsRegistered returns true if the extension is registered.
func (sl *Initial) IsRegistered() bool {
	return true
}

// SetParams sets external parameters for the extension.
func (sl *Initial) SetParams(params map[string]string) error {
	name := fmt.Sprintf("ext-%s-%s", InitialName, "extension")
	if p, ok := params[name]; ok {
		sl.InitialConfig.Extension = p
	}
	return nil
}

// GetName returns the name of the extension.
func (sl *Initial) GetName() string { return InitialName }

// WriteConfig writes the extension's configuration to the storage root.
func (sl *Initial) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.InitialConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

// check interface satisfaction
var (
	_ extension.Extension = &Initial{}
	_ extension.Initial   = &Initial{}
)
