package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type StorageRootBase struct {
	ctx              context.Context
	fsRO             OCFLFSRead
	fsRW             OCFLFS
	extensionFactory *ExtensionFactory
	extensionManager *ExtensionManager
	changed          bool
	logger           *logging.Logger
	version          OCFLVersion
	digest           checksum.DigestAlgorithm
	modified         bool
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(ctx context.Context, fs OCFLFSRead, defaultVersion OCFLVersion, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootBase, error) {
	var err error
	ocfl := &StorageRootBase{
		ctx:              ctx,
		fsRO:             fs,
		extensionFactory: extensionFactory,
		version:          defaultVersion,
		//		digest:           digest,
		logger: logger,
	}
	if rwFS, ok := fs.(OCFLFS); ok {
		ocfl.fsRW = rwFS
	}
	ocfl.extensionManager, err = NewExtensionManager()
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension manager")
	}
	return ocfl, nil
}

var errVersionMultiple = errors.New("multiple version files found")
var errVersionNone = errors.New("no version file found")
var errInvalidContent = errors.New("content of version declaration does not equal filename")

func (osr *StorageRootBase) IsModified() bool {
	return osr.modified
}
func (osr *StorageRootBase) setModified() {
	osr.modified = true
}

func (osr *StorageRootBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%s' ", osr.fsRO)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationErrors(osr.ctx, valError)
}

func (osr *StorageRootBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%s' ", osr.fsRO)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationWarnings(osr.ctx, valError)
}

func (osr *StorageRootBase) Init(version OCFLVersion, digest checksum.DigestAlgorithm, extensions []Extension) error {
	if osr.fsRW == nil {
		return errors.New("filesystem is read only")
	}
	var err error
	osr.logger.Debug()

	osr.version = version
	osr.digest = digest

	entities, err := osr.fsRO.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot read storage root directory")
	}
	if len(entities) > 0 {
		osr.addValidationError(E069, "storage root not empty")
		return errors.Wrapf(GetValidationError(version, E069), "storage root %v not empty", osr.fsRW)
	}

	rootConformanceDeclaration := "ocfl_" + string(osr.version)
	rootConformanceDeclarationFile := "0=" + rootConformanceDeclaration
	rcd, err := osr.fsRW.Create(rootConformanceDeclarationFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", rootConformanceDeclarationFile)
	}
	if _, err := rcd.Write([]byte(rootConformanceDeclaration + "\n")); err != nil {
		rcd.Close()
		return errors.Wrapf(err, "cannot write into '%s'", rootConformanceDeclarationFile)
	}
	if err := rcd.Close(); err != nil {
		return errors.Wrapf(err, "cannot close '%s'", rootConformanceDeclarationFile)
	}

	for _, ext := range extensions {
		if !ext.IsRegistered() {
			osr.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
		}
		if err := osr.extensionManager.Add(ext); err != nil {
			return errors.Wrapf(err, "cannot add extension %s", ext.GetName())
		}
	}
	osr.extensionManager.Finalize()
	subfs, err := osr.fsRW.SubFSRW("extensions")
	if err == nil {
		osr.extensionManager.SetFS(subfs)
		if err := osr.extensionManager.WriteConfig(); err != nil {
			return errors.Wrap(err, "cannot store extension configs")
		}
	}
	if err := osr.extensionManager.StoreRootLayout(osr.fsRW); err != nil {
		return errors.Wrap(err, "cannot store ocfl layout")
	}

	return nil
}

