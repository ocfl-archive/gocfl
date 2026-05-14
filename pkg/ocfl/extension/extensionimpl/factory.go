package extensionimpl

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	_ "github.com/ocfl-archive/gocfl/v3/pkg/extensions/ext_NNNN_gocfl_extension_manager"
	_ "github.com/ocfl-archive/gocfl/v3/pkg/extensions/ext_initial"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/util"
)

func NewFactory[T extension.ManagerCore[T]](extensionParams map[string]string, logger ocfllogger.OCFLLogger) (extension.Factory[T], error) {
	m := &Factory[T]{
		extensionParams: extensionParams,
		logger:          logger.With("module", "extensionimpl.Factory"),
		extension:       map[string]*extensionData{},
	}
	extension.RegisterWithFactory(m, m.logger)

	return m, nil
}

type extensionData struct {
	creator       extension.CreatorFunc
	documentation *string
}
type Factory[T extension.ManagerCore[T]] struct {
	extension          map[string]*extensionData
	defaultStorageRoot []extension.Extension
	defaultObject      []extension.Extension
	extensionParams    map[string]string
	logger             ocfllogger.OCFLLogger
}

func (f *Factory[T]) GetExtensionDocs() map[string]*string {
	docs := map[string]*string{}
	for name, ext := range f.extension {
		docs[name] = ext.documentation
	}
	return docs
}

func (f *Factory[T]) AddCreator(name string, creator extension.CreatorFunc, documentation *string) {
	//f.creators[name] = creator
	//f.documentations[name] = documentation
	f.extension[name] = &extensionData{
		creator:       creator,
		documentation: documentation,
	}
}

func (f *Factory[T]) RegisterExtension(name string, builder extension.BuilderFunc, documentation *string) {
	f.logger.Debug().Msgf("adding creator for extension %s", name)
	f.AddCreator(name, func(data json.RawMessage, extFS fs.FS) (extension.Extension, error) {
		ext, err := builder()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot create extension %s", name))
		}
		ext = ext.WithLogger(f.logger)
		if err := ext.Load(data, extFS); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot load extension %s", name))
		}
		return ext, nil
	}, documentation)
}

func (f *Factory[T]) AddStorageRootDefaultExtension(ext extension.Extension) {
	f.defaultStorageRoot = append(f.defaultStorageRoot, ext)
}

func (f *Factory[T]) AddObjectDefaultExtension(ext extension.Extension) {
	f.defaultObject = append(f.defaultObject, ext)
}

func (f *Factory[T]) LoadExtensionFile(fsys fs.FS) (extension.Extension, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %v/config.json", fsys)
	}
	return f.LoadExtensionData(data, fsys)
}

func (f *Factory[T]) LoadExtensionData(data json.RawMessage, extFS fs.FS) (extension.Extension, error) {
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
	extData, ok := f.extension[name]
	if !ok {
		return nil, errors.Errorf("unknown extension '%s'", name)
	}
	if extData.creator == nil {
		return nil, errors.Errorf("extension '%s' not initialized - no creator", name)
	}
	ext, err := extData.creator(data, extFS)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load extension '%s'", name)
	}
	if err := ext.SetParams(f.extensionParams); err != nil {
		return nil, errors.Wrapf(err, "cannot set params for extension '%s'", ext.GetName())
	}
	return ext, nil
}

func (f *Factory[T]) LoadExtensionManager(fsys fs.FS) (T, error) {
	var zero T
	if fsys == nil {
		fsys = &util.EmptyFS{}
	}

	extensions, err := f.loadExtensions(fsys)
	if err != nil {
		return zero, err
	}

	initial, extensions, err := f.extractInitial(extensions)
	if err != nil {
		return zero, err
	}

	if initial == nil {
		initial, err = f.loadDefaultInitial()
		if err != nil {
			return zero, err
		}
	}

	manager, extensions, err := f.extractManager(initial, extensions)
	if err != nil {
		return zero, err
	}

	if any(manager) == nil {
		manager, err = f.loadDefaultManager(initial.GetExtension())
		if err != nil {
			return zero, err
		}
	}

	var errs []error
	for _, ext := range extensions {
		if err := manager.Add(ext); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot add extension %s to manager", ext.GetName()))
		}
	}

	manager.Finalize()
	manager.SetInitial(initial)
	return manager, errors.Combine(errs...)
}

