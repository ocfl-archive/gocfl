package extension

import (
	"encoding/json"
	"fmt"

	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const InitialName = "initial"
const InitialDescription = "initial extension defines the name of the extension manager"

func GetInitialParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{
		/*
			{
				ExtensionName: InitialName,
				Functions:     []string{"add"},
				Param:         "extension",
				Description:   "name of the extension manager",
			},
		*/
	}, nil
}

func init() {
	extension.RegisterExtension(InitialName, NewInitial, GetInitialParams)
}

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

type InitialEntry struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type InitialConfig struct {
	*extension.ExtensionConfig
	Extension string `json:"extension"`
}
type Initial struct {
	*InitialConfig
	logger ocfllogger.OCFLLogger
}

func (sl *Initial) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	sl.logger = logger.With("extension", InitialName)
	return sl
}

func (sl *Initial) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, sl.InitialConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal InitialConfig '%s'", string(data))
	}
	return nil
}

func (sl *Initial) Terminate() error {
	return nil
}

func (sl *Initial) SetExtension(ext string) {
	sl.ExtensionName = ext
}

func (sl *Initial) GetExtension() string {
	return sl.InitialConfig.Extension
}

func (sl *Initial) GetConfig() any {
	return sl.InitialConfig
}

func (sl *Initial) IsRegistered() bool {
	return true
}

func (sl *Initial) SetParams(params map[string]string) error {
	name := fmt.Sprintf("ext-%s-%s", InitialName, "extension")
	if p, ok := params[name]; ok {
		sl.InitialConfig.Extension = p
	}
	return nil
}

func (sl *Initial) GetName() string { return InitialName }

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
