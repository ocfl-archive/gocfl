package storagelayout

import (
	"errors"
	"fmt"
)

type FlatDirect struct{}

func NewFlatDirect(config *Config) (*FlatDirect, error) {
	sl := &FlatDirect{}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}

	return sl, nil

}
func (sl *FlatDirect) Name() string { return "0002-flat-direct-storage-layout" }
func (sl *FlatDirect) ID2Path(id string) (string, error) {
	/*
		if len(id) > MAX_DIR_LEN {
			return "", errors.New(fmt.Sprintf("%s to long (max. %v)", id, MAX_DIR_LEN))
		}
		if FixFilename(id) != id {
			return "", errors.New(fmt.Sprintf("%s contains forbidden characters", id))
		}
	*/
	return id, nil
}
