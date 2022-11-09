package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"io"
)

type creatorFunc func(fs OCFLFS) (Extension, error)

type ExtensionFactory struct {
	creators           map[string]creatorFunc
	defaultStorageRoot []Extension
	defaultObject      []Extension
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

func (f *ExtensionFactory) AddStorageRootDefaultExtension(ext Extension) {
	f.defaultStorageRoot = append(f.defaultStorageRoot, ext)
}

func (f *ExtensionFactory) AddObjectDefaultExtension(ext Extension) {
	f.defaultObject = append(f.defaultObject, ext)
}

func (f *ExtensionFactory) Create(fs OCFLFS) (Extension, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config.json")
	}
	var temp = map[string]any{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal config '%s'", string(data))
	}
	nameVar, ok := temp["extensionName"]
	if !ok {
		return nil, errors.Errorf("no field 'extensionName' in config '%s'", string(data))
	}
	name, ok := nameVar.(string)
	if !ok {
		return nil, errors.Errorf("field 'extensionName' is not a string in config '%s'", string(data))
	}
	creator, ok := f.creators[name]
	if !ok {
		return nil, errors.Errorf("unknown extension '%s'", name)
	}
	ext, err := creator(fs)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize extension '%s'", name)
	}
	return ext, nil
}
