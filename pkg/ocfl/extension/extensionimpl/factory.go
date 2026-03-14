package extensionimpl

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/info"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

type Factory struct {
	creators           map[string]extension.CreatorFunc
	defaultStorageRoot []extension.Extension
	defaultObject      []extension.Extension
	extensionParams    map[string]string
	logger             ocfllogger.OCFLLogger
}

func NewFactory(extensionParams map[string]string, logger ocfllogger.OCFLLogger) (*Factory, error) {
	m := &Factory{
		creators:        map[string]extension.CreatorFunc{},
		extensionParams: extensionParams,
		logger:          logger.With("module", "extensionimpl.Factory"),
	}
	extension.RegisterWithFactory(m, m.logger)
	return m, nil
}

func (f *Factory) AddCreator(name string, creator extension.CreatorFunc) {
	f.creators[name] = creator
}

func (f *Factory) RegisterExtension(name string, builder extension.BuilderFunc) {
	f.logger.Debug().Msgf("adding creator for extension %s", name)
	f.AddCreator(name, func(fsys fs.FS) (extension.Extension, error) {
		ext, err := builder()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot create extension %s", name))
		}
		ext = ext.WithLogger(f.logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot load extension %s", name))
		}
		return ext, nil
	})
}

func (f *Factory) AddStorageRootDefaultExtension(ext extension.Extension) {
	f.defaultStorageRoot = append(f.defaultStorageRoot, ext)
}

func (f *Factory) AddObjectDefaultExtension(ext extension.Extension) {
	f.defaultObject = append(f.defaultObject, ext)
}

func (f *Factory) LoadExtensionFile(fsys fs.FS) (extension.Extension, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %v/config.json", fsys)
	}
	return f.LoadExtensionData(fsys, data)
}

func (f *Factory) LoadExtensionData(fsys fs.FS, data []byte) (extension.Extension, error) {
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
		return nil, errors.Wrapf(err, "cannot load extension '%s'", name)
	}
	if err := ext.SetParams(f.extensionParams); err != nil {
		return nil, errors.Wrapf(err, "cannot set params for extension '%s'", ext.GetName())
	}
	return ext, nil
}

func (f *Factory) LoadExtensionManager(fsys fs.FS) (extension.ManagerCore, error) {
	var errs = []error{}
	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			files = []fs.DirEntry{}
		} else {
			return nil, errors.Wrapf(err, "cannot read folder %v", fsys)
		}
	}
	var result = []extension.Extension{}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		fName := file.Name()
		sub, err := fs.Sub(fsys, fName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create subFS %s", file.Name())
		}

		ext, err := f.LoadExtensionFile(sub)
		if err != nil {
			//errs = append(errs, errors.Wrapf(err, "cannot create extension %s", file.Name()))
			f.logger.ValidationError(validation.W000, "extension %s not supported by gocfl %s - %s", file.Name(), info.Version, err.Error())
		} else {
			if !ext.IsRegistered() {
				f.logger.ValidationError(validation.W013, "extension %s is not registered", ext.GetName())
			}
			// warning if extension name is different from folder name and extension name is not 'initial'
			// todo: initial should follow the same rule
			if fName != ext.GetName() && fName != "initial" {
				f.logger.ValidationError(validation.W013, "extension %s has a different name than the folder", ext.GetName())
			}
			// we have the initial folder, but the extension is not initial. let's create the initial extension
			if fName == "initial" && ext.GetName() != "initial" {
				initialCreator, ok := f.creators[extension.DefaultExtensionInitialName]
				if !ok {
					return nil, errors.Errorf("no initial extension creator (%s) found", extension.DefaultExtensionInitialName)
				}
				initialExt, err := initialCreator(fsys)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot initialize extension %s", extension.DefaultExtensionInitialName)
				}
				initial, ok := initialExt.(extension.Initial)
				if !ok {
					return nil, errors.Errorf("'%s' extension is not an initial extension", extension.DefaultExtensionInitialName)
				}
				initial.SetExtension(ext.GetName())
				result = append(result, initial)
			}
			result = append(result, ext)
		}
	}
	// find the initial extension and remove it from extension list
	var initial extension.Initial
	var manager object.ExtensionManager
	var result2 = []extension.Extension{}
	for _, ext := range result {
		if ext.GetName() == "initial" {
			var ok bool
			initial, ok = ext.(extension.Initial)
			if !ok {
				errs = append(errs, errors.Errorf("extension %s is not an initial extension", ext.GetName()))
			}
			continue
		}
		result2 = append(result2, ext)
	}
	result = result2

	if initial != nil {
		result2 = []extension.Extension{}
		extManagerName := initial.GetExtension()
		for _, ext := range result {
			// extension is the manager extension
			if ext.GetName() == extManagerName {
				var ok bool
				manager, ok = ext.(object.ExtensionManager)
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
			initialCreator, ok := f.creators[extension.DefaultExtensionInitialName]
			if !ok {
				return nil, errors.Errorf("no initial extension creator (%s) found", extension.DefaultExtensionInitialName)
			}
			initialFS, _ := fs.Sub(fsys, extension.DefaultExtensionInitialName)
			initialExt, err := initialCreator(initialFS)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot initialize extension %s", extension.DefaultExtensionInitialName)
			}
			initial, ok = initialExt.(extension.Initial)
			if !ok {
				return nil, errors.Errorf("'%s' extension is not an initial extension", extension.DefaultExtensionInitialName)
			}
			initial.SetExtension(extension.DefaultExtensionManagerName)
		}
		// create default extension manager
		creator, ok := f.creators[extension.DefaultExtensionManagerName]
		if !ok {
			return nil, errors.Errorf("no default extension manager (%s) found", extension.DefaultExtensionManagerName)
		}
		extensionManagerFS, _ := fs.Sub(fsys, extension.DefaultExtensionManagerName)
		ext, err := creator(extensionManagerFS)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot initialize extension %s", extension.DefaultExtensionManagerName)
		}
		manager, ok = ext.(object.ExtensionManager)
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
	//manager.SetFS(fsys, false)
	return manager, errors.Combine(errs...)
}

var _ extension.Factory = (*Factory)(nil)
