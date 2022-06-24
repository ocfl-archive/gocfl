package storagelayout

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"io"
)

const FlatDirectName = "0002-flat-direct-storage-layout"

type FlatDirect struct {
	*Config
}

func NewFlatDirect(config *Config) (*FlatDirect, error) {
	sl := &FlatDirect{Config: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}
func (sl *FlatDirect) Name() string { return FlatDirectName }
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
func (sl *FlatDirect) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return emperror.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
