package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
)

type creatorFunc func(config []byte) (Extension, error)

type ExtensionFactory struct {
	creators map[string]creatorFunc
}

func NewExtensionFactory() (*ExtensionFactory, error) {
	m := &ExtensionFactory{
		creators: map[string]creatorFunc{},
	}
	return m, nil
}

func (f *ExtensionFactory) AddCreator(name string, creator creatorFunc) {
	f.creators[name] = creator
}

func (f *ExtensionFactory) Create(config []byte) (Extension, error) {
	var temp = map[string]any{}
	if err := json.Unmarshal(config, &temp); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal config '%s'", string(config))
	}
	nameVar, ok := temp["extensionName"]
	if !ok {
		return nil, errors.Errorf("no field 'extensionName' in config '%s'", string(config))
	}
	name, ok := nameVar.(string)
	if !ok {
		return nil, errors.Errorf("field 'extensionName' is not a string in config '%s'", string(config))
	}
	creator, ok := f.creators[name]
	if !ok {
		return nil, errors.Errorf("unknown extension '%s'", name)
	}
	ext, err := creator(config)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize extension '%s'", name)
	}
	return ext, nil
}
