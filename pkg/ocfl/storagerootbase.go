package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"io/fs"
	"path/filepath"
)

type StorageRootBase struct {
	ctx              context.Context
	fs               OCFLFS
	extensionFactory *ExtensionFactory
	extensionManager *ExtensionManager
	changed          bool
	logger           *logging.Logger
	version          OCFLVersion
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, defaultExtension Extension, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootBase, error) {
	var err error
	ocfl := &StorageRootBase{
		ctx:              ctx,
		fs:               fs,
		extensionFactory: extensionFactory,
		version:          defaultVersion,
		logger:           logger,
	}
	ocfl.extensionManager, err = NewExtensionManager()
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension manager")
	}
	if defaultExtension != nil {
		ocfl.extensionManager.Add(defaultExtension)
	}

	if err := ocfl.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

var errVersionMultiple = errors.New("multiple version files found")
var errVersionNone = errors.New("no version file found")

func (osr *StorageRootBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	addValidationErrors(osr.ctx, GetValidationError(osr.version, errno).AppendDescription(format, a...))
}

func (osr *StorageRootBase) Init() error {
	var err error
	osr.logger.Debug()

	entities, err := osr.fs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot get root entities")
	}
	if len(entities) == 0 {
		rootConformanceDeclaration := "ocfl_" + string(osr.version)
		rootConformanceDeclarationFile := "0=" + rootConformanceDeclaration
		rcd, err := osr.fs.Create(rootConformanceDeclarationFile)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", rootConformanceDeclarationFile)
		}
		defer rcd.Close()
		if _, err := rcd.Write([]byte(rootConformanceDeclaration + "\n")); err != nil {
			return errors.Wrapf(err, "cannot write into %s", rootConformanceDeclarationFile)
		}
		if err := osr.extensionManager.StoreConfigs(osr.fs.SubFS("extensions")); err != nil {
			return errors.Wrap(err, "cannot store extension configs")
		}
	} else {
		// read storage layout from extension folder...
		exts, err := osr.fs.ReadDir("extensions")
		if err != nil {
			// if directory does not exist - no problem
			if err != fs.ErrNotExist {
				return errors.Wrap(err, "cannot read extensions folder")
			}
			exts = []fs.DirEntry{}
		}
		for _, extFolder := range exts {
			extFolder := fmt.Sprintf("extensions/%s", extFolder.Name())
			ext, err := osr.extensionFactory.Create(osr.fs.SubFS(extFolder))
			if err != nil {
				return errors.Wrapf(err, "cannot create extension for config '%s'", extFolder)
			}
			if err := osr.extensionManager.Add(ext); err != nil {
				return errors.Wrapf(err, "cannot add extension '%s' to manager", extFolder)
			}
		}
	}
	return nil
}
func (osr *StorageRootBase) Context() context.Context { return osr.ctx }

func (osr *StorageRootBase) StoreExtensionConfig(name string, config any) error {
	extConfig := fmt.Sprintf("extensions/%s/config.json", name)
	cfgJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "cannot marshal extension %s config [%v]", name, config)
	}
	w, err := osr.fs.Create(extConfig)
	if err != nil {
		return errors.Wrapf(err, "cannot create file %s", extConfig)
	}
	if _, err := w.Write(cfgJson); err != nil {
		return errors.Wrapf(err, "cannot write file %s - %s", extConfig, string(cfgJson))
	}
	return nil
}

func (osr *StorageRootBase) GetFiles() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read folders of storage root")
	}
	var result = []string{}
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		result = append(result, dir.Name())
	}
	return result, nil
}

func (osr *StorageRootBase) GetFolders() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read folders of storage root")
	}
	var result = []string{}
	for _, dir := range dirs {
		if !dir.IsDir() || dir.Name() == "." || dir.Name() == ".." {
			continue
		}
		result = append(result, dir.Name())
	}
	return result, nil
}

// all folder trees, which end in a folder containing a file
func (osr *StorageRootBase) GetObjectFolders() ([]string, error) {
	var recurse func(base string) ([]string, error)
	recurse = func(base string) ([]string, error) {
		des, err := osr.fs.ReadDir(base)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read content of %s", base)
		}
		result := []string{}
		for _, de := range des {
			currPath := filepath.ToSlash(filepath.Join(base, de.Name()))
			// directory hierarchy must contain only folders, no files --> if file exists, it's an object folder
			if de.IsDir() {
				dirs, err := recurse(currPath)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot recurse into %s", currPath)
				}
				result = append(result, dirs...)
			} else {
				if de.Name() == "." || de.Name() == ".." {
					continue
				}
				result = append(result, base)
				break
			}

		}
		return result, nil
	}
	dirs, err := osr.GetFolders()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result = []string{}
	for _, dir := range dirs {
		if dir == "extensions" {
			continue
		}
		dirs, err := recurse(dir)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = append(result, dirs...)
	}
	return result, nil
}

func (osr *StorageRootBase) OpenObjectFolder(folder string) (Object, error) {
	version, err := getVersion(osr.ctx, osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		osr.addValidationError(E003, "no version in folder %s", folder)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s", folder)
	}
	return NewObject(osr.ctx, osr.fs.SubFS(folder), version, "", osr.logger)
}

func (osr *StorageRootBase) OpenObject(id string) (Object, error) {
	folder, err := osr.extensionManager.BuildStoragerootPath(osr, id)
	version, err := getVersion(osr.ctx, osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		return NewObject(osr.ctx, osr.fs.SubFS(folder), osr.version, id, osr.logger)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s for [%s]", folder, id)
	}
	return NewObject(osr.ctx, osr.fs.SubFS(folder), version, id, osr.logger)
}

func (osr *StorageRootBase) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html

	if err := osr.CheckDirectory(); err != nil {
		return errors.WithStack(err)
	} else {
		osr.logger.Infof("StorageRoot with version %s found", osr.version)
	}
	if err := osr.CheckObjects(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (osr *StorageRootBase) CheckDirectory() (err error) {
	// An OCFL Storage Root must contain a Root Conformance Declaration identifying it as such.
	files, err := osr.fs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot get files")
	}
	var version OCFLVersion
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			// check for version file
			if matches := OCFLVersionRegexp.FindStringSubmatch(file.Name()); matches != nil {
				// more than one version file is confusing...
				if version != "" {
					osr.addValidationError(E076, "additional version file \"%s\" in storage root", file.Name())
				} else {
					version = OCFLVersion(matches[1])
				}
			} else {
				// any files are ok -- https://ocfl.io/1.0/spec/#root-structure
			}
		}
	}
	// no version found
	if version == "" {
		osr.addValidationError(E076, "no version file in storage root")
		osr.addValidationError(E077, "no version file in storage root")
	} else {
		osr.version = version
	}
	return nil
}

func (osr *StorageRootBase) CheckObjects() error {
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return errors.Wrapf(err, "cannot get object folders")
	}
	for _, objectFolder := range objectFolders {
		fmt.Printf("object folder '%s'\n", objectFolder)
		objfs := osr.fs.SubFS(objectFolder)
		ctx := NewContextValidation(context.TODO())
		object, err := NewObject(ctx, objfs, "", "", osr.logger)
		if err != nil {
			return errors.Wrap(err, "cannot load object")
		}
		if err := object.Check(); err != nil {
			return errors.Wrapf(err, "check of %s failed", object.GetID())
		}
		showStatus(ctx)
	}
	return nil
}
