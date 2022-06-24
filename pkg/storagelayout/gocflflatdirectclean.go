package storagelayout

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"io"
)

const FlatDirectCleanName = "gocfl-flat-direct-clean"

type FlatDirectClean struct {
	*Config
}

func NewFlatDirectClean(config *Config) (*FlatDirectClean, error) {
	sl := &FlatDirectClean{Config: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}

func (sl *FlatDirectClean) Name() string { return FlatDirectCleanName }
func (sl *FlatDirectClean) ID2Path(id string) (string, error) {
	if len(id) > MAX_DIR_LEN {
		return "", errors.New(fmt.Sprintf("%s to long (max. %v)", id, MAX_DIR_LEN))
	}
	return FixFilename(id), nil
}

func (sl *FlatDirectClean) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return emperror.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
