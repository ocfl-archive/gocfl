package extension

import (
	"encoding/json"
	"fmt"

	"io"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	extension2 "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

const InitialName = "initial"
const InitialDescription = "initial extension defines the name of the extension manager"

func GetInitialParams() []*extensionimpl.ExtensionExternalParam {
	return []*extensionimpl.ExtensionExternalParam{
		{
			ExtensionName: InitialName,
			Functions:     []string{"add"},
			Param:         "extension",
			Description:   "name of the extension manager",
		},
	}
}

func NewInitialFS(fsys fs.FS) (*Initial, error) {
	var config = &InitialConfig{
		ExtensionConfig: &extension2.ExtensionConfig{
			ExtensionName: InitialName,
		},
		Extension: extension2.DefaultExtensionManagerName,
	}
	if fsys != nil {
		fp, err := fsys.Open("config.json")
		if err != nil {
			return nil, errors.Wrap(err, "cannot open config.json")
		}
		defer fp.Close()
		data, err := io.ReadAll(fp)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read config.json")
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal InitialConfig '%s'", string(data))
		}
	}
	return NewInitial(config)
}
func NewInitial(config *InitialConfig) (*Initial, error) {
	sl := &Initial{
		InitialConfig: config,
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type InitialEntry struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type InitialConfig struct {
	*extension2.ExtensionConfig
	Extension string `json:"extension"`
}
type Initial struct {
	*InitialConfig
	fsys fs.FS
}

func (sl *Initial) Load(fsys fs.FS) error {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
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

func (sl *Initial) GetFS() fs.FS {
	return sl.fsys
}

func (sl *Initial) GetConfig() any {
	return sl.InitialConfig
}

func (sl *Initial) IsRegistered() bool {
	return true
}

func (sl *Initial) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *Initial) SetParams(params map[string]string) error {
	name := fmt.Sprintf("ext-%s-%s", InitialName, "extension")
	if p, ok := params[name]; ok {
		sl.InitialConfig.Extension = p
	}
	return nil
}

func (sl *Initial) GetName() string { return InitialName }

func (sl *Initial) WriteConfig(streamfs.FS) error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
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
	_ extension2.Extension        = &Initial{}
	_ extension2.ExtensionInitial = &Initial{}
)
