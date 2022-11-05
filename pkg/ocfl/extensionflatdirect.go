package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
)

const StorageLayoutFlatDirectName = "0002-flat-direct-storage-layout"

type StorageLayoutFlatDirectConfig struct {
	*ExtensionConfig
}
type StorageLayoutFlatDirect struct {
	*StorageLayoutFlatDirectConfig
}

func NewStorageLayoutFlatDirect(data []byte) (*StorageLayoutFlatDirect, error) {
	var config = &StorageLayoutFlatDirectConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal StorageLayoutDirectCleanConfig '%s'", string(data))
	}

	sl := &StorageLayoutFlatDirect{config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}
func (sl *StorageLayoutFlatDirect) GetName() string { return StorageLayoutFlatDirectName }
func (sl *StorageLayoutFlatDirect) BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error) {
	return id, nil
}
func (sl *StorageLayoutFlatDirect) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
