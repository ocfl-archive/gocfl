package object

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
)

const PathDirectName = "NNNN-direct-path-object"

type PathDirectConfig struct {
	*Config
}

type PathDirect struct {
	*PathDirectConfig
}

func NewPathDirect(config *PathDirectConfig) (*PathDirect, error) {
	sl := &PathDirect{PathDirectConfig: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}
func (sl *PathDirect) Name() string { return PathDirectName }
func (sl *PathDirect) ExecutePath(id string) (string, error) {
	return id, nil
}
func (sl *PathDirect) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.PathDirectConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