func (osr *StorageRootBase) Load() error {
	var err error
	osr.logger.Debug()

	osr.version, err = getVersion(osr.ctx, osr.fsRO, ".", "ocfl_")
	if err != nil {
		switch err {
		case errVersionNone:
			osr.addValidationError(E003, "no version declaration file")
			osr.addValidationError(E004, "no version declaration file")
			osr.addValidationError(E005, "no version declaration file")
		case errVersionMultiple:
			osr.addValidationError(E003, "multiple version declaration files")
		case errInvalidContent:
			osr.addValidationError(E006, "invalid content")
		default:
			return errors.WithStack(err)
		}
		osr.version = Version1_0
	}
	// read storage layout from extension folder...
	exts, err := osr.fsRO.ReadDir("extensions")
	if err != nil {
		// if directory does not exist - no problem
		if err != fs.ErrNotExist {
			return errors.Wrap(err, "cannot read extensions folder")
		}
		exts = []fs.DirEntry{}
	}
	for _, extFolder := range exts {
		extFolder := fmt.Sprintf("extensions/%s", extFolder.Name())
		subfs, err := osr.fsRO.SubFS(extFolder)
		if err != nil {
			return errors.Wrapf(err, "cannot create subfs of %v for %s", osr.fsRO, extFolder)
		}

		if ext, err := osr.extensionFactory.Create(subfs); err != nil {
			osr.addValidationWarning(W000, "unknown extension in folder '%s'", extFolder)
			//return errors.Wrapf(err, "cannot create extension for config '%s'", extFolder)
		} else {
			if !ext.IsRegistered() {
				osr.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
			}
			if err := osr.extensionManager.Add(ext); err != nil {
				return errors.Wrapf(err, "cannot add extension '%s' to manager", extFolder)
			}
		}
	}
	return nil
}

func (osr *StorageRootBase) GetDigest() checksum.DigestAlgorithm { return osr.digest }

func (osr *StorageRootBase) SetDigest(digest checksum.DigestAlgorithm) {
	if osr.digest == "" {
		osr.digest = digest
	}
}

func (osr *StorageRootBase) GetVersion() OCFLVersion { return osr.version }

func (osr *StorageRootBase) Context() context.Context { return osr.ctx }

func (osr *StorageRootBase) CreateExtension(fs OCFLFSRead) (Extension, error) {
	return osr.extensionFactory.Create(fs)
}

func (osr *StorageRootBase) StoreExtensionConfig(name string, config any) error {
	if osr.fsRW == nil {
		return errors.New("filesystem is read only")
	}
	extConfig := fmt.Sprintf("extensions/%s/config.json", name)
	cfgJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "cannot marshal extension %s config [%v]", name, config)
	}
	w, err := osr.fsRW.Create(extConfig)
	if err != nil {
		return errors.Wrapf(err, "cannot create file %s", extConfig)
	}
	if _, err := w.Write(cfgJson); err != nil {
		return errors.Wrapf(err, "cannot write file %s - %s", extConfig, string(cfgJson))
	}
	return nil
}

func (osr *StorageRootBase) GetFiles() ([]string, error) {
	dirs, err := osr.fsRO.ReadDir("")
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
	dirs, err := osr.fsRO.ReadDir("")
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

//
// Object Functions
//

func (osr *StorageRootBase) ObjectExists(id string) (bool, error) {
	folder, err := osr.extensionManager.BuildStorageRootPath(osr, id)
	if err != nil {
		return false, errors.Wrapf(err, "cannot build storage path for id %s", id)
	}
	subFS, err := osr.fsRO.SubFS(folder)
	if err != nil {
		return false, errors.Wrapf(err, "cannot create subfs %s of %v", folder, osr.fsRO)
	}
	return subFS.HasContent(), nil
}

// all folder trees, which end in a folder containing a file
func (osr *StorageRootBase) GetObjectFolders() ([]string, error) {
	var recurse func(base string) ([]string, error)
	recurse = func(base string) ([]string, error) {
		des, err := osr.fsRO.ReadDir(base)
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

func (osr *StorageRootBase) LoadObjectByFolder(folder string) (Object, error) {
	version, err := getVersion(osr.ctx, osr.fsRO, folder, "ocfl_object_")
	if err == errVersionNone {
		osr.addValidationError(E003, "no version in folder '%s'", folder)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in '%s'", folder)
	}
	subfs, err := osr.fsRO.SubFS(folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of '%v' for '%s'", osr.fsRO, folder)
	}
	object, err := newObject(osr.ctx, subfs, version, osr, osr.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot instantiate object")
	}
	// load the object
	if err := object.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from folder '%s'", folder)
	}

	versionFloat, err := strconv.ParseFloat(string(version), 64)
	if err != nil {
		osr.addValidationError(E004, "invalid object version number '%s'", version)
	}
	rootVersionFloat, err := strconv.ParseFloat(string(osr.version), 64)
	if err != nil {
		osr.addValidationError(E075, "invalid root version number '%s'", version)
	}
	// TODO: check. could not find this rule in standard
	if versionFloat > rootVersionFloat {
		osr.addValidationError(E000, "root OCFL version declaration (%s) smaller than highest object version declaration (%s)", osr.version, version)
	}

	return object, nil
}

func (osr *StorageRootBase) LoadObjectByID(id string) (object Object, err error) {
	folder, err := osr.extensionManager.BuildStorageRootPath(osr, id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create folder from id '%s'", id)
	}
	return osr.LoadObjectByFolder(folder)
}

func (osr *StorageRootBase) CreateObject(id string, version OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, defaultExtensions []Extension) (Object, error) {
	if osr.fsRW == nil {
		return nil, errors.New("filesystem is read only")
	}
	folder, err := osr.extensionManager.BuildStorageRootPath(osr, id)
	subfs, err := osr.fsRW.SubFSRW(folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create sub fs of %v for '%s'", osr.fsRW, folder)
	}

	object, err := newObject(osr.ctx, subfs, version, osr, osr.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate object")
	}

	// create initial filesystem structure for new object
	if err = object.Init(id, digest, fixity, defaultExtensions); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && object.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, object.GetID())
	}

	return object, nil
}