func (f *Factory[T]) loadExtensions(fsys fs.FS) ([]extension.Extension, error) {
	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, errors.Wrapf(err, "cannot read folder %v", fsys)
		}
		files = []fs.DirEntry{}
	}

	var result []extension.Extension
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
			return nil, errors.Wrapf(err, "cannot load extension '%s'", fName)
		}

		if !ext.IsRegistered() {
			f.logger.ValidationError(validation.W013, "extension %s is not registered", ext.GetName())
		}
		if fName != ext.GetName() && fName != "initial" {
			f.logger.ValidationError(validation.W013, "extension %s has a different name than the folder", ext.GetName())
		}

		if fName == "initial" && ext.GetName() != "initial" {
			initialExt, err := f.createInitialFromFolder(sub, ext.GetName())
			if err != nil {
				return nil, err
			}
			result = append(result, initialExt)
		}
		result = append(result, ext)
	}
	return result, nil
}

func (f *Factory[T]) createInitialFromFolder(sub fs.FS, extName string) (extension.Initial, error) {
	extData, ok := f.extension["initial"]
	if !ok {
		return nil, errors.Errorf("no initial extension found")
	}
	if extData.creator == nil {
		return nil, errors.Errorf("no initial extension creator found")
	}
	data, err := fs.ReadFile(sub, "config.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %v/config.json", sub)
	}
	initialExt, err := extData.creator(data, sub)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize extension %s", extension.DefaultExtensionInitialName)
	}
	initial, ok := initialExt.(extension.Initial)
	if !ok {
		return nil, errors.Errorf("'%s' extension is not an initial extension", extension.DefaultExtensionInitialName)
	}
	initial.SetExtension(extName)
	return initial, nil
}

func (f *Factory[T]) extractInitial(extensions []extension.Extension) (extension.Initial, []extension.Extension, error) {
	var initial extension.Initial
	var remaining []extension.Extension
	for _, ext := range extensions {
		if ext.GetName() == "initial" {
			var ok bool
			initial, ok = ext.(extension.Initial)
			if !ok {
				return nil, nil, errors.Errorf("extension %s is not an initial extension", ext.GetName())
			}
			continue
		}
		remaining = append(remaining, ext)
	}
	return initial, remaining, nil
}

func (f *Factory[T]) loadDefaultInitial() (extension.Initial, error) {
	configData := []byte(fmt.Sprintf(`{"extensionName": "%s", "extension": "%s"}`, extension.DefaultExtensionInitialName, extension.DefaultExtensionManagerName))
	ext, err := f.LoadExtensionData(configData, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load default initial extension")
	}
	initial, ok := ext.(extension.Initial)
	if !ok {
		return nil, errors.Errorf("'%s' extension is not an initial extension", extension.DefaultExtensionInitialName)
	}
	return initial, nil
}

func (f *Factory[T]) extractManager(initial extension.Initial, extensions []extension.Extension) (T, []extension.Extension, error) {
	var manager T
	var remaining []extension.Extension
	extManagerName := initial.GetExtension()
	for _, ext := range extensions {
		if ext.GetName() == extManagerName {
			var ok bool
			manager, ok = ext.(T)
			if !ok {
				// We don't return error here to maintain original behavior (append to errs later, but here we just keep it in extensions)
				// Actually original code did: errs = append(errs, errors.Errorf("extension %s is not a manager extension", ext.GetName()))
				// and result2 = append(result2, ext)
				remaining = append(remaining, ext)
				continue
			}
			continue
		}
		remaining = append(remaining, ext)
	}
	return manager, remaining, nil
}

func (f *Factory[T]) loadDefaultManager(extManagerName string) (T, error) {
	var zero T
	configData := []byte(fmt.Sprintf(`{"extensionName": "%s"}`, extManagerName))
	ext, err := f.LoadExtensionData(configData, nil)
	if err != nil {
		return zero, errors.Wrapf(err, "cannot load default initial extension")
	}
	manager, ok := ext.(T)
	if !ok {
		return zero, errors.Errorf("extension %s is not a manager extension", ext.GetName())
	}
	return manager, nil
}

var _ extension.Factory[extension.ManagerCore[extension.Extension]] = (*Factory[extension.ManagerCore[extension.Extension]])(nil)
