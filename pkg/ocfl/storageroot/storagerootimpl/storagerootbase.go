package storagerootimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
	"golang.org/x/exp/slices"
)

// NewOCFL creates an empty OCFL structure
func NewStorageRootBase(ctx context.Context, fact factory.Factory, defaultVersion version.OCFLVersion, extensionFactory extension.Factory, logger ocfllogger.OCFLLogger) *StorageRootBase {
	var err error
	ocfl := &StorageRootBase{
		ctx:     ctx,
		factory: fact,
		//fsys:             fsys,
		extensionFactory: extensionFactory,
		version:          defaultVersion,
		//extensionManager: extensionManager,
		logger: logger,
	}
	if err != nil {
		return nil
	}
	return ocfl
}

type StorageRootBase struct {
	ctx              context.Context
	fsys             fs.FS
	extensionFactory extension.Factory
	extensionManager storageroot.ExtensionManager
	logger           ocfllogger.OCFLLogger
	version          version.OCFLVersion
	digest           checksum.DigestAlgorithm
	modified         bool
	factory          factory.Factory
}

func (osr *StorageRootBase) GetExtensionManager() storageroot.ExtensionManager {
	return osr.extensionManager
}

func (osr *StorageRootBase) GetOCFLVersion() version.OCFLVersion {
	return osr.factory.GetVersion()
}

func (osr *StorageRootBase) GetLoader(sourceFS fs.FS, extensionFactor extension.Factory) storageroot.Loader {
	return osr.factory.NewStorageRootLoader(osr.ctx).WithStorageRoot(osr).WithExtensionFactory(extensionFactor).WithFS(sourceFS)
}

func (osr *StorageRootBase) GetInitializer(storageRootFS streamfs.FS) storageroot.Initializer {
	return osr.factory.NewStorageRootInitializer(osr.ctx).WithStorageRoot(osr).WithFS(storageRootFS)
}

//var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

func (osr *StorageRootBase) WithFS(fsys fs.FS) storageroot.StorageRoot {
	osr.fsys = fsys
	return osr
}
func (osr *StorageRootBase) WithExtensionManager(extensionManager storageroot.ExtensionManager) storageroot.StorageRoot {
	osr.extensionManager = extensionManager
	return osr
}

func (osr *StorageRootBase) GetFS() fs.FS {
	return osr.fsys
}
func (osr *StorageRootBase) String() string {
	return fmt.Sprintf("StorageRoot: %v", osr.fsys)
}

func (osr *StorageRootBase) IsModified() bool {
	return osr.modified
}
func (osr *StorageRootBase) SetModified() {
	osr.modified = true
}

func (osr *StorageRootBase) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%v' ", osr.fsys)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationErrors(osr.ctx, valError))
}

func (osr *StorageRootBase) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(osr.version, errno).AppendDescription(format, a...).AppendContext("storage root '%v' ", osr.fsys)
	_, file, line, _ := runtime.Caller(1)
	osr.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationWarnings(osr.ctx, valError))
}

func (osr *StorageRootBase) GetDigest() checksum.DigestAlgorithm { return osr.digest }

func (osr *StorageRootBase) SetDigest(digest checksum.DigestAlgorithm) {
	if osr.digest == "" {
		osr.digest = digest
	}
}

func (osr *StorageRootBase) GetVersion() version.OCFLVersion { return osr.version }

func (osr *StorageRootBase) Context() context.Context { return osr.ctx }

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
	subFS, err := writefs.Sub(osr.fsys, folder)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "cannot create subfs %s of %v", folder, osr.fsys)
	}
	dirs, err := fs.ReadDir(subFS, "/")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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

func (osr *StorageRootBase) IdToFolder(id string) (folder string, err error) {
	folder, err = osr.extensionManager.BuildStorageRootPath(osr, id)
	return folder, errors.WithStack(err)
}

func (osr *StorageRootBase) CreateObject(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, objectExtensionFactory *extensionimpl.Factory, objectExtensionManager object.ExtensionManager) (object.Object, error) {
	folder, err := osr.extensionManager.BuildStorageRootPath(osr, id)
	streamFS, ok := osr.fsys.(streamfs.FS)
	if !ok {
		return nil, errors.New("underlying filesystem does not support writing")
	}
	subfs, err := streamfs.Sub(streamFS, folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create sub fs of %v for '%s'", osr.fsys, folder)
	}

	obj := osr.factory.NewObject(osr.ctx).WithExtensionManager(objectExtensionManager)
	initializer := obj.GetInitializer(subfs)
	if err := initializer.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && obj.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, obj.GetID())
	}

	return obj, nil
}

//
// Check functions
//

func (osr *StorageRootBase) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html

	if err := osr.CheckDirectory(); err != nil {
		return errors.WithStack(err)
	} else {
		osr.logger.Info().Msgf("StorageRoot with version '%s' found", osr.version)
	}
	/*
		if err := osr.CheckObjects(); err != nil {
			return errors.WithStack(err)
		}
	*/

	return nil
}

func (osr *StorageRootBase) CheckDirectory() (err error) {
	// An OCFL Storage Root must contain a Root Conformance Declaration identifying it as such.
	files, err := fs.ReadDir(osr.fsys, ".")
	if err != nil {
		return errors.Wrap(err, "cannot get files")
	}
	var ver version.OCFLVersion
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			// check for version file
			if matches := version.OCFLStorageRootVersionNamasteRegexp.FindStringSubmatch(file.Name()); matches != nil {
				// more than one version file is confusing...
				if ver != "" {
					if err := osr.AddValidationError(validation.E076, "additional version file '%s' in storage root", file.Name()); err != nil {
						return errors.Wrapf(err, "cannot add validation error %s", validation.E076)
					}
				} else {
					ver = version.OCFLVersion(matches[1])
				}
			} else {
				// any files are ok -- https://ocfl.io/1.0/spec/#root-structure
			}
		}
	}
	// no version found
	if ver == "" {
		if err := osr.AddValidationError(validation.E076, "no version file in storage root"); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", validation.E076)
		}
		if err := osr.AddValidationError(validation.E077, "no version file in storage root"); err != nil {
			return errors.Wrapf(err, "cannot add validation error %s", validation.E077)
		}
	} else {
		osr.version = ver
	}
	return nil
}

func (osr *StorageRootBase) Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error {
	if _, err := fmt.Fprintf(w, "Storage Root\n"); err != nil {
		return errors.Wrap(err, "cannot write to writer")
	}
	if _, err := fmt.Fprintf(w, "OCFL Version: %s\n", osr.GetVersion()); err != nil {
		return errors.Wrap(err, "cannot write to writer")
	}
	if slices.Contains(statInfo, object.StatExtensionConfigs) || len(statInfo) == 0 {
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
	return nil
}

var _ storageroot.StorageRoot = (*StorageRootBase)(nil)
