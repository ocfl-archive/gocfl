package ocfl

import (
	"bytes"
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
	"io"
	"io/fs"
	"path/filepath"
)

type StorageRootBase struct {
	ctx     context.Context
	fs      OCFLFS
	changed bool
	logger  *logging.Logger
	layout  storageroot.StorageLayout
	version OCFLVersion
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (*StorageRootBase, error) {
	ocfl := &StorageRootBase{ctx: ctx, fs: fs, version: defaultVersion, layout: defaultStorageLayout, logger: logger}

	if err := ocfl.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

//var NAMASTERootVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_([0-9]+\\.[0-9]+)")
//var NAMASTEObjectVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_object_([0-9]+\\.[0-9]+)")

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

		configFile := fmt.Sprintf("extensions/%s/config.json", osr.layout.Name())
		extConfig, err := osr.fs.Create(configFile)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", configFile)
		}
		defer extConfig.Close()
		if err := osr.layout.WriteConfig(extConfig); err != nil {
			return errors.Wrap(err, "cannot write config")
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
		var layout storageroot.StorageLayout
		for _, extFolder := range exts {
			extConfig := fmt.Sprintf("extensions/%s/config.json", extFolder.Name())
			configReader, err := osr.fs.Open(extConfig)
			if err != nil {
				return errors.Wrapf(err, "cannot open %s for reading", extConfig)
			}
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, configReader); err != nil {
				return errors.Wrapf(err, "cannot read %s", extConfig)
			}
			if layout, err = storageroot.NewStorageLayout(buf.Bytes()); err != nil {
				osr.logger.Warningf("%s not a storage layout: %v", extConfig, err)
				continue
			}
		}
		if layout == nil {
			// ...or set to default
			if layout, err = storageroot.NewDefaultStorageLayout(); err != nil {
				return errors.Wrap(err, "cannot initiate default storage layout")
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
	version, err := getVersion(osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		osr.addValidationError(E003, "no version in folder %s", folder)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s", folder)
	}
	return NewObject(osr.ctx, osr.fs.SubFS(folder), version, "", osr.logger)
}

func (osr *StorageRootBase) OpenObject(id string) (Object, error) {
	folder, err := osr.layout.ExecuteID(id)
	version, err := getVersion(osr.fs, folder, "ocfl_object_")
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
	var multiErr = []error{}
	var version OCFLVersion
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			// check for version file
			if matches := OCFLVersionRegexp.FindStringSubmatch(file.Name()); matches != nil {
				// more than one version file is confusing...
				if version != "" {
					multiErr = append(multiErr, errors.WithStack(GetValidationError(osr.version, E003)))
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
		multiErr = append(multiErr, errors.WithStack(GetValidationError(osr.version, E004)))
	} else {
		osr.version = version
	}
	return errors.Combine(multiErr...)
}

func (osr *StorageRootBase) CheckObjects() error {
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return errors.Wrapf(err, "cannot get object folders")
	}
	for _, objectFolder := range objectFolders {
		osr.logger.Infof("checking folder %s", objectFolder)
		obj, err := osr.OpenObjectFolder(objectFolder)
		if err != nil {
			return errors.Wrapf(err, "cannot open folder %s", objectFolder)
		}
		if err := obj.Check(); err != nil {
			return errors.Wrapf(err, "folder %s not ok", objectFolder)
		}
	}
	return nil
}
