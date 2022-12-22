package ocfl

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/op/go-logging"
	"io"
)

type creatorFunc func(fs OCFLFSRead) (Extension, error)

type ExtensionFactory struct {
	creators           map[string]creatorFunc
	defaultStorageRoot []Extension
	defaultObject      []Extension
	logger             *logging.Logger
}

func NewExtensionFactory(logger *logging.Logger) (*ExtensionFactory, error) {
	m := &ExtensionFactory{
		creators: map[string]creatorFunc{},
		logger:   logger,
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

func (f *ExtensionFactory) Create(fsys OCFLFSRead) (Extension, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open config.json")
	}
	defer fp.Close()
	data := bytes.NewBuffer(nil)
	io.Copy(data, fp)
	return f.create(fsys, data.Bytes())
}

func (f *ExtensionFactory) create(fsys OCFLFSRead, data []byte) (Extension, error) {
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
	ext, err := creator(fsys)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize extension '%s'", name)
	}
	return ext, nil
}

func (f *ExtensionFactory) CreateExtensions(fsys OCFLFSRead) ([]Extension, error) {
	files, err := fsys.ReadDir(".")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read folder storageroot")
	}
	var result = []Extension{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		sub, err := fsys.SubFS(file.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create subFS %s", file.Name())
		}

		ext, err := f.Create(sub)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create extension %s", file.Name())
		}
		result = append(result, ext)
	}
	return result, nil
}

func (f *ExtensionFactory) LoadExtensions(fsys OCFLFSRead) ([]Extension, error) {
	extensions, err := f.CreateExtensions(fsys)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extensions")
	}
	return extensions, nil
}
