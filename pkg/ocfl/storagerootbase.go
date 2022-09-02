package ocfl

import (
	"bytes"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
	"io"
	"io/fs"
)

type StorageRootBase struct {
	fs      OCFLFS
	changed bool
	logger  *logging.Logger
	layout  storageroot.StorageLayout
	version string
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(fs OCFLFS, defaultVersion string, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (*StorageRootBase, error) {
	ocfl := &StorageRootBase{fs: fs, version: defaultVersion, layout: defaultStorageLayout, logger: logger}

	if err := ocfl.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

//var NAMASTERootVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_([0-9]+\\.[0-9]+)")
//var NAMASTEObjectVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_object_([0-9]+\\.[0-9]+)")

var errVersionMultiple = errors.New("multiple version files found")
var errVersionNone = errors.New("no version file found")

func (osr *StorageRootBase) Init() error {
	var err error
	osr.logger.Debug()

	entities, err := osr.fs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot get root entities")
	}
	if len(entities) == 0 {
		rootConformanceDeclaration := "ocfl_" + osr.version
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

func (osr *StorageRootBase) GetObjectFolders() ([]string, error) {
	dirs, err := osr.GetFolders()
	if err != nil {
		return nil, err
	}
	var result = []string{}
	for _, dir := range dirs {
		if dir == "extensions" {
			continue
		}
		result = append(result, dir)
	}
	return result, nil
}

func (osr *StorageRootBase) OpenObjectFolder(folder string) (Object, error) {
	version, err := getVersion(osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		return nil, GetValidationError(osr.version, E003)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s", folder)
	}
	return NewObject(osr.fs, folder, version, "", osr.logger)
}

func (osr *StorageRootBase) OpenObject(id string) (Object, error) {
	folder, err := osr.layout.ExecuteID(id)
	version, err := getVersion(osr.fs, folder, "ocfl_object_")
	if err == errVersionNone {
		return NewObject(osr.fs, folder, osr.version, id, osr.logger)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in %s for [%s]", folder, id)
	}
	return NewObject(osr.fs, folder, version, id, osr.logger)
}

func (osr *StorageRootBase) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return errors.Wrapf(err, "cannot get object folders")
	}
	multiError := emperror.NewMultiErrorBuilder()
	for _, objectFolder := range objectFolders {
		osr.logger.Infof("checking folder %s", objectFolder)
		obj, err := osr.OpenObjectFolder(objectFolder)
		if err != nil {
			return errors.Wrapf(err, "cannot instantiate OCFL ObjectBase at %s", objectFolder)
		}
		if err := obj.Check(); err != nil {
			multiError.Add(err)
		}
	}
	return multiError.ErrOrNil()
}
