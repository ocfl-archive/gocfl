package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
)

type Type int

const (
	StorageRootExtension Type = 1 << 0
	ObjectExtension      Type = 1 << 1
)

type creatorFunc func(config []byte) (Extension, error)

type Extension interface {
	GetName() string
	IsStorageRoot() bool
	IsObject() bool
	Do(object ocfl.Object) (any, error)
}

type Manager struct {
	creators              map[string]creatorFunc
	objectExtensions      []Extension
	storageRootExtensions []Extension
}

func NewManager() (*Manager, error) {
	m := &Manager{
		storageRootExtensions: []Extension{},
		objectExtensions:      []Extension{},
		creators:              map[string]creatorFunc{},
	}
	return m, nil
}

func (f *Manager) AddCreator(name string, creator creatorFunc) {
	f.creators[name] = creator
}

func (f *Manager) CreateStorageRoot(config []byte) error {
	ext, err := f.create(config)
	if err != nil {
		return err
	}
	if !ext.IsStorageRoot() {
		return errors.Errorf("extension '%s' is not a storage root extension")
	}
	f.storageRootExtensions = append(f.storageRootExtensions, ext)
	return nil
}

func (f *Manager) CreateObject(config []byte) error {
	ext, err := f.create(config)
	if err != nil {
		return err
	}
	if !ext.IsObject() {
		return errors.Errorf("extension '%s' is not an object extension")
	}
	f.objectExtensions = append(f.objectExtensions, ext)
	return nil
}

func (f *Manager) create(config []byte) (Extension, error) {
	var temp = map[string]any{}
	if err := json.Unmarshal(config, temp); err != nil {
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
