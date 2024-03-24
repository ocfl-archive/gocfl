package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io/fs"
)

type creatorFunc func(fsys fs.FS) (Extension, error)

type ExtensionFactory struct {
	creators           map[string]creatorFunc
	defaultStorageRoot []Extension
	defaultObject      []Extension
	extensionParams    map[string]string
	logger             zLogger.ZWrapper
}

func NewExtensionFactory(params map[string]string, logger zLogger.ZWrapper) (*ExtensionFactory, error) {
	m := &ExtensionFactory{
		creators:        map[string]creatorFunc{},
		extensionParams: params,
		logger:          logger,
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

func (f *ExtensionFactory) Create(fsys fs.FS) (Extension, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %v/config.json", fsys)
	}
	return f.create(fsys, data)
}

func (f *ExtensionFactory) create(fsys fs.FS, data []byte) (Extension, error) {
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
	if err := ext.SetParams(f.extensionParams); err != nil {
		return nil, errors.Wrapf(err, "cannot set params for extension '%s'", ext.GetName())
	}
	return ext, nil
}

func (f *ExtensionFactory) CreateExtensions(fsys fs.FS, validation Validation) (ExtensionManager, error) {
	var errs = []error{}
	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read folder storageroot")
	}
	var result = []Extension{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		fName := file.Name()
		sub, err := fs.Sub(fsys, fName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create subFS %s", file.Name())
		}

		ext, err := f.Create(sub)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot create extension %s", file.Name()))
		} else {
			if !ext.IsRegistered() {
				if validation != nil {
					validation.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
				}
			}
			result = append(result, ext)
		}
	}
	// find the initial extension
	var initial ExtensionInitial
	var manager ExtensionManager
	var result2 = []Extension{}
	for _, ext := range result {
		if ext.GetName() != "initial" {
			result2 = append(result2, ext)
			continue
		}
		var ok bool
		initial, ok = ext.(ExtensionInitial)
		if !ok {
			errs = append(errs, errors.Errorf("extension %s is not an initial extension", ext.GetName()))
		}
	}
	result = result2

	if initial != nil {
		result2 = []Extension{}
		extManagerName := initial.GetExtension()
		for _, ext := range result {
			if ext.GetName() != extManagerName {
				result2 = append(result2, ext)
				continue
			}
			var ok bool
			manager, ok = ext.(ExtensionManager)
			if !ok {
				errs = append(errs, errors.Errorf("extension %s is not a manager extension", ext.GetName()))
				result2 = append(result2, ext)
			}
		}
		result = result2
	}
	if manager == nil {
		creator, ok := f.creators[DefaultExtensionManagerName]
		if !ok {
			return nil, errors.Errorf("no default extension manager (%s) found", DefaultExtensionManagerName)
		}
		ext, err := creator(nil)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create default extension manager")
		}
		manager, ok = ext.(ExtensionManager)
		if !ok {
			return nil, errors.Errorf("default extension manager is not a manager extension")
		}
	}
	for _, ext := range result {
		if err := manager.Add(ext); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot add extension %s to manager", ext.GetName()))
		}
	}
	manager.Finalize()
	manager.SetInitial(initial)
	manager.SetFS(fsys)
	return manager, errors.Combine(errs...)
}

func (f *ExtensionFactory) LoadExtensions(fsys fs.FS, validation Validation) (ExtensionManager, error) {
	manager, err := f.CreateExtensions(fsys, validation)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extensions")
	}
	return manager, nil
}
