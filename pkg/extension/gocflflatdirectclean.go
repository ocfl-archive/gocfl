package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
)

const FlatDirectCleanName = "gocfl-flat-direct-clean"

type FlatDirectClean struct {
	*FlatDirectCleanConfig
}

type FlatDirectCleanConfig struct {
	*Config
	MaxLen int `json:"maxLen,omitempty"`
}

func NewFlatDirectClean(config *FlatDirectCleanConfig) (*FlatDirectClean, error) {
	if config.MaxLen == 0 {
		config.MaxLen = 255
	}
	sl := &FlatDirectClean{FlatDirectCleanConfig: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}

func (sl *FlatDirectClean) Name() string { return FlatDirectCleanName }
func (sl *FlatDirectClean) ID2Path(id string) (string, error) {
	id = FixFilename(id)
	if len(id) > sl.MaxLen {
		return "", errors.New(fmt.Sprintf("%s to long (max. %v)", id, sl.MaxLen))
	}
	return id, nil
}

func (sl *FlatDirectClean) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
