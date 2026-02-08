package objectimpl

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	object2 "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"golang.org/x/exp/slices"
)

//const VERSION = "1.0"

//var objectConformanceDeclaration = fmt.Sprintf("0=ocfl_object_%s", VERSION)

// todo: check WithWriteable() and repair incorrect use...

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(ctx context.Context, factory factorytypes.Factory, defaultVersion version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, extensionManager extensiontypes.ExtensionManagerCore, logger zLogger.ZLogger) *ObjectBase {
	ocfl := &ObjectBase{
		extensionFactory: extensionFactory,
		extensionManager: extensionManager.(object2.ExtensionManager),
		ctx:              ctx,
		fsys:             nil,
		i:                factory.NewInventory(ctx).WithWriteable(),
		//versionFolders:     []string{},
		versionInventories: map[string]inventory.Inventory{},
		changed:            false,
		logger:             logger,
		version:            defaultVersion,
		digest:             "",
		echo:               false,
		updateFiles:        []string{},
		area:               "",
		factory:            factory,
	}
	return ocfl
}

type ObjectBase struct {
	//	storageRoot        storageroot.StorageRoot
	extensionFactory *extensionimpl.ExtensionFactory
	extensionManager object2.ExtensionManager
	ctx              context.Context
	fsys             fs.FS
	i                inventory.Inventory
	//versionFolders     []string
	versionInventories map[string]inventory.Inventory
	changed            bool
	logger             zLogger.ZLogger
	version            version.OCFLVersion
	digest             checksum.DigestAlgorithm
	echo               bool
	updateFiles        []string
	area               string
	factory            factorytypes.Factory
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

func (object *ObjectBase) WithFS(fsys fs.FS) object2.Object {
	object.fsys = fsys
	return object
}

func (object *ObjectBase) GetExtensionManager() object2.ExtensionManager {
	return object.extensionManager
}

func (object *ObjectBase) IsModified() bool { return object.i.IsModified() }

func (object *ObjectBase) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%v' - '%s'", object.fsys, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationErrors(object.ctx, valError))
}

func (object *ObjectBase) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%v' - '%s'", object.fsys, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationWarnings(object.ctx, valError))
}

func (object *ObjectBase) GetMetadata() (*inventory.Metadata, error) {
	inv := object.GetInventory()
	if inv == nil {
		return nil, errors.Errorf("inventory is nil")
	}

	result := &inventory.Metadata{
		ID:              object.GetID(),
		Head:            inv.GetHead(),
		Files:           map[string]*inventory.FileMetadata{},
		DigestAlgorithm: object.GetDigestAlgorithm(),
		Versions:        map[string]*inventory.VersionMetadata{},
	}
	versions := inv.GetVersions()
	versionStrings := []string{}
	for v, ver := range versions.Iterate() {
		result.Versions[v.String()] = &inventory.VersionMetadata{
			Created: ver.GetCreated(),
			Message: ver.GetMessage(),
			Name:    ver.GetUser().GetName(),
			Address: ver.GetUser().GetAddress(),
		}
		versionStrings = append(versionStrings, v.String())
	}
	// sort version strings in ascending order
	slices.SortFunc(versionStrings, func(a, b string) int {
		a = strings.TrimPrefix(a, "v0")
		b = strings.TrimPrefix(b, "v0")
		ia, _ := strconv.Atoi(a)
		ib, _ := strconv.Atoi(b)
		return cmp.Compare(ia, ib)
	})
	extensionMetadata, err := object.extensionManager.GetMetadata(object)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get extension metadata for object '%s'", object.GetID())
	}
	if objectMeta, ok := extensionMetadata[""]; ok {
		/*
			for key, val := range objectMeta {
				result.Extension[key] = val
			}

		*/
		result.Extension = objectMeta
	}
	manifest := inv.GetManifest()
	fixity := inv.GetFixity()
	for digest, fnames := range manifest.Iterate() {
		if len(fnames) == 0 {
			continue
		}
		fm := &inventory.FileMetadata{
			Checksums:    map[checksum.DigestAlgorithm]string{},
			InternalName: fnames,
			VersionName:  map[string][]string{},
			Extension:    map[string]any{},
		}
		fm.Checksums = fixity.Checksums(fnames[0])
		for v, ver := range versions.Iterate() {
			for d, externalNames := range ver.GetState().Iterate() {
				if digest == d {
					if _, ok := fm.VersionName[v.String()]; !ok {
						fm.VersionName[v.String()] = []string{}
					}
					fm.VersionName[v.String()] = append(fm.VersionName[v.String()], externalNames...)
					break
				}
			}
		}
		if emAny, ok := extensionMetadata[digest]; ok {
			if em, ok := emAny.(map[string]any); ok {
				fm.Extension = em
			}
		}
		result.Files[digest] = fm
	}
	return result, nil
}

