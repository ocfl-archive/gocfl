package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/docs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
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
	fsys             fs.FS
	extensionFactory *ExtensionFactory
	extensionManager ExtensionManager
	changed          bool
	logger           zLogger.ZWrapper
	version          OCFLVersion
	digest           checksum.DigestAlgorithm
	modified         bool
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(ctx context.Context, fsys fs.FS, defaultVersion OCFLVersion, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, logger zLogger.ZWrapper) (*StorageRootBase, error) {
	var err error
	ocfl := &StorageRootBase{
		ctx:              ctx,
		fsys:             fsys,
		extensionFactory: extensionFactory,
		version:          defaultVersion,
		extensionManager: extensionManager,
		logger:           logger,
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension manager")
	}
	return ocfl, nil
}

var errVersionMultiple = errors.New("multiple version files found")
var errVersionNone = errors.New("no version file found")
var errInvalidContent = errors.New("content of version declaration does not equal filename")

func (osr *StorageRootBase) String() string {
	return fmt.Sprintf("StorageRootBase: %v", osr.fsys)
}

func (osr *StorageRootBase) IsModified() bool {
	return osr.modified
}
func (osr *StorageRootBase) setModified() {
	osr.modified = true
}

func (osr *StorageRootBase) addValidationError(errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%v' ", osr.fsys)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(addValidationErrors(osr.ctx, valError))
}

func (osr *StorageRootBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%v' ", osr.fsys)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(addValidationWarnings(osr.ctx, valError))
}

func (osr *StorageRootBase) Init(version OCFLVersion, digest checksum.DigestAlgorithm, extensionManager ExtensionManager) error {
	var err error
	osr.logger.Debug()

	osr.version = version
	osr.digest = digest

	entities, err := fs.ReadDir(osr.fsys, ".")
	if err != nil {
		return errors.Wrapf(err, "cannot read storage root directory '%v'", osr.fsys)
	}
	if len(entities) > 0 {
		if err := osr.addValidationError(E069, "storage root not empty"); err != nil {
			return errors.Wrapf(err, "cannot add validation error %v", E069)
		}
		return errors.Wrapf(GetValidationError(version, E069), "storage root %v not empty", osr.fsys)
	}

	rootConformanceDeclaration := "ocfl_" + string(osr.version)
	rootConformanceDeclarationFile := "0=" + rootConformanceDeclaration

	if err := writefs.WriteFile(osr.fsys, rootConformanceDeclarationFile, []byte(rootConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot write %s", rootConformanceDeclarationFile)
	}

	extDocs, err := docs.ExtensionDocs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot read extension docs")
	}
	for _, extDoc := range extDocs {
		if extDoc.IsDir() {
			continue
		}
		extDocFileContent, err := docs.ExtensionDocs.ReadFile(extDoc.Name())
		if err != nil {
			return errors.Wrapf(err, "cannot open extension doc %s", extDoc.Name())
		}
		extDocFile, err := writefs.Create(osr.fsys, extDoc.Name())
		if err != nil {
			return errors.Wrapf(err, "cannot create extension doc %s", extDoc.Name())
		}
		if _, err := extDocFile.Write(extDocFileContent); err != nil {
			return errors.Wrapf(err, "cannot write extension doc %s", extDoc.Name())
		}
		if err := extDocFile.Close(); err != nil {
			return errors.Wrapf(err, "cannot close extension doc %s", extDoc.Name())
		}
	}

	subfs, err := writefs.SubFSCreate(osr.fsys, "extensions")
	if err == nil {
		osr.extensionManager.SetFS(subfs)
		if err := osr.extensionManager.WriteConfig(); err != nil {
			return errors.Wrap(err, "cannot store extension configs")
		}
	}
	if err := osr.extensionManager.StoreRootLayout(osr.fsys); err != nil {
		return errors.Wrap(err, "cannot store ocfl layout")
	}

	return nil
}

