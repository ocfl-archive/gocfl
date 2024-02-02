package ocfl

import (
	"bytes"
	"cmp"
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

//const VERSION = "1.0"

//var objectConformanceDeclaration = fmt.Sprintf("0=ocfl_object_%s", VERSION)

type ObjectBase struct {
	storageRoot        StorageRoot
	extensionManager   *ExtensionManager
	ctx                context.Context
	fsys               fs.FS
	i                  Inventory
	versionFolders     []string
	versionInventories map[string]Inventory
	changed            bool
	logger             *logging.Logger
	version            OCFLVersion
	digest             checksum.DigestAlgorithm
	echo               bool
	updateFiles        []string
	area               string
}

// newObjectBase creates an empty ObjectBase structure
func newObjectBase(ctx context.Context, fsys fs.FS, defaultVersion OCFLVersion, storageRoot StorageRoot, logger *logging.Logger) (*ObjectBase, error) {
	ocfl := &ObjectBase{
		ctx:         ctx,
		fsys:        fsys,
		version:     defaultVersion,
		storageRoot: storageRoot,
		extensionManager: &ExtensionManager{
			extensions:        []Extension{},
			storageRootPath:   []ExtensionStorageRootPath{},
			objectContentPath: []ExtensionObjectContentPath{},
			ExtensionManagerConfig: &ExtensionManagerConfig{
				Sort:      map[string][]string{},
				Exclusion: map[string][][]string{},
			},
		},
		logger: logger,
	}
	return ocfl, nil
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

func (o *ObjectBase) GetExtensionManager() *ExtensionManager {
	return o.extensionManager
}

func (object *ObjectBase) IsModified() bool { return object.i.IsModified() }

func (object *ObjectBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%v' - '%s'", object.fsys, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationErrors(object.ctx, valError)
}

func (object *ObjectBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%v' - '%s'", object.fsys, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationWarnings(object.ctx, valError)
}

func (object *ObjectBase) GetMetadata() (*ObjectMetadata, error) {
	inventory := object.GetInventory()
	if inventory == nil {
		return nil, errors.Errorf("inventory is nil")
	}

	result := &ObjectMetadata{
		ID:              object.GetID(),
		Head:            inventory.GetHead(),
		Files:           map[string]*FileMetadata{},
		DigestAlgorithm: object.GetDigestAlgorithm(),
		Versions:        map[string]*VersionMetadata{},
	}
	manifest := inventory.GetManifest()
	versions := inventory.GetVersions()
	fixity := inventory.GetFixity()
	versionStrings := []string{}
	for v, version := range versions {
		result.Versions[v] = &VersionMetadata{
			Created: version.Created.Time,
			Message: version.Message.string,
			Name:    version.User.Name.string,
			Address: version.User.Address.string,
		}
		versionStrings = append(versionStrings, v)
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
	for digest, fnames := range manifest {
		if len(fnames) == 0 {
			continue
		}
		fm := &FileMetadata{
			Checksums:    map[checksum.DigestAlgorithm]string{},
			InternalName: fnames,
			VersionName:  map[string][]string{},
			Extension:    map[string]any{},
		}
		fm.Checksums = fixity.Checksums(fnames[0])
		for v, version := range versions {
			for d, fnames := range version.State.State {
				if digest == d {
					if _, ok := fm.VersionName[v]; !ok {
						fm.VersionName[v] = []string{}
					}
					fm.VersionName[v] = append(fm.VersionName[v], fnames...)
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

func (object *ObjectBase) Stat(w io.Writer, statInfo []StatInfo) error {
	fmt.Fprintf(w, "[%s] Path: %s\n", object.GetID(), object.GetDigestAlgorithm())
	i := object.GetInventory()
	fmt.Fprintf(w, "[%s] Head: %s\n", object.GetID(), i.GetHead())
	f := i.GetFixity()
	algs := []string{}
	for alg, _ := range f {
		algs = append(algs, string(alg))
	}
	fmt.Fprintf(w, "[%s] Fixity: %s\n", object.GetID(), strings.Join(algs, ", "))
	m := i.GetManifest()
	cnt := 0
	for _, fs := range m {
		cnt += len(fs)
	}
	fmt.Fprintf(w, "[%s] Manifest: %v files (%v unique files)\n", object.GetID(), cnt, len(m))
	if slices.Contains(statInfo, StatObjectVersions) || len(statInfo) == 0 {
		for vString, version := range i.GetVersions() {
			fmt.Fprintf(w, "[%s] Version %s\n", object.GetID(), vString)
			fmt.Fprintf(w, "[%s]     User: %s (%s)\n", object.GetID(), version.User.User.Name.string, version.User.User.Address.string)
			fmt.Fprintf(w, "[%s]     Created: %s\n", object.GetID(), version.Created.String())
			fmt.Fprintf(w, "[%s]     Message: %s\n", object.GetID(), version.Message.string)
			if slices.Contains(statInfo, StatObjectVersionState) || len(statInfo) == 0 {
				state := version.State.State
				for cs, sList := range state {
					for _, s := range sList {
						fmt.Fprintf(w, "[%s]        %s\n", object.GetID(), s)
						if slices.Contains(statInfo, StatObjectManifest) || len(statInfo) == 0 {
							ms, ok := m[cs]
							if ok {
								for _, m := range ms {
									fmt.Fprintf(w, "[%s]           %s\n", object.GetID(), m)
								}
							}
						}
					}
				}
			}
		}
	}
	if slices.Contains(statInfo, StatObjectExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(object.extensionManager.ExtensionManagerConfig, "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		fmt.Fprintf(w, "[%s] Initial Extension:\n---\n%s\n---\n", object.GetID(), string(data))
		fmt.Fprintf(w, "[%s] Extension Configurations:\n", object.GetID())
		for _, ext := range object.extensionManager.extensions {
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

func (object *ObjectBase) CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error) {
	inventory, err := newInventory(object.ctx, object, "new", object.GetVersion(), object.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := inventory.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize empty inventory")
	}

	return inventory, inventory.Finalize(true)
}
func (object *ObjectBase) GetInventory() Inventory {
	return object.i
}

func (object *ObjectBase) loadInventory(data []byte, folder string) (Inventory, error) {
	anyMap := map[string]any{}
	if err := json.Unmarshal(data, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(data))
	}
	var version OCFLVersion
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
		version = Version1_1
	case "https://ocfl.io/1.0/spec/#inventory":
		version = Version1_0
	default:
		// if we don't know anything use the old stuff
		version = Version1_0
	}
	inventory, err := newInventory(object.ctx, object, folder, version, object.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := json.Unmarshal(data, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			addValidationErrors(object.ctx, GetValidationError(version, E033).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", object.fsys))
			addValidationErrors(object.ctx, GetValidationError(version, E034).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", object.fsys))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(object.ctx, GetValidationError(version, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]).AppendContext("object '%v'", object.fsys))
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inventory, inventory.Finalize(false)
}

var inventorySideCarFormat = regexp.MustCompile(`^([a-fA-F0-9]+)\s+inventory.json$`)

// loadInventory loads inventory from existing Object
func (object *ObjectBase) LoadInventory(folder string) (Inventory, error) {
	// load inventory file
	filename := filepath.ToSlash(filepath.Join(folder, "inventory.json"))
	inventoryBytes, err := fs.ReadFile(object.fsys, filename)
	if err != nil {
		if errors.Cause(err) == fs.ErrNotExist {
			return nil, err
		}
		return newInventory(object.ctx, object, folder, object.version, object.logger)
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
		if errors.Cause(err) == fs.ErrNotExist {
			object.addValidationError(E058, "sidecar '%v/%s' does not exist", object.fsys, sidecarPath)
		} else {
			object.addValidationError(E060, "cannot read sidecar '%v/%s': %v", object.fsys, sidecarPath, err.Error())
		}
		//		object.addValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		//if !strings.HasSuffix(digestString, " inventory.json") {
		matches := inventorySideCarFormat.FindStringSubmatch(digestString)
		if /* matches == nil || */ len(matches) == 0 {
			object.addValidationError(E061, "no suffix \" inventory.json\" in '%v/%s'", object.fsys, sidecarPath)
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
				object.addValidationError(E060, "'%s' != '%s'", digestString, inventoryDigestString)
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

func (object *ObjectBase) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensions []Extension) error {
	object.logger.Debugf("%s", id)

	objectConformanceDeclaration := "ocfl_object_" + string(object.version)
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration

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
	if err != nil && err != fs.ErrNotExist {
		return errors.Wrapf(err, "cannot read '%v/%s'", object.fsys, ".")
	}
	if len(cnt) > 0 {
		return fmt.Errorf("'%v/%s' is not empty", ".", object.fsys)
	}
	rfp, err := writefs.Create(object.fsys, objectConformanceDeclarationFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", object.fsys, objectConformanceDeclarationFile)
	}
	if _, err := rfp.Write([]byte(objectConformanceDeclaration + "\n")); err != nil {
		rfp.Close()
		return errors.Wrapf(err, "cannot write into '%v/%s'", object.fsys, objectConformanceDeclarationFile)
	}
	if err := rfp.Close(); err != nil {
		return errors.Wrapf(err, "cannot close '%v/%s'", object.fsys, objectConformanceDeclarationFile)
	}

	for _, ext := range extensions {
		if !ext.IsRegistered() {
			object.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
		}
		if err := object.extensionManager.Add(ext); err != nil {
			return errors.Wrapf(err, "cannot add extension '%s'", ext.GetName())
		}
	}
	object.extensionManager.Finalize()

	subfs, err := writefs.SubFSCreate(object.fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	object.extensionManager.SetFS(subfs)

	// check fixity here
	algs := []checksum.DigestAlgorithm{
		checksum.DigestSHA512,
		checksum.DigestSHA256,
	}
	algs = append(algs, object.extensionManager.GetFixityDigests()...)
	slices.Sort(algs)
	algs = slices.Compact(algs)
	if !sliceContains(algs, fixity) {
		return errors.Errorf("forbidden digest algorithm for fixity %v. Supported algorithms are %v. (to fix try to use extension 0001-digest-algorithms)", fixity, algs)
	}

	object.i, err = object.CreateInventory(id, digest, fixity)
	return nil
}

func (object *ObjectBase) Load() (err error) {
	// first check whether object already exists
	//object.version, err = GetObjectVersion(object.ctx, object.fs)
	//if err != nil {
	//	return err
	//}
	// read path from extension folder...
	exts, err := fs.ReadDir(object.fsys, "extensions")
	if err != nil {
		// if directory does not exist - no problem
		if err != fs.ErrNotExist {
			return errors.Wrapf(err, "cannot read extensions folder %v/%s", object.fsys, "extensions")
		}
		exts = []fs.DirEntry{}
	}
	for _, extFolder := range exts {
		if !extFolder.IsDir() {
			object.addValidationError(E067, "invalid file '%v/%s' in extension dir", object.fsys, extFolder.Name())
			continue
		}
		extConfig := fmt.Sprintf("extensions/%s", extFolder.Name())
		subfs, err := fs.Sub(object.fsys, extConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot create subfs of %v for '%s'", object.fsys, extConfig)
		}
		if ext, err := object.storageRoot.CreateExtension(subfs); err != nil {
			//return errors.Wrapf(err, "create extension of extensions/%s", extFolder.Name())
			object.addValidationWarning(W000, "cannot initialize extension in folder '%s'", subfs)
		} else {
			if !ext.IsRegistered() {
				object.addValidationWarning(W013, "extension '%s' is not registered", ext.GetName())
			}
			if err := object.extensionManager.Add(ext); err != nil {
				return errors.Wrapf(err, "cannot add extension '%s'", extFolder.Name())
			}
		}
	}

	subfs, err := fs.Sub(object.fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	object.extensionManager.SetFS(subfs)

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
	if err := object.i.echoDelete(object.updateFiles, basePath); err != nil {
		return errors.Wrap(err, "cannot remove deleted files from inventory")
	}
	return nil
}

func (object *ObjectBase) Close() error {
	object.logger.Infof(fmt.Sprintf("Closing object '%s'", object.GetID()))
	if !(object.i.IsWriteable()) {
		return nil
	}

	if !object.i.IsModified() {
		return nil
	}
	object.storageRoot.setModified()
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
	object.logger.Debugf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)
	object.echo = echo

	subfs, err := writefs.SubFSCreate(object.fsys, "extensions")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fsys, "extensions")
	}
	object.extensionManager.SetFS(subfs)

	if err := object.i.NewVersion(msg, UserName, UserAddress); err != nil {
		return nil, errors.Wrap(err, "cannot create new object version")
	}
	if err := object.extensionManager.UpdateObjectBefore(object); err != nil {
		return nil, errors.Wrapf(err, "cannot execute ext.UpdateObjectBefore()")
	}
	var versionFS fs.FS
	return versionFS, nil
}

func (object *ObjectBase) EndUpdate() error {
	object.logger.Infof(fmt.Sprintf("EndUpdate of object '%s'", object.GetID()))
	if !(object.i.IsWriteable()) {
		object.logger.Warningf(fmt.Sprintf("object '%s' not writeable", object.GetID()))
		return nil
	}
	if !(object.i.IsModified()) {
		object.logger.Infof(fmt.Sprintf("object '%s' not modified", object.GetID()))
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
		if _, err := object.StartUpdate(nil, "automated version", "gocfl", "https://github.com/je4/gocfl", false); err != nil {
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
	object.logger.Debugf("walking '%v'", fsys)
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
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

func (object *ObjectBase) addReader(r io.ReadCloser, versionFS fs.FS, names *NamesStruct, noExtensionHook bool) (string, error) {

	digestAlgorithms := object.i.GetFixityDigestAlgorithm()

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
		_ = pw.Close()
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%s' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
		wg.Wait()
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

func (object *ObjectBase) BuildNames(files []string, area string) (*NamesStruct, error) {
	var err error
	result := &NamesStruct{
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

	object.logger.Infof("adding file %s:%v", area, files)

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

	digestAlgorithms := object.i.GetFixityDigestAlgorithm()
	var digest string

	names, err := object.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot names for '%s'", path)

	}

	object.logger.Infof("adding file %s:%s", area, path)

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
			object.logger.Infof("[%s] '%s' already exists. ignoring", object.GetID(), newPath)
			return nil
		}
		// file already ingested, but new virtual name
		if dups := object.i.GetDuplicates(digest); len(dups) > 0 {
			object.logger.Infof("[%s] file with same content as '%s' already exists. creating virtual copy", object.GetID(), newPath)
			if err := object.i.CopyFile(newPath, digest); err != nil {
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
	object.logger.Infof("adding file %s:%s", area, path)

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

		digestAlgorithms := object.i.GetFixityDigestAlgorithm()

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
				object.logger.Infof("[%s] '%s' already exists. ignoring", object.GetID(), newPath)
				return nil
			}
			// file already ingested, but new virtual name
			if dups := object.i.GetDuplicates(digest); len(dups) > 0 {
				object.logger.Infof("[%s] file with same content as '%s' already exists. creating virtual copy", object.GetID(), newPath)
				if err := object.i.CopyFile(newPath, digest); err != nil {
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
	virtualFilename = filepath.ToSlash(virtualFilename)
	object.logger.Debugf("removing '%s' [%s]", virtualFilename, digest)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := object.i.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilename, digest)
	}
	if !dup {
		object.logger.Debugf("'%s' [%s] not in archive - ignoring", virtualFilename, digest)
		return nil
	}
	if err := object.i.DeleteFile(virtualFilename); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilename)
	}
	return nil

}

func (object *ObjectBase) RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error {
	virtualFilenameSource = filepath.ToSlash(virtualFilenameSource)
	object.logger.Debugf("removing '%s' [%s]", virtualFilenameSource, digest)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := object.i.AlreadyExists(virtualFilenameSource, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilenameSource, digest)
	}
	if !dup {
		object.logger.Debugf("'%s' [%s] not in archive - ignoring", virtualFilenameSource, digest)
		return nil
	}
	if err := object.i.RenameFile(virtualFilenameSource, virtualFilenameDest); err != nil {
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

func (object *ObjectBase) GetVersion() OCFLVersion {
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
				object.addValidationError(E015, "extra file '%s' in version directory '%s'", ve.Name(), version)
			}
		}
	}
	return nil
}

func (object *ObjectBase) checkFilesAndVersions() error {
	// create list of version content directories
	versionContents := map[string]string{}
	versionStrings := object.i.GetVersionStrings()

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b string) int {
		if object.i.VersionLessOrEqual(a, b) && a != b {
			return -1
		} else {
			if a == b {
				return 0
			} else {
				return 1
			}
		}
	})

	for _, ver := range versionStrings {
		versionContents[ver] = object.i.GetContentDir()
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
						object.addValidationWarning(W002, "extra dir '%s' in version '%s'", path, ver)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[ver] = append(objectContentFiles[ver], path)
						objectContentFilesFlat = append(objectContentFilesFlat, path)
					} else {
						/*
							if !strings.HasPrefix(path, inventoryFile) {
								object.addValidationWarning(W002, "extra file '%s' in version '%s'", path, ver)
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
				if errors.Cause(err) != fs.ErrNotExist {
					return errors.Wrapf(err, "cannot stat '%s'", versionContent)
				}
			} else {
				if fi.IsDir() {
					object.addValidationWarning(W003, "empty content folder '%s'", versionContent)
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
		contentDir = versionInventories[versionStrings[0]].GetRealContentDir()
	}
	for _, ver := range versionStrings {
		inv := versionInventories[ver]
		if inv == nil {
			continue
		}
		if contentDir != inv.GetRealContentDir() {
			object.addValidationError(E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", inv.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := inv.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", ver)
		}
		digestAlg := inv.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{inv.GetContentDir()}
		versionEntries, err := fs.ReadDir(object.fsys, ver)
		if err != nil {
			object.addValidationError(E010, "cannot read version folder '%s'", ver)
			continue
			//			return errors.Wrapf(err, "cannot read dir '%s'", ver)
		}
		for _, entry := range versionEntries {
			if entry.IsDir() {
				if !slices.Contains(allowedDirs, entry.Name()) {
					object.addValidationWarning(W002, "extra dir '%s' in version directory '%s'", entry.Name(), ver)
				}
			} else {
				if !slices.Contains(allowedFiles, entry.Name()) {
					object.addValidationError(E015, "extra file '%s' in version directory '%s'", entry.Name(), ver)
				}
			}
		}
	}

	for key := 0; key < len(versionStrings)-1; key++ {
		v1 := versionStrings[key]
		vi1, ok := versionInventories[v1]
		if !ok {
			object.addValidationWarning(W010, "no inventory for version '%s'", versionStrings[key])
			continue
			// return errors.Errorf("no inventory for version '%s'", versionStrings[key])
		}
		v2 := versionStrings[key+1]
		vi2, ok := versionInventories[v2]
		if !ok {
			object.addValidationWarning(W000, "no inventory for version '%s'", versionStrings[key+1])
			continue
		}
		if !SpecIsLessOrEqual(vi1.GetSpec(), vi2.GetSpec()) {
			object.addValidationError(E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", v1, vi1.GetSpec(), v2, vi2.GetSpec())
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion]; ok {
			if !lastInv.IsEqual(object.i) {
				object.addValidationError(E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}

	id := object.i.GetID()
	digestAlg := object.i.GetDigestAlgorithm()
	versions := object.i.GetVersions()
	for ver, verInventory := range versionInventories {
		// check for id consistency
		if id != verInventory.GetID() {
			object.addValidationError(E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", id, ver, verInventory.GetID())
		}
		if verInventory.GetHead() != "" && verInventory.GetHead() != ver {
			object.addValidationError(E040, "wrong head '%s' in manifest for version '%s'", verInventory.GetHead(), ver)
		}

		if verInventory.GetDigestAlgorithm() != digestAlg {
			object.addValidationError(W000, "different digest algorithm '%s' in version '%s'", verInventory.GetDigestAlgorithm(), ver)
		}

		for verVer, verVersion := range verInventory.GetVersions() {
			testV, ok := versions[verVer]
			if !ok {
				object.addValidationError(E066, "version '%s' in version folder '%s' not in object root manifest", ver, verVer)
			}
			if !testV.EqualState(verVersion) {
				object.addValidationError(E066, "version '%s' in version folder '%s' not equal to version in object root manifest", ver, verVer)
			}
			if !testV.EqualMeta(verVersion) {
				object.addValidationError(W011, "version '%s' in version folder '%s' has different metadata as version in object root manifest", ver, verVer)
			}
		}
	}

	//
	// all files in any manifest must belong to a physical file #E092
	//
	for inventoryVersion, inventory := range versionInventories {
		manifestFiles := inventory.GetFilesFlat()
		for _, manifestFile := range manifestFiles {
			if !slices.Contains(objectFilesFlat, manifestFile) {
				object.addValidationError(E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}

	rootManifestFiles := object.i.GetFilesFlat()
	for _, manifestFile := range rootManifestFiles {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			object.addValidationError(E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
		}
	}

	//
	// all object content files must belong to manifest
	//

	latestVersion := ""

	for objectContentVersion, objectContentVersionFiles := range objectContentFiles {
		if latestVersion == "" {
			latestVersion = objectContentVersion
		}
		if object.i.VersionLessOrEqual(latestVersion, objectContentVersion) {
			latestVersion = objectContentVersion
		}
		// check version inventories
		for inventoryVersion, versionInventory := range versionInventories {
			if versionInventory.VersionLessOrEqual(objectContentVersion, inventoryVersion) {
				versionManifestFiles := versionInventory.GetFilesFlat()
				for _, objectContentVersionFile := range objectContentVersionFiles {
					// check all inventories which are less in version
					if !slices.Contains(versionManifestFiles, objectContentVersionFile) {
						object.addValidationError(E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, inventoryVersion)
					}
				}
			}
		}
		rootVersion := object.i.GetHead()
		if object.i.VersionLessOrEqual(objectContentVersion, rootVersion) {
			rootManifestFiles := object.i.GetFilesFlat()
			for _, objectContentVersionFile := range objectContentVersionFiles {
				// check all inventories which are less in version
				if !slices.Contains(rootManifestFiles, objectContentVersionFile) {
					object.addValidationError(E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, rootVersion)
				}
			}
		}
	}

	return nil
}

func (object *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	object.logger.Infof("object '%s' with object version '%s' found", object.GetID(), object.GetVersion())
	// check folders
	versions := object.i.GetVersionStrings()

	// check for allowed files and directories
	allowedDirs := append(versions, "logs", "extensions")
	versionCounter := 0
	entries, err := fs.ReadDir(object.fsys, ".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !slices.Contains(allowedDirs, entry.Name()) {
				object.addValidationError(E001, "invalid directory '%s' found", entry.Name())
				// could it be a version folder?
				if _, err := strconv.Atoi(strings.TrimLeft(entry.Name(), "v0")); err == nil {
					if err2 := object.checkVersionFolder(entry.Name()); err2 == nil {
						object.addValidationError(E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			if slices.Contains(versions, entry.Name()) {
				err := object.checkVersionFolder(entry.Name())
				if err != nil {
					return errors.WithStack(err)
				}
				versionCounter++
			}
		} else {
			if !allowedFilesRegexp.MatchString(entry.Name()) {
				object.addValidationError(E001, "invalid file '%s' found", entry.Name())
			}
		}
	}

	if versionCounter != len(versions) {
		object.addValidationError(E010, "number of versions in inventory (%v) does not fit versions in filesystem (%v)", versionCounter, len(versions))
	}

	if err := object.checkFilesAndVersions(); err != nil {
		return errors.WithStack(err)
	}

	dAlgs := []checksum.DigestAlgorithm{object.i.GetDigestAlgorithm()}
	dAlgs = append(dAlgs, object.i.GetFixityDigestAlgorithm()...)
	return nil
}

// create checksums of all content files
func (object *ObjectBase) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	// get all possible digest algs
	digestAlgorithms, err := object.getAllDigests()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get digests")
	}

	result := map[checksum.DigestAlgorithm]map[string][]string{}
	versions := object.i.GetVersionStrings()
	for _, version := range versions {
		if err := fs.WalkDir(
			object.fsys,
			//fmt.Sprintf("%s/%s", version, object.i.GetContentDir()),
			version,
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

var objectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

// helper functions

func (object *ObjectBase) getVersionInventories() (map[string]Inventory, error) {
	if object.versionInventories != nil {
		return object.versionInventories, nil
	}

	versionStrings := object.i.GetVersionStrings()

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b string) int {
		if object.i.VersionLessOrEqual(a, b) && a != b {
			return -1
		} else {
			if a == b {
				return 0
			} else {
				return 1
			}
		}
	})
	versionInventories := map[string]Inventory{}
	for _, ver := range versionStrings {
		vi, err := object.LoadInventory(ver)
		if err != nil {
			if errors.Cause(err) == fs.ErrNotExist {
				object.addValidationWarning(W010, "no inventory for version '%s'", ver)
				continue
			}
			return nil, errors.Wrapf(err, "cannot load inventory from folder '%s'", ver)
		}
		versionInventories[ver] = vi
	}
	object.versionInventories = versionInventories
	return object.versionInventories, nil
}

func (object *ObjectBase) getAllDigests() ([]checksum.DigestAlgorithm, error) {
	versionInventories, err := object.getVersionInventories()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get version inventories")
	}
	allDigestAlgs := []checksum.DigestAlgorithm{object.i.GetDigestAlgorithm()}
	for _, vi := range versionInventories {
		allDigestAlgs = append(allDigestAlgs, vi.GetDigestAlgorithm())
		for digestAlg, _ := range vi.GetFixity() {
			allDigestAlgs = append(allDigestAlgs, digestAlg)
		}
	}
	slices.Sort(allDigestAlgs)
	allDigestAlgs = slices.Compact(allDigestAlgs)
	return allDigestAlgs, nil
}

func (object *ObjectBase) Extract(fsys fs.FS, version string, withManifest bool, area string) error {
	var manifest strings.Builder
	var err error
	var digestAlg = object.i.GetDigestAlgorithm()
	if err := object.i.IterateStateFiles(version, func(internals, externals []string, digest string) error {
		for _, external := range externals {
			external, err = object.extensionManager.BuildObjectExtractPath(object, external, area)
			if err != nil {
				errCause := errors.Cause(err)
				if errors.Is(errCause, ExtensionObjectExtractPathWrongAreaError) {
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
				object.logger.Debugf("writing '%v/%s' -> '%v/%s'", object.fsys, internal, fsys, external)
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
	object.logger.Debugf("object '%s' extracted", object.GetID())
	return nil
}

func (object *ObjectBase) GetAreaPath(area string) (string, error) {
	path, err := object.extensionManager.GetAreaPath(object, area)
	return path, errors.WithStack(err)
}
