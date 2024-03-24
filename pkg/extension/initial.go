package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

const InitialName = "initial"
const InitialDescription = "initial extension defines the name of the extension manager"

func GetInitialParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
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
		ExtensionConfig: &ocfl.ExtensionConfig{
			ExtensionName: InitialName,
		},
		Extension: "NNNN-gocfl-extension-manager",
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
	*ocfl.ExtensionConfig
	Extension string `json:"extension"`
}
type Initial struct {
	*InitialConfig
	fsys fs.FS
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

func (sl *Initial) SetFS(fsys fs.FS) {
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

func (sl *Initial) WriteConfig() error {
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
	_ ocfl.Extension        = &Initial{}
	_ ocfl.ExtensionInitial = &Initial{}
)
