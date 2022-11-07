package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
)

const StorageLayoutFlatDirectName = "0002-flat-direct-storage-layout"

type StorageLayoutFlatDirectConfig struct {
	*ocfl.ExtensionConfig
}
type StorageLayoutFlatDirect struct {
	*StorageLayoutFlatDirectConfig
}

func NewStorageLayoutFlatDirectFS(fs ocfl.OCFLFS) (*StorageLayoutFlatDirect, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &StorageLayoutFlatDirectConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewStorageLayoutFlatDirect(config)
}
func NewStorageLayoutFlatDirect(config *StorageLayoutFlatDirectConfig) (*StorageLayoutFlatDirect, error) {
	sl := &StorageLayoutFlatDirect{config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}
func (sl *StorageLayoutFlatDirect) GetName() string { return StorageLayoutFlatDirectName }
func (sl *StorageLayoutFlatDirect) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *StorageLayoutFlatDirect) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	return id, nil
}