func (object *ObjectBase) Stat(w io.Writer, statInfo []object2.StatInfo) error {
	fmt.Fprintf(w, "[%s] Path: %s\n", object.GetID(), object.GetDigestAlgorithm())
	i := object.GetInventory()
	fmt.Fprintf(w, "[%s] Head: %s\n", object.GetID(), i.GetHead())
	fixity := i.GetFixity()
	algs := []string{}
	for alg := range fixity.GetDigestAlgorithms() {
		algs = append(algs, string(alg))
	}
	fmt.Fprintf(w, "[%s] Fixity: %s\n", object.GetID(), strings.Join(algs, ", "))
	manifest := i.GetManifest()
	cnt := 0
	for _, fs := range manifest.Iterate() {
		cnt += len(fs)
	}

	var uniqueFileCount int
	for _, _ = range manifest.Iterate() {
		uniqueFileCount++
	}
	fmt.Fprintf(w, "[%s] Manifest: %v files (%v unique files)\n", object.GetID(), cnt, uniqueFileCount)
	if slices.Contains(statInfo, object2.StatObjectVersions) || len(statInfo) == 0 {
		for vString, ver := range i.GetVersions().Iterate() {
			fmt.Fprintf(w, "[%s] Version %s\n", object.GetID(), vString)
			fmt.Fprintf(w, "[%s]     User: %s (%s)\n", object.GetID(), ver.GetUser().GetName(), ver.GetUser().GetAddress())
			fmt.Fprintf(w, "[%s]     Created: %s\n", object.GetID(), ver.GetCreated().String())
			fmt.Fprintf(w, "[%s]     Message: %s\n", object.GetID(), ver.GetMessage())
			if slices.Contains(statInfo, object2.StatObjectVersionState) || len(statInfo) == 0 {
				for cs, sList := range ver.GetState().Iterate() {
					for _, s := range sList {
						fmt.Fprintf(w, "[%s]        %s\n", object.GetID(), s)
						if slices.Contains(statInfo, object2.StatObjectManifest) || len(statInfo) == 0 {
							ms, err := manifest.GetFiles(cs)
							if err != nil {
								if errors.Is(err, inventory.DigestNotFound) {
									continue
								}
								return errors.Wrapf(err, "cannot get files for manifest '%s'", object.GetID())
							}
							for _, m := range ms {
								fmt.Fprintf(w, "[%s]           %s\n", object.GetID(), m)
							}
						}
					}
				}
			}
		}
	}
	if slices.Contains(statInfo, object2.StatObjectExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(object.extensionManager.GetConfig(), "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		fmt.Fprintf(w, "[%s] Initial Extension:\n---\n%s\n---\n", object.GetID(), string(data))
		fmt.Fprintf(w, "[%s] Extension Configurations:\n", object.GetID())
		for _, ext := range object.extensionManager.GetExtensions() {
			cfg := ext.GetConfig()
			str, _ := json.MarshalIndent(cfg, "", "  ")

			fmt.Fprintf(w, "---\n%s\n", str)
		}
	}
	return nil
}

func (object *ObjectBase) GetFS() fs.FS {
	return object.fsys
}

func (object *ObjectBase) CreateInventory(id string, digestAlg checksum.DigestAlgorithm, fixityAlgs []checksum.DigestAlgorithm) (inventory.Inventory, error) {
	fixity := object.factory.NewFixity(object.ctx).WithAlgorithms(fixityAlgs...)
	inventory := object.factory.NewInventory(object.ctx).
		WithID(id).
		WithDigestAlgorithm(digestAlg).
		WithFixity(fixity)

	/*
		inventory, err := inventory.NewInventory(object.ctx, "new", object.GetOCFLVersion(), object.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
	/*
		if err := inventory.Init(id, digest, fixity); err != nil {
			return nil, errors.Wrap(err, "cannot initialize empty inventory")
		}
	*/

	return inventory, inventory.Finalize(true)
}
func (object *ObjectBase) GetInventory() inventory.Inventory {
	return object.i
}

func (object *ObjectBase) loadInventory(data []byte, folder string) (inventory.Inventory, error) {
	anyMap := map[string]any{}
	if err := json.Unmarshal(data, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(data))
	}
	var ver version.OCFLVersion
	t, ok := anyMap["type"]
	if !ok {
		return nil, errors.New("no type in inventory")
	}
	sStr, ok := t.(string)
	if !ok {
		return nil, errors.Errorf("type not a string in inventory - '%v'", t)
	}
	switch sStr {
	case "https://ocfl.io/1.1/spec/#inventory":
		ver = version.Version1_1
	case "https://ocfl.io/1.0/spec/#inventory":
		ver = version.Version1_0
	case "https://ocfl.io/2.0/spec/#inventory":
		ver = version.Version2_0
	default:
		// if we don't know anything use the old stuff
		ver = version.Version1_0
	}
	inventory := object.factory.NewInventory(object.ctx).WithWriteable()
	/*
		inventory, err := inventory.NewInventory(object.ctx, folder, ver, object.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
	if err := json.Unmarshal(data, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			validation.AddValidationErrors(object.ctx, validation.GetValidationError(ver, validation.E033).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", object.fsys))
			validation.AddValidationErrors(object.ctx, validation.GetValidationError(ver, validation.E034).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", object.fsys))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				validation.AddValidationErrors(object.ctx, validation.GetValidationError(ver, validation.E040).AppendDescription("head is not of string type: %v", jsonMap["head"]).AppendContext("object '%v'", object.fsys))
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inventory, inventory.Finalize(false)
}

var inventorySideCarFormat = regexp.MustCompile(`^([a-fA-F0-9]+)\s+inventory.json$`)

// loadInventory loads inventory from existing Object
func (object *ObjectBase) LoadInventory(folder string) (inventory.Inventory, error) {
	// load inventory file
	filename := filepath.ToSlash(filepath.Join(folder, "inventory.json"))
	inventoryBytes, err := fs.ReadFile(object.fsys, filename)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			return nil, err
		}
		inventory := object.factory.NewInventory(object.ctx).WithWriteable()
		return inventory, nil
	}
	inventory, err := object.loadInventory(inventoryBytes, folder)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := inventory.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(object.fsys, sidecarPath)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			object.AddValidationError(validation.E058, "sidecar '%v/%s' does not exist", object.fsys, sidecarPath)
		} else {
			object.AddValidationError(validation.E060, "cannot read sidecar '%v/%s': %v", object.fsys, sidecarPath, err.Error())
		}
		//		object.AddValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		//if !strings.HasSuffix(digestString, " inventory.json") {
		matches := inventorySideCarFormat.FindStringSubmatch(digestString)
		if /* matches == nil || */ len(matches) == 0 {
			object.AddValidationError(validation.E061, "no suffix \" inventory.json\" in '%v/%s'", object.fsys, sidecarPath)
		} else {
			//digestString = strings.TrimSpace(strings.TrimSuffix(digestString, " inventory.json"))
			digestString = matches[1]
			h, err := checksum.GetHash(digest)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - '%s'", string(digest)))
			}
			h.Reset()
			h.Write(inventoryBytes)
			sumBytes := h.Sum(nil)
			inventoryDigestString := fmt.Sprintf("%x", sumBytes)
			if digestString != inventoryDigestString {
				object.AddValidationError(validation.E060, "'%s' != '%s'", digestString, inventoryDigestString)
			}
		}
	}
	return inventory, inventory.Finalize(false)
}