//
// Check functions
//

func (osr *StorageRootBase) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html

	if err := osr.CheckDirectory(); err != nil {
		return errors.WithStack(err)
	} else {
		osr.logger.Infof("StorageRoot with version '%s' found", osr.version)
	}
	if err := osr.CheckObjects(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (osr *StorageRootBase) CheckDirectory() (err error) {
	// An OCFL Storage Root must contain a Root Conformance Declaration identifying it as such.
	files, err := osr.fsRO.ReadDir(".")
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
					osr.addValidationError(E076, "additional version file '%s' in storage root", file.Name())
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
func (osr *StorageRootBase) CheckObject(objectFolder string) error {
	fmt.Printf("object folder '%s'\n", objectFolder)
	object, err := osr.LoadObjectByFolder(objectFolder)
	if err != nil {
		osr.addValidationError(E001, "invalid folder '%s': %v", objectFolder, err)
		//			return errors.Wrapf(err, "cannot load object from folder '%s'", objectFolder)
	} else {
		if err := object.Check(); err != nil {
			return errors.Wrapf(err, "check of '%s' failed", object.GetID())
		}
	}
	return nil
}

func (osr *StorageRootBase) CheckObjects() error {
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return errors.Wrapf(err, "cannot get object folders")
	}
	for _, objectFolder := range objectFolders {
		if err := osr.CheckObject(objectFolder); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (osr *StorageRootBase) Stat(w io.Writer, path string, id string, statInfo []StatInfo) error {
	fmt.Fprintf(w, "Storage Root\n")
	fmt.Fprintf(w, "OCFL Version: %s\n", osr.GetVersion())
	if slices.Contains(statInfo, StatExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(osr.extensionManager.ExtensionManagerConfig, "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		fmt.Fprintf(w, "Initial Extension:\n---\n%s\n---\n", string(data))
		fmt.Fprintf(w, "Extension Configurations:\n")
		for _, ext := range osr.extensionManager.extensions {
			fmt.Fprintf(w, "---\n%s\n", ext.GetConfigString())
		}
	}

	if path == "" && id == "" {
		objectFolders, err := osr.GetObjectFolders()
		if err != nil {
			return errors.Wrap(err, "cannot get object folders")
		}
		fmt.Fprintf(w, "Object Folders: %s\n", strings.Join(objectFolders, ", "))
		data, err := json.MarshalIndent(osr.extensionManager.ExtensionManagerConfig, "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerconfig")
		}
		if slices.Contains(statInfo, StatExtensionConfigs) || len(statInfo) == 0 {
			fmt.Fprintf(w, "Initial Extension:\n---\n%s\n---\n", string(data))
			fmt.Fprintf(w, "Extension Configurations:\n")
			for _, ext := range osr.extensionManager.extensions {
				fmt.Fprintf(w, "---\n%s\n", ext.GetConfigString())
			}
		}
		if slices.Contains(statInfo, StatObjects) || len(statInfo) == 0 {
			for _, oFolder := range objectFolders {
				o, err := osr.LoadObjectByFolder(oFolder)
				if err != nil {
					return errors.Wrapf(err, "cannot open object in folder '%s'", oFolder)
				}
				fmt.Fprintf(w, "Object: %s\n", oFolder)
				if err := o.Stat(w, statInfo); err != nil {
					return errors.Wrapf(err, "cannot show stats for object in folder '%s'", oFolder)
				}
			}
		}
	} else {
		var o Object
		var err error
		if path != "" {
			o, err = osr.LoadObjectByFolder(path)
		} else {
			o, err = osr.LoadObjectByID(id)
		}
		if err != nil {
			fmt.Fprintf(w, "cannot load object '%s%s': %v\n", path, id, err)
			return errors.Wrapf(err, "cannot load object '%s%s'", path, id)
		}
		fmt.Fprintf(w, "Object: %s%s\n", path, id)
		if err := o.Stat(w, statInfo); err != nil {
			return errors.Wrapf(err, "cannot show stats for object '%s%s'", path, id)
		}
	}
	return nil
}

func (osr *StorageRootBase) ExtractMeta(path, id string) (*StorageRootMetadata, error) {
	var result = &StorageRootMetadata{
		Objects: map[string]*ObjectMetadata{},
	}
	if path == "" && id == "" {
		osr.logger.Debug("Extracting storage root with all objects")
		objectFolders, err := osr.GetObjectFolders()
		if err != nil {
			return nil, errors.Wrap(err, "cannot get object folders")
		}
		for _, oFolder := range objectFolders {
			o, err := osr.LoadObjectByFolder(oFolder)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot open object in folder '%s'", oFolder)
			}
			result.Objects[o.GetID()], err = o.GetMetadata()
			if err != nil {
				return nil, errors.Wrapf(err, "cannot extract metadata from object '%s'", o.GetID())
			}
		}
	} else {
		osr.logger.Debugf("Extracting object '%s%s'", path, id)
		var o Object
		var err error
		if path != "" {
			o, err = osr.LoadObjectByFolder(path)
		} else {
			o, err = osr.LoadObjectByID(id)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "cannot load object '%s%s'", path, id)
		}
		result.Objects[o.GetID()], err = o.GetMetadata()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot extract metadata from object '%s'", o.GetID())
		}
	}
	osr.logger.Debugf("extraction done")
	return result, nil
}

