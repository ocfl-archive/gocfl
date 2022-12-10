package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
)

const PathDirectName = "NNNN-direct-path-layout"

type PathDirectConfig struct {
	*Config
	fs ocfl.OCFLFS
}

type PathDirect struct {
	*PathDirectConfig
}

func NewPathDirectFS(fs ocfl.OCFLFS) (ocfl.Extension, error) {
	fp, err := fs.Open("config.json")
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

func (sl *PathDirectConfig) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}
func (sl *PathDirect) GetName() string { return PathDirectName }
func (sl *PathDirect) WriteLayout(fs ocfl.OCFLFS) error {
	configWriter, err := fs.Create("ocfl_layout.json")
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
	configWriter, err := sl.fs.Create("config.json")
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

func (sl *PathDirect) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	return id, nil
}
func (sl *PathDirect) BuildObjectContentPath(object ocfl.Object, id string) (string, error) {
	return id, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                  = &PathDirect{}
	_ ocfl.ExtensionStoragerootPath   = &PathDirect{}
	_ ocfl.ExtensionObjectContentPath = &PathDirect{}
)