func (object *ObjectBase) StoreInventory(version bool, objectRoot bool) error {
	if object.fsys == nil {
		return errors.Errorf("read only filesystem '%v'", object.fsys)
	}
	object.logger.Debug()

	// check whether object filesystem is writeable
	if !object.i.IsWriteable() {
		return errors.New("inventory not writeable - not updated")
	}

	// create inventory.json from inventory
	iFileName := "inventory.json"
	jsonBytes, err := json.MarshalIndent(object.i, "", "   ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(object.i.GetDigestAlgorithm())
	if err != nil {
		return errors.Wrapf(err, "invalid digest algorithm '%s'", string(object.i.GetDigestAlgorithm()))
	}
	if _, err := h.Write(jsonBytes); err != nil {
		return errors.Wrapf(err, "cannot create checksum of manifest")
	}
	checksumBytes := h.Sum(nil)
	checksumString := fmt.Sprintf("%x %s", checksumBytes, iFileName)

	if objectRoot {
		iWriter, err := writefs.Create(object.fsys, iFileName)
		if err != nil {
			iWriter.Close()
			return errors.Wrap(err, "cannot create inventory.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			return errors.Wrap(err, "cannot write to inventory.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", object.fsys, iFileName)
		}
		csFileName := fmt.Sprintf("inventory.json.%s", string(object.i.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(object.fsys, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", object.fsys, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			iCSWriter.Close()
			return errors.Wrapf(err, "cannot write to '%v/%s'", object.fsys, csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", object.fsys, csFileName)
		}
	}
	if version {
		iFileName = fmt.Sprintf("%s/inventory.json", object.i.GetHead())
		iWriter, err := writefs.Create(object.fsys, iFileName)
		if err != nil {
			return errors.Wrap(err, "cannot create inventory.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			iWriter.Close()
			return errors.Wrap(err, "cannot write to inventory.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", object.fsys, iFileName)
		}
		csFileName := fmt.Sprintf("%s/inventory.json.%s", object.i.GetHead(), string(object.i.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(object.fsys, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", object.fsys, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			iCSWriter.Close()
			return errors.Wrapf(err, "cannot write to '%s'", csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", object.fsys, csFileName)
		}
	}
	return nil
}

func (object *ObjectBase) StoreExtensions() error {
	object.logger.Debug()

	if err := object.extensionManager.WriteConfig(); err != nil {
		return errors.Wrap(err, "cannot store extension configs")
	}
	return nil
}

func (object *ObjectBase) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensionManager extensiontypes.ExtensionManagerCore) error {
	object.logger.Debug().Msgf("%s", id)
	objectConformanceDeclaration := "ocfl_object_" + string(object.version)
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration

	object.extensionManager = extensionManager.(object2.ExtensionManager)

	// first check whether object is not empty
	fp, err := object.fsys.Open(objectConformanceDeclarationFile)
	if err == nil {
		// not empty, close it and return error
		if err := fp.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%s'", objectConformanceDeclarationFile)
		}
		return fmt.Errorf("cannot create object '%s'. '%v/%s' already exists", id, object.fsys, objectConformanceDeclarationFile)
	}
	cnt, err := fs.ReadDir(object.fsys, ".")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errors.Wrapf(err, "cannot read '%v/%s'", object.fsys, ".")
	}
	if len(cnt) > 0 {
		return fmt.Errorf("'%v/%s' is not empty", ".", object.fsys)
	}
	if _, err := writefs.WriteFile(object.fsys, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", object.fsys, objectConformanceDeclarationFile)
	}

	subfs, err := writefs.SubFSCreate(object.fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	object.extensionManager.SetFS(subfs, true)

	// check fixity here
	algs := []checksum.DigestAlgorithm{
		checksum.DigestSHA512,
		checksum.DigestSHA256,
	}
	algs = append(algs, object.extensionManager.GetFixityDigests()...)
	slices.Sort(algs)
	algs = slices.Compact(algs)
	if !util.SliceContains(algs, fixity) {
		return errors.Errorf("forbidden digest algorithm for fixity %v. Supported algorithms are %v. (to fix try to use extension 0001-digest-algorithms)", fixity, algs)
	}

	object.i, err = object.CreateInventory(id, digest, fixity)
	object.i.WithWriteable()
	return nil
}

func (object *ObjectBase) Load() (err error) {
	extFolder, err := writefs.Sub(object.fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	manager, err := object.extensionFactory.CreateExtensions(extFolder, object)
	if err != nil {
		object.AddValidationWarning(validation.W000, "cannot initialize all extensions in folder '%s': %v", extFolder, err)
		if manager == nil {
			return errors.Wrap(err, "cannot create extension manager")
		}
	}
	object.extensionManager = manager.(object2.ExtensionManager)

	// load the inventory
	if object.i, err = object.LoadInventory("."); err != nil {
		return errors.Wrap(err, "cannot load inventory.json of root")
	}
	return nil
}

func (object *ObjectBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return object.i.GetDigestAlgorithm()
}

func (object *ObjectBase) echoDelete() error {
	slices.Sort(object.updateFiles)
	object.updateFiles = slices.Compact(object.updateFiles)
	basePath, err := object.extensionManager.BuildObjectStatePath(object, ".", "")
	if err != nil {
		return errors.Wrap(err, "cannot build external path for '.'")
	}
	if basePath == "." {
		basePath = ""
	}
	version := object.i.GetVersions().GetVersion(object.i.GetHead())
	if _, err := version.EchoDelete(object.updateFiles, basePath); err != nil {
		return errors.Wrap(err, "cannot remove deleted files from inventory")
	}
	return nil
}

func (object *ObjectBase) Close() error {
	object.logger.Info().Msgf(fmt.Sprintf("Closing object '%s'", object.GetID()))
	if !(object.i.IsWriteable()) {
		return nil
	}

	if !object.i.IsModified() {
		return nil
	}
	//object.storageRoot.setModified()
	if err := object.i.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := object.StoreInventory(false, true); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}
	if err := object.StoreExtensions(); err != nil {
		return errors.Wrap(err, "cannot store extensions")
	}
	return nil
}

func (object *ObjectBase) StartUpdate(sourceFS fs.FS, msg string, UserName string, UserAddress string, echo bool) (fs.FS, error) {
	object.logger.Debug().Msgf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)
	object.echo = echo

	subfs, err := writefs.SubFSCreate(object.fsys, "extensions")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	// todo: bad style, changes extensionManager in factory
	object.extensionManager.SetFS(subfs, true)

	if err := object.i.GetVersions().NewVersion(object.i.GetHead(), msg, UserName, UserAddress); err != nil {
		return nil, errors.Wrap(err, "cannot create new object version")
	}
	if err := object.extensionManager.UpdateObjectBefore(object); err != nil {
		return nil, errors.Wrapf(err, "cannot execute ext.UpdateObjectBefore()")
	}
	var versionFS fs.FS
	return versionFS, nil
}

func (object *ObjectBase) EndUpdate() error {
	object.logger.Info().Msgf(fmt.Sprintf("EndUpdate of object '%s'", object.GetID()))
	if !(object.i.IsWriteable()) {
		object.logger.Warn().Msgf(fmt.Sprintf("object '%s' not writeable", object.GetID()))
		return nil
	}
	if !(object.i.IsModified()) {
		object.logger.Info().Msgf(fmt.Sprintf("object '%s' not modified", object.GetID()))
		return nil
	}

	if object.echo {
		if err := object.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot delete files")
		}
	}
	if err := object.extensionManager.UpdateObjectAfter(object); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
	}

	if err := object.i.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := object.StoreInventory(true, false); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}

	if needVersion, err := object.extensionManager.NeedNewVersion(object); err != nil {
		return errors.Wrapf(err, "cannot execute ext.NeedNewVersion()")
	} else if needVersion {
		if _, err := object.StartUpdate(nil, "automated version", "gocfl", "https://github.com/ocfl-archive/gocfl", false); err != nil {
			return errors.Wrap(err, "cannot create new version")
		}
		if err := object.extensionManager.DoNewVersion(object); err != nil {
			return errors.Wrapf(err, "cannot execute ext.DoNewVersion()")
		}
		/*
			if err := object.extensionManager.UpdateObjectAfter(object); err != nil {
				return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
			}
		*/
		if err := object.EndUpdate(); err != nil {
			return errors.Wrap(err, "cannot end update")
		}
	}
	return nil
}

func (object *ObjectBase) BeginArea(area string) {
	object.area = area
	object.updateFiles = []string{}
}

func (object *ObjectBase) EndArea() error {
	if object.echo {
		if err := object.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot remove files")
		}
	}
	object.updateFiles = []string{}
	object.area = ""
	return nil
}

func (object *ObjectBase) AddFolder(fsys fs.FS, versionFS fs.FS, checkDuplicate bool, area string) error {
	object.logger.Debug().Msgf("walking '%v'", fsys)
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		if info.Name() == "." {
			return nil
		}
		path = filepath.ToSlash(path)
		if err := object.AddFile(fsys, versionFS, path, checkDuplicate, area, false, info.IsDir()); err != nil {
			return errors.Wrapf(err, "cannot add file '%s'", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (object *ObjectBase) addReader(r io.ReadCloser, versionFS fs.FS, names *object2.NamesStruct, noExtensionHook bool) (string, error) {

	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range object.i.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}

	var digest string

	object.updateFiles = append(object.updateFiles, names.ExternalPaths...)

	if !slices.Contains(digestAlgorithms, object.i.GetDigestAlgorithm()) {
		digestAlgorithms = append(digestAlgorithms, object.i.GetDigestAlgorithm())
	}

	writer, err := writefs.Create(object.fsys, names.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "cannot create '%s'", names.ManifestPath)
	}
	defer writer.Close()

	var checksums map[checksum.DigestAlgorithm]string
	if noExtensionHook {
		checksums, err = checksum.Copy(digestAlgorithms, r, writer)
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%v' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		pr, pw := io.Pipe()
		extErrors := make(chan error, 1)
		go func() {
			defer wg.Done()
			if err := object.extensionManager.StreamObject(object, pr, names.ExternalPaths, names.InternalPath); err != nil {
				extErrors <- err
			}
		}()
		checksums, err = checksum.Copy(digestAlgorithms, r, writer, pw)
		if err := pw.Close(); err != nil {
			object.logger.Error().Err(err).Msg("cannot close pipe writer")
		}
		wg.Wait()
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%s' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
		close(extErrors)
		select {
		case err, ok := <-extErrors:
			if ok {
				return "", errors.Wrapf(err, "error on StreamObject() extension hook for object '%s'", object.GetID())
			}
		default:
		}
	}

	if digest == "" {
		var ok bool
		digest, ok = checksums[object.i.GetDigestAlgorithm()]
		if !ok {
			return "", errors.Errorf("digest '%s' not generated", object.i.GetDigestAlgorithm())
		}
	} else {
		checksums[object.i.GetDigestAlgorithm()] = digest
	}
	if err := object.i.AddFile(names.ExternalPaths, names.ManifestPath, checksums); err != nil {
		return "", errors.Wrapf(err, "cannot append '%v'/'%s' to inventory", names.ExternalPaths, names.InternalPath)
	}

	return digest, nil
}

func (object *ObjectBase) BuildNames(files []string, area string) (*object2.NamesStruct, error) {
	var err error
	result := &object2.NamesStruct{
		ExternalPaths: []string{},
	}
	for _, file := range files {
		externalPath, err := object.extensionManager.BuildObjectStatePath(object, file, area)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", file)
		}
		result.ExternalPaths = append(result.ExternalPaths, externalPath)
	}
	result.InternalPath, err = object.extensionManager.BuildObjectManifestPath(object, files[0], area)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create manifest path for '%s'", files[0])
	}
	result.ManifestPath = object.i.BuildManifestName(result.InternalPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", result.InternalPath)
	}
	return result, nil
}

func (object *ObjectBase) AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error) {
	if len(files) == 0 {
		return "", errors.New("no files given")
	}
	if !object.i.IsWriteable() {
		return "", errors.New("object not writeable")
	}
	path := files[0]
	names, err := object.BuildNames(files, area)

	object.logger.Info().Msgf("adding file %s:%v", area, files)

	if !noExtensionHook {
		if err := object.extensionManager.AddFileBefore(object, nil, path, names.InternalPath, area, false); err != nil {
			return "", errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var digest string
	if !isDir {
		digest, err = object.addReader(r, nil, names, noExtensionHook)
		if err != nil {
			return "", errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	} else {
		io.Copy(io.Discard, r)
	}

	if !noExtensionHook {
		if err := object.extensionManager.AddFileAfter(object, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return "", errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return digest, nil
}

func (object *ObjectBase) AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	ver := object.i.GetVersions().GetVersion(object.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", object.i.GetHead())
	}

	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range object.i.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}
	var digest string

	names, err := object.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot names for '%s'", path)

	}

	object.logger.Info().Msgf("adding file %s:%s", area, path)

	newPath, err := object.extensionManager.BuildObjectStatePath(object, path, area)
	if err != nil {
		return errors.Wrapf(err, "cannot map external path '%s'", path)
	}

	object.updateFiles = append(object.updateFiles, newPath)

	var dataReader = bytes.NewReader(data)
	if checkDuplicate {
		// do the checksum
		digest, err = checksum.Checksum(dataReader, object.i.GetDigestAlgorithm())
		if err != nil {
			return errors.Wrapf(err, "cannot create digest of '%s'", path)
		}
		// set filepointer to beginning
		if _, err := dataReader.Seek(0, 0); err != nil {
			return errors.Wrapf(err, "cannot seek in datareader")
		}
		// if file is already there we do nothing
		dup, err := object.i.AlreadyExists(newPath, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
		}
		if dup {
			object.logger.Info().Msgf("[%s] '%s' already exists. ignoring", object.GetID(), newPath)
			return nil
		}
		// file already ingested, but new virtual name
		if dups, _ := object.i.GetManifest().GetFiles(digest); len(dups) > 0 {
			object.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", object.GetID(), newPath)
			if _, err := ver.CopyFile(newPath, digest); err != nil {
				return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
			}
			return nil
		}
	} else {
		if !slices.Contains(digestAlgorithms, object.i.GetDigestAlgorithm()) {
			digestAlgorithms = append(digestAlgorithms, object.i.GetDigestAlgorithm())
		}
	}

	if !noExtensionHook {
		if err := object.extensionManager.AddFileBefore(object, nil, path, names.InternalPath, area, false); err != nil {
			return errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var r = io.NopCloser(dataReader)
	if !isDir {
		digest, err = object.addReader(r, nil, names, noExtensionHook)
		if err != nil {
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	}

	if !noExtensionHook {
		if err := object.extensionManager.AddFileAfter(object, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (object *ObjectBase) AddFile(fsys fs.FS, versionFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	object.logger.Info().Msgf("adding file %s:%s", area, path)

	ver := object.i.GetVersions().GetVersion(object.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", object.i.GetHead())
	}

	path = filepath.ToSlash(path)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	names, err := object.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
	}

	targetFilename := object.i.BuildManifestName(names.InternalPath)

	var digest string
	if !isDir {

		digestAlgorithms := []checksum.DigestAlgorithm{}
		for alg := range object.i.GetFixity().GetDigestAlgorithms() {
			digestAlgorithms = append(digestAlgorithms, alg)
		}

		file, err := fsys.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot open file '%v/%s'", fsys, path)
		}
		newPath, err := object.extensionManager.BuildObjectStatePath(object, path, area)
		if err != nil {
			file.Close()
			return errors.Wrapf(err, "cannot map external path '%s'", path)
		}

		object.updateFiles = append(object.updateFiles, newPath)

		if checkDuplicate {
			// do the checksum
			digest, err = checksum.Checksum(file, object.i.GetDigestAlgorithm())
			if err != nil {
				return errors.Wrapf(err, "cannot create digest of '%s'", path)
			}
			// set filepointer to beginning
			if seeker, ok := file.(io.Seeker); ok {
				// if we have a seeker, we just seek
				if _, err := seeker.Seek(0, 0); err != nil {
					panic(err)
				}
			} else {
				// otherwise reopen it
				file, err = fsys.Open(path)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s'", fsys, path)
				}
			}
			// if file is already there we do nothing
			dup, err := object.i.AlreadyExists(newPath, digest)
			if err != nil {
				return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
			}
			if dup {
				object.logger.Info().Msgf("[%s] '%s' already exists. ignoring", object.GetID(), newPath)
				return nil
			}
			// file already ingested, but new virtual name
			if dups, _ := object.i.GetManifest().GetFiles(digest); len(dups) > 0 {
				object.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", object.GetID(), newPath)
				if _, err := ver.CopyFile(newPath, digest); err != nil {
					return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
				}
				return nil
			}
		} else {
			if !slices.Contains(digestAlgorithms, object.i.GetDigestAlgorithm()) {
				digestAlgorithms = append(digestAlgorithms, object.i.GetDigestAlgorithm())
			}
		}
		if !noExtensionHook {
			if err := object.extensionManager.AddFileBefore(object, nil, path, names.InternalPath, area, isDir); err != nil {
				return errors.Wrapf(err, "error on AddFileBefore() extension hook")
			}
		}

		digest, err = object.addReader(file, versionFS, names, noExtensionHook)
		if err != nil {
			file.Close()
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
		if err := file.Close(); err != nil {
			return errors.Wrapf(err, "cannot close file '%s'", path)
		}

	}
	if !noExtensionHook {
		if err := object.extensionManager.AddFileAfter(object, fsys, []string{path}, targetFilename, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (object *ObjectBase) DeleteFile(virtualFilename string, digest string) error {
	ver := object.i.GetVersions().GetVersion(object.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", object.i.GetHead())
	}
	virtualFilename = filepath.ToSlash(virtualFilename)
	object.logger.Debug().Msgf("removing '%s' [%s]", virtualFilename, digest)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := object.i.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilename, digest)
	}
	if !dup {
		object.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilename, digest)
		return nil
	}
	if _, err := ver.DeleteFile(virtualFilename); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilename)
	}
	return nil

}

func (object *ObjectBase) RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error {
	ver := object.i.GetVersions().GetVersion(object.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", object.i.GetHead())
	}
	virtualFilenameSource = filepath.ToSlash(virtualFilenameSource)
	object.logger.Debug().Msgf("removing '%s' [%s]", virtualFilenameSource, digest)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := object.i.AlreadyExists(virtualFilenameSource, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilenameSource, digest)
	}
	if !dup {
		object.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilenameSource, digest)
		return nil
	}
	if _, err := ver.RenameFile(virtualFilenameSource, virtualFilenameDest); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilenameSource)
	}
	return nil

}