func (osr *StorageRootBase) Load() error {
	var err error
	osr.logger.Debug()

	osr.version, err = getVersion(osr.ctx, osr.fsys, ".", "ocfl_")
	if err != nil {
		switch err {
		case errVersionNone:
			if err := osr.addValidationError(E003, "no version declaration file"); err != nil {
				return errors.Wrapf(err, "cannot add validation error %v", E003)
			}
			if err := osr.addValidationError(E004, "no version declaration file"); err != nil {
				return errors.Wrapf(err, "cannot add validation error %v", E004)
			}
			if err := osr.addValidationError(E005, "no version declaration file"); err != nil {
				return errors.Wrapf(err, "cannot add validation error %v", E005)
			}
		case errVersionMultiple:
			if err := osr.addValidationError(E003, "multiple version declaration files"); err != nil {
				return errors.Wrapf(err, "cannot add validation error %v", E003)
			}
		case errInvalidContent:
			if err := osr.addValidationError(E006, "invalid content"); err != nil {
				return errors.Wrapf(err, "cannot add validation error %v", E006)
			}
		default:
			return errors.WithStack(err)
		}
		osr.version = Version1_0
	}

	extFSys, err := fs.Sub(osr.fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for extensions", osr.fsys)
	}
	extensionManager, err := osr.extensionFactory.CreateExtensions(extFSys, osr)
	if err != nil {
		return errors.Wrap(err, "cannot create extension manager")
	}
	osr.extensionManager = extensionManager
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

func (osr *StorageRootBase) CreateExtension(fsys fs.FS) (Extension, error) {
	return osr.extensionFactory.Create(fsys)
}

func (osr *StorageRootBase) CreateExtensions(fsys fs.FS, validation Validation) (ExtensionManager, error) {
	exts, err := osr.extensionFactory.CreateExtensions(fsys, validation)
	return exts, errors.WithStack(err)
}

func (osr *StorageRootBase) StoreExtensionConfig(name string, config any) error {
	extConfig := fmt.Sprintf("extensions/%s/config.json", name)
	cfgJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "cannot marshal extension %s config [%v]", name, config)
	}
	w, err := writefs.Create(osr.fsys, extConfig)
	if err != nil {
		return errors.Wrapf(err, "cannot create file %s", extConfig)
	}
	if _, err := w.Write(cfgJson); err != nil {
		return errors.Wrapf(err, "cannot write file %s - %s", extConfig, string(cfgJson))
	}
	return nil
}

func (osr *StorageRootBase) GetFiles() ([]string, error) {
	dirs, err := fs.ReadDir(osr.fsys, "")
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
	dirs, err := fs.ReadDir(osr.fsys, "")
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
	subFS, err := fs.Sub(osr.fsys, folder)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "cannot create subfs %s of %v", folder, osr.fsys)
	}
	dirs, err := fs.ReadDir(subFS, "")
	if err != nil {
		if err == fs.ErrNotExist {
			return false, nil
		}
		return false, errors.Wrapf(err, "cannot read content of %s", folder)
	}
	return len(dirs) > 0, nil
}

// all folder trees, which end in a folder containing a file
func (osr *StorageRootBase) GetObjectFolders() ([]string, error) {
	var recurse func(base string) ([]string, error)
	recurse = func(base string) ([]string, error) {
		des, err := fs.ReadDir(osr.fsys, base)
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
	version, err := getVersion(osr.ctx, osr.fsys, folder, "ocfl_object_")
	if err == errVersionNone {
		if err := osr.addValidationError(E003, "no version in folder '%s'", folder); err != nil {
			return nil, errors.Wrapf(err, "cannot add validation error %s", E003)
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version in '%s'", folder)
	}
	subfs, err := fs.Sub(osr.fsys, folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of '%v' for '%s'", osr.fsys, folder)
	}
	extFSys, err := fs.Sub(subfs, "extensions")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of '%v' for '%s'", subfs, "extensions")
	}
	extensionManager, err := osr.extensionFactory.CreateExtensions(extFSys, osr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	object, err := newObject(osr.ctx, subfs, version, osr, extensionManager, osr.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot instantiate object")
	}
	// load the object
	if err := object.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from folder '%s'", folder)
	}

	versionFloat, err := strconv.ParseFloat(string(version), 64)
	if err != nil {
		if err := osr.addValidationError(E004, "invalid object version number '%s'", version); err != nil {
			return nil, errors.Wrapf(err, "cannot add validation error %s", E004)
		}
	}
	rootVersionFloat, err := strconv.ParseFloat(string(osr.version), 64)
	if err != nil {
		if err := osr.addValidationError(E075, "invalid root version number '%s'", version); err != nil {
			return nil, errors.Wrapf(err, "cannot add validation error %s", E075)
		}
	}
	// TODO: check. could not find this rule in standard
	if versionFloat > rootVersionFloat {
		if err := osr.addValidationError(E000, "root OCFL version declaration (%s) smaller than highest object version declaration (%s)", osr.version, version); err != nil {
			return nil, errors.Wrapf(err, "cannot add validation error %s", E000)
		}
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

func (osr *StorageRootBase) CreateObject(id string, version OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager ExtensionManager) (Object, error) {
	folder, err := osr.extensionManager.BuildStorageRootPath(osr, id)
	subfs, err := writefs.SubFSCreate(osr.fsys, folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create sub fs of %v for '%s'", osr.fsys, folder)
	}

	object, err := newObject(osr.ctx, subfs, version, osr, manager, osr.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate object")
	}

	// create initial filesystem structure for new object
	if err = object.Init(id, digest, fixity, manager); err != nil {
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
	files, err := fs.ReadDir(osr.fsys, ".")
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
					if err := osr.addValidationError(E076, "additional version file '%s' in storage root", file.Name()); err != nil {
						return errors.Wrapf(err, "cannot add validation error %s", E076)
					}
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
		if err := osr.addValidationError(E076, "no version file in storage root"); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", E076)
		}
		if err := osr.addValidationError(E077, "no version file in storage root"); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", E077)
		}
	} else {
		osr.version = version
	}
	return nil
}
func (osr *StorageRootBase) CheckObjectByFolder(objectFolder string) error {
	fmt.Printf("object folder '%s'\n", objectFolder)
	object, err := osr.LoadObjectByFolder(objectFolder)
	if err != nil {
		if err := osr.addValidationError(E001, "invalid folder '%s': %v", objectFolder, err); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", E001)
		}
		//			return errors.Wrapf(err, "cannot load object from folder '%s'", objectFolder)
	} else {
		if err := object.Check(); err != nil {
			return errors.Wrapf(err, "check of '%s' failed", object.GetID())
		}
	}
	return nil
}

