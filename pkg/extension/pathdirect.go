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
func (sl *PathDirect) IsObjectExtension() bool      { return false }
func (sl *PathDirect) IsStoragerootExtension() bool { return true }
func (sl *PathDirect) GetName() string              { return PathDirectName }
func (sl *PathDirect) WriteConfig(fs ocfl.OCFLFS) error {
	configWriter, err := fs.Create("config.json")
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
func (sl *PathDirect) BuildObjectContentPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	return id, nil
}