func (osr *StorageRootBase) Extract(fs OCFLFS, path, id, version string, withManifest bool) error {
	if version == "" {
		version = "latest"
	}
	if path == "" && id == "" {
		osr.logger.Debugf("Extracting storage root with all objects version '%s'", version)
		objectFolders, err := osr.GetObjectFolders()
		if err != nil {
			return errors.Wrap(err, "cannot get object folders")
		}
		for _, oFolder := range objectFolders {
			o, err := osr.LoadObjectByFolder(oFolder)
			if err != nil {
				return errors.Wrapf(err, "cannot open object in folder '%s'", oFolder)
			}
			subFS, err := fs.SubFSRW(oFolder)
			if err != nil {
				return errors.Wrapf(err, "cannot create subfolder '%s' of '%s'", oFolder, fs.String())
			}
			if err := o.Extract(subFS, version, withManifest); err != nil {
				return errors.Wrapf(err, "cannot extract object in folder '%s'", oFolder)
			}
		}
	} else {
		osr.logger.Debugf("Extracting object '%s%s' with version '%s'", path, id, version)
		var o Object
		var err error
		if path != "" {
			o, err = osr.LoadObjectByFolder(path)
		} else {
			o, err = osr.LoadObjectByID(id)
		}
		if err != nil {
			return errors.Wrapf(err, "cannot load object '%s%s'", path, id)
		}
		if err := o.Extract(fs, version, withManifest); err != nil {
			return errors.Wrapf(err, "cannot extract object '%s%s'", path, id)
		}
	}
	osr.logger.Debugf("extraction done")
	return nil
}
