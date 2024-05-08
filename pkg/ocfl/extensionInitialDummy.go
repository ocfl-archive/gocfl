package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
)

func NewInitialDummyFS(fsys fs.FS) (Extension, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &ExtensionManagerConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewInitialDummy(config)
}

func NewInitialDummy(config *ExtensionManagerConfig) (*InitialDummy, error) {
	sl := &InitialDummy{ExtensionManagerConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type InitialDummy struct {
	*ExtensionManagerConfig
}

func (dummy *InitialDummy) Terminate() error {
	return nil
}

func (dummy *InitialDummy) GetFS() fs.FS {
	//TODO implement me
	panic("implement me")
}

func (dummy *InitialDummy) GetConfig() any {
	return dummy.ExtensionManagerConfig
}

func (dummy *InitialDummy) IsRegistered() bool {
	return false
}

func (dummy *InitialDummy) GetName() string {
	return "dummy"
}

func (dummy *InitialDummy) SetParams(params map[string]string) error {
	return nil
}

func (dummy *InitialDummy) GetConfigString() string {
	//TODO implement me
	panic("implement me")
}

func (dummy *InitialDummy) WriteConfig() error {
	panic("never call me")
}

func (dummy *InitialDummy) SetFS(fsys fs.FS, create bool) {
	panic("never call me")
}

var (
	_ Extension = &InitialDummy{}
)
