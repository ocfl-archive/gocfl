package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/gocfl/v2/version"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io/fs"
)

type creatorFunc func(fsys fs.FS) (Extension, error)

type ExtensionFactory struct {
	creators           map[string]creatorFunc
	defaultStorageRoot []Extension
	defaultObject      []Extension
	extensionParams    map[string]string
	logger             zLogger.ZLogger
}

func NewExtensionFactory(params map[string]string, logger zLogger.ZLogger) (*ExtensionFactory, error) {
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
			//errs = append(errs, errors.Wrapf(err, "cannot create extension %s", file.Name()))
			validation.addValidationWarning(W000, "extension %s not supported by gocfl %s", file.Name(), version.VERSION)
		} else {
			if !ext.IsRegistered() {
				if validation != nil {
					validation.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
				}
			}
			// warning if extension name is different from folder name and extension name is not 'initial'
			// todo: initial should follow the same rule
			if fName != ext.GetName() && fName != "initial" {
				if validation != nil {
					validation.addValidationWarning(W013, "extension '%s' has a different name than the folder", ext.GetName())
				}
			}
			// we have the initial folder, but the extension is not initial. let's create the initial extension
			if fName == "initial" && ext.GetName() != "initial" {
				initialCreator, ok := f.creators[DefaultExtensionInitialName]
				if !ok {
					return nil, errors.Errorf("no initial extension creator (%s) found", DefaultExtensionInitialName)
				}
				initialExt, err := initialCreator(nil)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot create initial extension %s", DefaultExtensionInitialName)
				}
				initial, ok := initialExt.(ExtensionInitial)
				if !ok {
					return nil, errors.Errorf("'%s' extension is not an initial extension", DefaultExtensionInitialName)
				}
				initial.SetExtension(ext.GetName())
				result = append(result, initial)
			}
			result = append(result, ext)
		}
	}
	// find the initial extension and remove it from extension list
	var initial ExtensionInitial
	var manager ExtensionManager
	var result2 = []Extension{}
	for _, ext := range result {
		if ext.GetName() == "initial" {
			var ok bool
			initial, ok = ext.(ExtensionInitial)
			if !ok {
				errs = append(errs, errors.Errorf("extension %s is not an initial extension", ext.GetName()))
			}
			continue
		}
		result2 = append(result2, ext)
	}
	result = result2

	if initial != nil {
		result2 = []Extension{}
		extManagerName := initial.GetExtension()
		for _, ext := range result {
			// extension is the manager extension
			if ext.GetName() == extManagerName {
				var ok bool
				manager, ok = ext.(ExtensionManager)
				if !ok {
					errs = append(errs, errors.Errorf("extension %s is not a manager extension", ext.GetName()))
					result2 = append(result2, ext)
				}
				continue
			}
			result2 = append(result2, ext)
		}
		result = result2
	}
	if manager == nil && initial != nil {
		errs = append(errs, errors.Errorf("manager extension %s found", initial.GetExtension()))
	}

	// something bad had happened. create functional extension manager structure
	if manager == nil {
		// create initial extension if necessary
		if initial == nil {
			initialCreator, ok := f.creators[DefaultExtensionInitialName]
			if !ok {
				return nil, errors.Errorf("no initial extension creator (%s) found", DefaultExtensionInitialName)
			}
			initialExt, err := initialCreator(nil)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create initial extension %s", DefaultExtensionInitialName)
			}
			initial, ok = initialExt.(ExtensionInitial)
			if !ok {
				return nil, errors.Errorf("'%s' extension is not an initial extension", DefaultExtensionInitialName)
			}
			initial.SetExtension(DefaultExtensionManagerName)
		}
		// create default extension manager
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
	// do final steps
	manager.Finalize()
	manager.SetInitial(initial)
	manager.SetFS(fsys, false)
	return manager, errors.Combine(errs...)
}

func (f *ExtensionFactory) LoadExtensions(fsys fs.FS, validation Validation) (ExtensionManager, error) {
	manager, err := f.CreateExtensions(fsys, validation)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extensions")
	}
	return manager, nil
}
