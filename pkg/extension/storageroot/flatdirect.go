package storageroot

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
)

const StorageLayoutFlatDirectName = "0002-flat-direct-storage-layout"

type StorageLayoutFlatDirect struct {
	*Config
}

func NewStorageLayoutFlatDirect(config *Config) (*StorageLayoutFlatDirect, error) {
	sl := &StorageLayoutFlatDirect{Config: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}
func (sl *StorageLayoutFlatDirect) Name() string { return StorageLayoutFlatDirectName }
func (sl *StorageLayoutFlatDirect) ExecuteID(id string) (string, error) {
	/*
		if len(id) > MAX_DIR_LEN {
			return "", errors.New(fmt.Sprintf("%s to long (max. %v)", id, MAX_DIR_LEN))
		}
		if CleanPath(id) != id {
			return "", errors.New(fmt.Sprintf("%s contains forbidden characters", id))
		}
	*/
	return id, nil
}
func (sl *StorageLayoutFlatDirect) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
