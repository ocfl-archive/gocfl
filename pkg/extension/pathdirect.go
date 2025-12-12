package extension

import (
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"

	"io"
	"io/fs"
)

const PathDirectName = "NNNN-direct-path-layout"

func NewPathDirectFS(fsys fs.FS) (extension.Extension, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &PathDirectConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewPathDirect(config)
}

func NewPathDirect(config *PathDirectConfig) (*PathDirect, error) {
	sl := &PathDirect{PathDirectConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type PathDirectConfig struct {
	*Config
	fsys fs.FS
}

type PathDirect struct {
	*PathDirectConfig
}

func (sl *PathDirect) Terminate() error {
	return nil
}

func (sl *PathDirect) GetFS() fs.FS {
	return sl.fsys
}

func (sl *PathDirect) GetConfig() any {
	return sl.PathDirectConfig
}

func (sl *PathDirect) IsRegistered() bool {
	return false
}

func (sl *PathDirectConfig) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *PathDirect) SetParams(params map[string]string) error {
	return nil
}

func (sl *PathDirect) GetName() string { return PathDirectName }

func (sl *PathDirect) WriteLayout(fsys fs.FS) error {
	configWriter, err := writefs.Create(fsys, "ocfl_layout.json")
	if err != nil {
		return errors.Wrap(err, "cannot open ocfl_layout.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(struct {
		Extension   string `json:"extension"`
		Description string `json:"description"`
	}{
		Extension:   StorageLayoutFlatDirectName,
		Description: StorageLayoutFlatDirectDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *PathDirect) WriteConfig() error {
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
	if err := jenc.Encode(sl.PathDirectConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *PathDirect) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
	return id, nil
}
func (sl *PathDirect) BuildObjectManifestPath(object object.Object, originalPath string, area string) (string, error) {
	return originalPath, nil
}

// check interface satisfaction
var (
	_ extension.Extension                  = &PathDirect{}
	_ storageroot.ExtensionStorageRootPath = &PathDirect{}
	_ object.ExtensionObjectContentPath    = &PathDirect{}
)
