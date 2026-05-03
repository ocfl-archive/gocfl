package extensionimpl

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	_ "github.com/ocfl-archive/gocfl/v3/pkg/extensions/ext_NNNN_gocfl_extension_manager"
	_ "github.com/ocfl-archive/gocfl/v3/pkg/extensions/ext_initial"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/util"
)

func NewFactory(extensionParams map[string]string, defaultObjectExtensionFS fs.FS, logger ocfllogger.OCFLLogger) (*Factory, error) {
	if defaultObjectExtensionFS == nil {
		defaultObjectExtensionFS = &util.EmptyFS{}
	}
	m := &Factory{
		creators:                 map[string]extension.CreatorFunc{},
		extensionParams:          extensionParams,
		logger:                   logger.With("module", "extensionimpl.Factory"),
		defaultObjectExtensionFS: defaultObjectExtensionFS,
		documentations:           map[string]*string{},
	}
	extension.RegisterWithFactory(m, m.logger)
	return m, nil
}

type Factory struct {
	creators                 map[string]extension.CreatorFunc
	defaultStorageRoot       []extension.Extension
	defaultObject            []extension.Extension
	extensionParams          map[string]string
	logger                   ocfllogger.OCFLLogger
	defaultObjectExtensionFS fs.FS
	documentations           map[string]*string
}

func (f *Factory) GetExtensionDocs() map[string]*string {
	return f.documentations
}

func (f *Factory) AddCreator(name string, creator extension.CreatorFunc, documentation *string) {
	f.creators[name] = creator
	f.documentations[name] = documentation
}

func (f *Factory) RegisterExtension(name string, builder extension.BuilderFunc, documentation *string) {
	f.logger.Debug().Msgf("adding creator for extension %s", name)
	f.AddCreator(name, func(data json.RawMessage) (extension.Extension, error) {
		ext, err := builder()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot create extension %s", name))
		}
		ext = ext.WithLogger(f.logger)
		if err := ext.Load(data); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot load extension %s", name))
		}
		return ext, nil
	}, documentation)
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
	return f.LoadExtensionData(data)
}

func (f *Factory) LoadExtensionData(data json.RawMessage) (extension.Extension, error) {
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
	ext, err := creator(data)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load extension '%s'", name)
	}
	if err := ext.SetParams(f.extensionParams); err != nil {
		return nil, errors.Wrapf(err, "cannot set params for extension '%s'", ext.GetName())
	}
	return ext, nil
}

func (f *Factory) LoadExtensionManager(fsys fs.FS) (extension.ManagerCore, error) {
	if fsys == nil {
		fsys = &util.EmptyFS{}
	}
	var errs = []error{}

	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, errors.Wrapf(err, "cannot read folder %v", fsys)
		}
		files = []fs.DirEntry{}
	}
	var result = []extension.Extension{}
	for _, file := range files {
		if !file.IsDir() {
			f.logger.ValidationError(validation.E067, "extension file '%s' is not a directory", file.Name())
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
			// todo: check with a list of registered extensions against W013. gocfl does not support all registered extensions...
			return nil, errors.Wrapf(err, "cannot load extension '%s'", fName)
			//f.logger.ValidationError(validation.W013, "extension %s not supported by gocfl %s - %s", file.Name(), info.Version, err.Error())
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
				data, err := fs.ReadFile(sub, "config.json")
				if err != nil {
					return nil, errors.Wrapf(err, "cannot read %v/config.json", sub)
				}
				initialExt, err := initialCreator(data)
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

	if initial == nil {
		configData := []byte(fmt.Sprintf(`{"extensionName": "%s", "extension": "%s"}`, extension.DefaultExtensionInitialName, extension.DefaultExtensionManagerName))
		ext, err := f.LoadExtensionData(configData)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot load default initial extension")
		}
		var ok bool
		initial, ok = ext.(extension.Initial)
		if !ok {
			return nil, errors.Errorf("'%s' extension is not an initial extension", extension.DefaultExtensionInitialName)
		}
	}
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

	// something bad had happened. create functional extension manager structure
	if manager == nil {
		data, err := fs.ReadFile(f.defaultObjectExtensionFS, fmt.Sprintf("%s/config.json", extension.DefaultExtensionManagerName))
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, errors.Wrapf(err, "cannot read %v/config.json", f.defaultObjectExtensionFS)
			}
			data = []byte(fmt.Sprintf(`{"extensionName": "%s"}`, extension.DefaultExtensionManagerName))
		}

		ext, err := f.LoadExtensionData(data)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot initialize extension %s", extension.DefaultExtensionManagerName)
		}
		var ok bool
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