func (osr *StorageRootBase) CheckObjectByID(objectID string) error {
	fmt.Printf("object id '%s'\n", objectID)
	object, err := osr.LoadObjectByID(objectID)
	if err != nil {
		if err := osr.addValidationError(E001, "invalid id '%s': %v", objectID, err); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", E001)
		}
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
		if err := osr.CheckObjectByFolder(objectFolder); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (osr *StorageRootBase) Stat(w io.Writer, path string, id string, statInfo []StatInfo) error {
	if _, err := fmt.Fprintf(w, "Storage Root\n"); err != nil {
		return errors.Wrap(err, "cannot write to writer")
	}
	if _, err := fmt.Fprintf(w, "OCFL Version: %s\n", osr.GetVersion()); err != nil {
		return errors.Wrap(err, "cannot write to writer")
	}
	if slices.Contains(statInfo, StatExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(osr.extensionManager.GetConfig(), "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		if _, err := fmt.Fprintf(w, "Initial Extension:\n---\n%s\n---\n", string(data)); err != nil {
			return errors.Wrap(err, "cannot write to writer")
		}
		if _, err := fmt.Fprintf(w, "Extension Configurations:\n"); err != nil {
			return errors.Wrap(err, "cannot write to writer")
		}
		for _, ext := range osr.extensionManager.GetExtensions() {
			cfg := ext.GetConfig()
			str, _ := json.MarshalIndent(cfg, "", "  ")

			if _, err := fmt.Fprintf(w, "---\n%s\n", str); err != nil {
				return errors.Wrap(err, "cannot write to writer")
			}
		}
	}

	if path == "" && id == "" {
		objectFolders, err := osr.GetObjectFolders()
		if err != nil {
			return errors.Wrap(err, "cannot get object folders")
		}
		if _, err := fmt.Fprintf(w, "Object Folders: %s\n", strings.Join(objectFolders, ", ")); err != nil {
			return errors.Wrap(err, "cannot write to writer")
		}
		data, err := json.MarshalIndent(osr.extensionManager.GetConfig(), "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerconfig")
		}
		if slices.Contains(statInfo, StatExtensionConfigs) || len(statInfo) == 0 {
			if _, err := fmt.Fprintf(w, "Initial Extension:\n---\n%s\n---\n", string(data)); err != nil {
				return errors.Wrap(err, "cannot write to writer")
			}
			if _, err := fmt.Fprintf(w, "Extension Configurations:\n"); err != nil {
				return errors.Wrap(err, "cannot write to writer")
			}
			for _, ext := range osr.extensionManager.GetExtensions() {
				cfg := ext.GetConfig()
				str, _ := json.MarshalIndent(cfg, "", "  ")

				if _, err := fmt.Fprintf(w, "---\n%s\n", str); err != nil {
					return errors.Wrap(err, "cannot write to writer")
				}
			}
		}
		if slices.Contains(statInfo, StatObjects) || len(statInfo) == 0 {
			for _, oFolder := range objectFolders {
				o, err := osr.LoadObjectByFolder(oFolder)
				if err != nil {
					return errors.Wrapf(err, "cannot open object in folder '%s'", oFolder)
				}
				if _, err := fmt.Fprintf(w, "Object: %s\n", oFolder); err != nil {
					return errors.Wrap(err, "cannot write to writer")
				}
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
			if _, err := fmt.Fprintf(w, "cannot load object '%s%s': %v\n", path, id, err); err != nil {
				return errors.Wrap(err, "cannot write to writer")
			}
			return errors.Wrapf(err, "cannot load object '%s%s'", path, id)
		}
		if _, err := fmt.Fprintf(w, "Object: %s%s\n", path, id); err != nil {
			return errors.Wrap(err, "cannot write to writer")
		}
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

func (osr *StorageRootBase) Extract(fsys fs.FS, path, id, version string, withManifest bool, area string) error {
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
			subFS, err := fs.Sub(fsys, oFolder)
			if err != nil {
				return errors.Wrapf(err, "cannot create subfolder '%s' of '%v'", oFolder, fsys)
			}
			if err := o.Extract(subFS, version, withManifest, ""); err != nil {
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
		if err := o.Extract(fsys, version, withManifest, area); err != nil {
			return errors.Wrapf(err, "cannot extract object '%s%s'", path, id)
		}
	}
	osr.logger.Debugf("extraction done")
	return nil
}