func (object *ObjectBase) GetID() string {
	if object.i == nil {
		return ""
	}
	return object.i.GetID()
}

func (object *ObjectBase) GetOCFLVersion() version.OCFLVersion {
	return object.version
}

var allowedFilesRegexp = regexp.MustCompile("^(inventory.json(\\.sha512|\\.sha384|\\.sha256|\\.sha1|\\.md5)?|0=ocfl_object_[0-9]+\\.[0-9]+)$")

func (object *ObjectBase) checkVersionFolder(version string) error {
	versionEntries, err := fs.ReadDir(object.fsys, version)
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder '%s'", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !allowedFilesRegexp.MatchString(ve.Name()) {
				object.AddValidationError(validation.E015, "extra file '%s' in version directory '%s'", ve.Name(), version)
			}
		}
	}
	return nil
}

func (object *ObjectBase) checkFilesAndVersions() error {
	// create list of version content directories
	versionContents := map[string]string{}
	versionStrings := ocfl.SeqToSlice(object.i.GetVersions().GetVersionNumbers())

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b *inventory.VersionNumber) int {
		if a.Less(b) {
			return -1
		}

		if a.Equal(b) {
			return 0
		}

		return 1
	})

	for _, ver := range versionStrings {
		versionContents[ver.String()] = object.i.GetContentDir()
	}

	// load object content files
	objectContentFiles := map[string][]string{}
	objectContentFilesFlat := []string{}
	objectFilesFlat := []string{}
	for ver, cont := range versionContents {
		// load all object version content files
		versionContent := ver + "/" + cont
		//inventoryFile := ver + "/inventory.json"
		if _, ok := objectContentFiles[ver]; !ok {
			objectContentFiles[ver] = []string{}
		}
		fs.WalkDir(
			object.fsys,
			ver,
			func(path string, d fs.DirEntry, err error) error {
				path = filepath.ToSlash(path)
				if d.IsDir() {
					if !strings.HasPrefix(path, versionContent) && path != ver && !strings.HasPrefix(ver+"/"+object.i.GetContentDir(), path) {
						object.AddValidationWarning(validation.W002, "extra dir '%s' in version '%s'", path, ver)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[ver] = append(objectContentFiles[ver], path)
						objectContentFilesFlat = append(objectContentFilesFlat, path)
					} else {
						/*
							if !strings.HasPrefix(path, inventoryFile) {
								object.AddValidationWarning(W002, "extra file '%s' in version '%s'", path, ver)
							}
						*/
					}
				}
				return nil
			},
		)
		if len(objectContentFiles[ver]) == 0 {
			fi, err := fs.Stat(object.fsys, versionContent)
			if err != nil {
				if !errors.Is(errors.Cause(err), fs.ErrNotExist) {
					return errors.Wrapf(err, "cannot stat '%s'", versionContent)
				}
			} else {
				if fi.IsDir() {
					object.AddValidationWarning(validation.W003, "empty content folder '%s'", versionContent)
				}
			}
		}
	}
	// load all inventories
	versionInventories, err := object.getVersionInventories()
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}

	csDigestFiles, err := object.createContentManifest()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := object.i.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrap(err, "cannot check file digests for object root")
	}

	contentDir := ""
	if len(versionStrings) > 0 {
		contentDir = versionInventories[versionStrings[0].String()].GetRealContentDir()
	}
	for _, ver := range versionStrings {
		inv := versionInventories[ver.String()]
		if inv == nil {
			continue
		}
		if contentDir != inv.GetRealContentDir() {
			object.AddValidationError(validation.E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", inv.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := inv.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", ver)
		}
		digestAlg := inv.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{inv.GetContentDir()}
		versionEntries, err := fs.ReadDir(object.fsys, ver.String())
		if err != nil {
			object.AddValidationError(validation.E010, "cannot read version folder '%s'", ver)
			continue
			//			return errors.Wrapf(err, "cannot read dir '%s'", ver)
		}
		for _, entry := range versionEntries {
			if entry.IsDir() {
				if !slices.Contains(allowedDirs, entry.Name()) {
					object.AddValidationWarning(validation.W002, "extra dir '%s' in version directory '%s'", entry.Name(), ver)
				}
			} else {
				if !slices.Contains(allowedFiles, entry.Name()) {
					object.AddValidationError(validation.E015, "extra file '%s' in version directory '%s'", entry.Name(), ver)
				}
			}
		}
	}

	for key := 0; key < len(versionStrings)-1; key++ {
		v1 := versionStrings[key]
		vi1, ok := versionInventories[v1.String()]
		if !ok {
			object.AddValidationWarning(validation.W010, "no inventory for version '%s'", versionStrings[key])
			continue
			// return errors.Errorf("no inventory for version '%s'", versionStrings[key])
		}
		v2 := versionStrings[key+1]
		vi2, ok := versionInventories[v2.String()]
		if !ok {
			object.AddValidationWarning(validation.W000, "no inventory for version '%s'", versionStrings[key+1])
			continue
		}
		if !inventory.SpecIsLessOrEqual(vi1.GetSpec(), vi2.GetSpec()) {
			object.AddValidationError(validation.E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", v1, vi1.GetSpec(), v2, vi2.GetSpec())
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion.String()]; ok {
			if !lastInv.Equals(object.i) {
				object.AddValidationError(validation.E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}

	id := object.i.GetID()
	digestAlg := object.i.GetDigestAlgorithm()
	versions := object.i.GetVersions()
	for ver, verInventory := range versionInventories {
		// check for id consistency
		if id != verInventory.GetID() {
			object.AddValidationError(validation.E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", id, ver, verInventory.GetID())
		}
		if verInventory.GetHead().IsValid() && verInventory.GetHead().String() != ver {
			object.AddValidationError(validation.E040, "wrong head '%s' in manifest for version '%s'", verInventory.GetHead(), ver)
		}

		if verInventory.GetDigestAlgorithm() != digestAlg {
			object.AddValidationError(validation.W000, "different digest algorithm '%s' in version '%s'", verInventory.GetDigestAlgorithm(), ver)
		}

		for versionNumber, vVersion := range verInventory.GetVersions().Iterate() {
			testV := versions.GetVersion(versionNumber)
			if testV == nil {
				object.AddValidationError(validation.E066, "version '%s' in version folder '%s' not in object root manifest", vVersion, versionNumber)
			}
			if !testV.Equals(vVersion) {
				object.AddValidationError(validation.E066, "version '%s' in version folder '%s' not equal to version in object root manifest", vVersion, versionNumber)
			}
		}
	}

	//
	// all files in any manifest must belong to a physical file #E092
	//
	for inventoryVersion, inventory := range versionInventories {
		for manifestFile := range inventory.GetManifest().GetFilesFlat() {
			if !slices.Contains(objectFilesFlat, manifestFile) {
				object.AddValidationError(validation.E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}

	for manifestFile := range object.i.GetManifest().GetFilesFlat() {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			object.AddValidationError(validation.E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
		}
	}

	//
	// all object content files must belong to manifest
	//

	latestVersion := inventory.NewVersionNumber()

	for objectContentVersion, objectContentVersionFiles := range objectContentFiles {
		objectContentVersionNumber := inventory.NewVersionNumber().WithString(objectContentVersion)
		if !latestVersion.IsValid() {
			latestVersion = objectContentVersionNumber
		}
		if latestVersion.Less(objectContentVersionNumber) {
			latestVersion = objectContentVersionNumber
		}
		// check version inventories
		for inventoryVersion, versionInventory := range versionInventories {
			inventoryVersionNumber := inventory.NewVersionNumber().WithString(inventoryVersion)
			if objectContentVersionNumber.Less(inventoryVersionNumber) {
				versionManifestFiles := ocfl.SeqToSlice(versionInventory.GetManifest().GetFilesFlat())
				for _, objectContentVersionFile := range objectContentVersionFiles {
					// check all inventories which are less in version
					if !slices.Contains(versionManifestFiles, objectContentVersionFile) {
						object.AddValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, inventoryVersion)
					}
				}
			}
		}
		rootVersion := object.i.GetHead()
		if objectContentVersionNumber.Less(rootVersion) {
			rootManifestFiles := ocfl.SeqToSlice(object.i.GetManifest().GetFilesFlat())
			for _, objectContentVersionFile := range objectContentVersionFiles {
				// check all inventories which are less in version
				if !slices.Contains(rootManifestFiles, objectContentVersionFile) {
					object.AddValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, rootVersion)
				}
			}
		}
	}

	return nil
}

func (object *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	object.logger.Info().Msgf("object '%s' with object version '%s' found", object.GetID(), object.GetOCFLVersion())
	// check folders

	// check for allowed files and directories
	allowedDirs := []string{"logs", "extensions"}
	for v := range object.i.GetVersions().GetVersionNumbers() {
		allowedDirs = append(allowedDirs, v.String())
	}
	versionCounter := 0
	entries, err := fs.ReadDir(object.fsys, ".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !slices.Contains(allowedDirs, entry.Name()) {
				object.AddValidationError(validation.E001, "invalid directory '%s' found", entry.Name())
				// could it be a version folder?
				if _, err := strconv.Atoi(strings.TrimLeft(entry.Name(), "v0")); err == nil {
					if err2 := object.checkVersionFolder(entry.Name()); err2 == nil {
						object.AddValidationError(validation.E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			for v := range object.i.GetVersions().GetVersionNumbers() {
				if v.String() == entry.Name() {
					if err := object.checkVersionFolder(entry.Name()); err != nil {
						return errors.WithStack(err)
					}
					versionCounter++
					break
				}
			}
		} else {
			if !allowedFilesRegexp.MatchString(entry.Name()) {
				object.AddValidationError(validation.E001, "invalid file '%s' found", entry.Name())
			}
		}
	}

	invVersionCounter := len(ocfl.SeqToSlice(object.i.GetVersions().GetVersionNumbers()))
	if versionCounter != invVersionCounter {
		object.AddValidationError(validation.E010, "number of version in inventory (%v) does not fit version in filesystem (%v)", versionCounter, invVersionCounter)
	}

	if err := object.checkFilesAndVersions(); err != nil {
		return errors.WithStack(err)
	}

	dAlgs := []checksum.DigestAlgorithm{object.i.GetDigestAlgorithm()}
	dAlgs = append(dAlgs, ocfl.SeqToSlice(object.i.GetFixity().GetDigestAlgorithms())...)
	return nil
}

// create checksums of all content files
func (object *ObjectBase) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	// get all possible digest algs
	digestAlgorithms := append(ocfl.SeqToSlice(object.GetInventory().GetFixity().GetDigestAlgorithms()), object.GetDigestAlgorithm())

	result := map[checksum.DigestAlgorithm]map[string][]string{}
	for versionNumber := range object.i.GetVersions().GetVersionNumbers() {
		if err := fs.WalkDir(
			object.fsys,
			//fmt.Sprintf("%s/%s", version, object.i.GetContentDir()),
			versionNumber.String(),
			func(path string, d fs.DirEntry, err error) error {
				//object.logger.Debug(path)
				if d.IsDir() {
					return nil
				}
				fname := path // filepath.ToSlash(filepath.Join(version, path))
				fp, err := object.fsys.Open(fname)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s'", object.fsys, fname)
				}
				defer fp.Close()
				css, err := checksum.Copy(digestAlgorithms, fp, &checksum.NullWriter{})
				if err != nil {
					return errors.Wrapf(err, "cannot read and create checksums for file '%s'", fname)
				}
				for d, cs := range css {
					if _, ok := result[d]; !ok {
						result[d] = map[string][]string{}
					}
					if _, ok := result[d][cs]; !ok {
						result[d][cs] = []string{}
					}
					result[d][cs] = append(result[d][cs], fname)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrapf(err, "cannot walk content dir '%s'", object.i.GetContentDir())
		}
	}
	return result, nil
}

var ObjectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

// helper functions

func (object *ObjectBase) getVersionInventories() (map[string]inventory.Inventory, error) {
	if len(object.versionInventories) > 0 {
		return object.versionInventories, nil
	}

	versionStrings := ocfl.SeqToSlice(object.i.GetVersions().GetVersionNumbers())

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b *inventory.VersionNumber) int {
		if a.Less(b) {
			return -1
		}

		if a.Equal(b) {
			return 0
		}

		return 1
	})
	versionInventories := map[string]inventory.Inventory{}
	for _, ver := range versionStrings {
		vi, err := object.LoadInventory(ver.String())
		if err != nil {
			if errors.Is(errors.Cause(err), fs.ErrNotExist) {
				object.AddValidationWarning(validation.W010, "no inventory for version '%s'", ver)
				continue
			}
			return nil, errors.Wrapf(err, "cannot load inventory from folder '%s'", ver)
		}
		versionInventories[ver.String()] = vi
	}
	object.versionInventories = versionInventories
	return object.versionInventories, nil
}

/*
func (object *ObjectBase) getAllDigests() ([]checksum.DigestAlgorithm, error) {
	versionInventories, err := object.getVersionInventories()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get version inventories")
	}
	allDigestAlgs := []checksum.DigestAlgorithm{object.i.GetDigestAlgorithm()}
	for _, vi := range versionInventories {
		allDigestAlgs = append(allDigestAlgs, vi.GetDigestAlgorithm())
		for digestAlg := range vi.GetFixity().GetDigestAlgorithms() {
			allDigestAlgs = append(allDigestAlgs, digestAlg)
		}
	}
	slices.Sort(allDigestAlgs)
	allDigestAlgs = slices.Compact(allDigestAlgs)
	return allDigestAlgs, nil
}
*/

func (object *ObjectBase) Extract(fsys fs.FS, version *inventory.VersionNumber, withManifest bool, area string) error {
	var manifest strings.Builder
	var err error
	var digestAlg = object.i.GetDigestAlgorithm()
	if err := object.i.IterateFiles(version, func(internals, externals []string, digest string) error {
		for _, external := range externals {
			external, err = object.extensionManager.BuildObjectExtractPath(object, external, area)
			if err != nil {
				errCause := errors.Cause(err)
				if errors.Is(errCause, object2.ExtensionObjectExtractPathWrongAreaError) {
					return nil
				}
				return errors.Wrapf(err, "cannot map path '%s'", external)
			}
			if err := func() error {
				if len(internals) == 0 {
					return errors.Errorf("no internal paths for '%v'", externals)
				}
				internal := internals[0]
				src, err := object.fsys.Open(internal)
				if err != nil {
					return errors.Wrapf(err, "cannot open '%v/%s'", object.fsys, internal)
				}
				defer src.Close()
				target, err := writefs.Create(fsys, external)
				if err != nil {
					return errors.Wrapf(err, "cannot create '%v/%s'", fsys, external)
				}
				defer target.Close()
				object.logger.Debug().Msgf("writing '%v/%s' -> '%v/%s'", object.fsys, internal, fsys, external)
				copyDigests, err := checksum.Copy([]checksum.DigestAlgorithm{digestAlg}, src, target)
				if err != nil {
					return errors.Wrapf(err, "error copying '%v/%s' -> '%v/%s'", object.fsys, internal, fsys, external)
				}
				copyDigest, ok := copyDigests[digestAlg]
				if !ok {
					return errors.Errorf("no digest '%s' generatied", digestAlg)
				}
				if copyDigest != digest {
					return errors.Errorf("invalid digest for '%s' - [%s] != [%s]", internal, copyDigests, digest)
				}
				return nil
			}(); err != nil {
				return err
			}
			if withManifest {
				manifest.WriteString(fmt.Sprintf("%s %s\n", digest, external))
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot iterate external files")
	}
	if withManifest {
		manifestName := fmt.Sprintf("manifest.%s", digestAlg)
		fp, err := writefs.Create(fsys, manifestName)
		if err != nil {
			return errors.Wrapf(err, "cannot crate manifest file %v/%s", fsys, manifestName)
		}
		if _, err := io.WriteString(fp, manifest.String()); err != nil {
			return errors.Wrapf(err, "cannot write manifest file %v/%s", fsys, manifestName)
		}
		defer fp.Close()
	}
	object.logger.Debug().Msgf("object '%s' extracted", object.GetID())
	return nil
}

func (object *ObjectBase) GetAreaPath(area string) (string, error) {
	path, err := object.extensionManager.GetAreaPath(object, area)
	return path, errors.WithStack(err)
}

var _ object2.Object = (*ObjectBase)(nil)
