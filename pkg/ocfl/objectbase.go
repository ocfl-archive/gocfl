package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

//const VERSION = "1.0"

//var objectConformanceDeclaration = fmt.Sprintf("0=ocfl_object_%s", VERSION)

type ObjectBase struct {
	storageRoot        StorageRoot
	extensionManager   *ExtensionManager
	ctx                context.Context
	fs                 OCFLFS
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
func newObjectBase(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, storageRoot StorageRoot, logger *logging.Logger) (*ObjectBase, error) {
	ocfl := &ObjectBase{
		ctx:         ctx,
		fs:          fs,
		version:     defaultVersion,
		storageRoot: storageRoot,
		extensionManager: &ExtensionManager{
			extensions:        []Extension{},
			storageRootPath:   []ExtensionStorageRootPath{},
			objectContentPath: []ExtensionObjectContentPath{},
		},
		logger: logger,
	}
	return ocfl, nil
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

func (object *ObjectBase) IsModified() bool { return object.i.IsModified() }

func (object *ObjectBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%s' - '%s'", object.fs, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationErrors(object.ctx, valError)
}

func (object *ObjectBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) {
	valError := GetValidationError(object.version, errno).AppendDescription(format, a...).AppendContext("object '%s' - '%s'", object.fs, object.GetID())
	_, file, line, _ := runtime.Caller(1)
	object.logger.Debugf("[%s:%v] %s", file, line, valError.Error())
	addValidationWarnings(object.ctx, valError)
}

func (object *ObjectBase) GetFS() OCFLFS {
	return object.fs
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
			addValidationErrors(object.ctx, GetValidationError(version, E033).AppendDescription("json syntax error: %v", err2).AppendContext("object '%s'", object.fs))
			addValidationErrors(object.ctx, GetValidationError(version, E034).AppendDescription("json syntax error: %v", err2).AppendContext("object '%s'", object.fs))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(object.ctx, GetValidationError(version, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]).AppendContext("object '%s'", object.fs))
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inventory, inventory.Finalize(false)
}

// loadInventory loads inventory from existing Object
func (object *ObjectBase) LoadInventory(folder string) (Inventory, error) {
	// load inventory file
	filename := filepath.ToSlash(filepath.Join(folder, "inventory.json"))
	iFp, err := object.fs.Open(filename)
	if object.fs.IsNotExist(err) {
		return nil, err
		//object.addValidationError(E063, "no inventory file in '%s'", object.fs.String())
	}
	if err != nil {
		return newInventory(object.ctx, object, folder, object.version, object.logger)
		//return nil, errors.Wrapf(err, "cannot open '%s'", filename)
	}
	// read inventory into memory
	inventoryBytes, err := io.ReadAll(iFp)
	iFp.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read '%s'", filename)
	}
	inventory, err := object.loadInventory(inventoryBytes, folder)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := inventory.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(object.fs, sidecarPath)
	if err != nil {
		if object.fs.IsNotExist(err) {
			object.addValidationError(E058, "sidecar '%s' does not exist", sidecarPath)
		} else {
			object.addValidationError(E060, "cannot read sidecar '%s': %v", sidecarPath, err.Error())
		}
		//		object.addValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		if !strings.HasSuffix(digestString, " inventory.json") {
			object.addValidationError(E061, "no suffix \" inventory.json\" in '%s'", sidecarPath)
		} else {
			digestString = strings.TrimSpace(strings.TrimSuffix(digestString, " inventory.json"))
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

func (object *ObjectBase) StoreInventory() error {
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
	iWriter, err := object.fs.Create(iFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "cannot write to inventory.json")
	}
	iFileName = fmt.Sprintf("%s/inventory.json", object.i.GetHead())
	iWriter, err = object.fs.Create(iFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "cannot write to inventory.json")
	}
	csFileName := fmt.Sprintf("inventory.json.%s", string(object.i.GetDigestAlgorithm()))
	iCSWriter, err := object.fs.Create(csFileName)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return errors.Wrapf(err, "cannot write to '%s'", csFileName)
	}
	csFileName = fmt.Sprintf("%s/inventory.json.%s", object.i.GetHead(), string(object.i.GetDigestAlgorithm()))
	iCSWriter, err = object.fs.Create(csFileName)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return errors.Wrapf(err, "cannot write to '%s'", csFileName)
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
	fp, err := object.fs.Open(objectConformanceDeclarationFile)
	if err == nil {
		// not empty, close it and return error
		if err := fp.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%s'", objectConformanceDeclarationFile)
		}
		return fmt.Errorf("cannot create object '%s'. '%s' already exists", id, objectConformanceDeclarationFile)
	}
	cnt, err := object.fs.ReadDir(".")
	if err != nil && err != fs.ErrNotExist {
		return errors.Wrapf(err, "cannot read '%s'", ".")
	}
	if len(cnt) > 0 {
		return fmt.Errorf("'%s' is not empty", ".")
	}
	rfp, err := object.fs.Create(objectConformanceDeclarationFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", objectConformanceDeclarationFile)
	}
	defer rfp.Close()
	if _, err := rfp.Write([]byte(objectConformanceDeclaration + "\n")); err != nil {
		return errors.Wrapf(err, "cannot write into '%s'", objectConformanceDeclarationFile)
	}

	for _, ext := range extensions {
		if err := object.extensionManager.Add(ext); err != nil {
			return errors.Wrapf(err, "cannot add extension '%s'", ext.GetName())
		}
	}
	object.extensionManager.Finalize()

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
	exts, err := object.fs.ReadDir("extensions")
	if err != nil {
		// if directory does not exist - no problem
		if err != fs.ErrNotExist {
			return errors.Wrap(err, "cannot read extensions folder")
		}
		exts = []fs.DirEntry{}
	}
	for _, extFolder := range exts {
		if !extFolder.IsDir() {
			object.addValidationError(E067, "invalid file '%s' in extension dir", extFolder.Name())
			continue
		}
		extConfig := fmt.Sprintf("extensions/%s", extFolder.Name())
		subfs, err := object.fs.SubFS(extConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot create subfs of %v for '%s'", object.fs, extConfig)
		}
		if ext, err := object.storageRoot.CreateExtension(subfs); err != nil {
			//return errors.Wrapf(err, "create extension of extensions/%s", extFolder.Name())
			object.addValidationWarning(W013, "unknown extension in folder '%s'", extFolder.Name())
		} else {
			if err := object.extensionManager.Add(ext); err != nil {
				return errors.Wrapf(err, "cannot add extension '%s'", extFolder.Name())
			}
		}
	}
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
	basePath, err := object.extensionManager.BuildObjectExternalPath(object, ".", object.area)
	if err != nil {
		return errors.Wrap(err, "cannot build external path for '.'")
	}
	if err := object.i.echoDelete(object.updateFiles, basePath); err != nil {
		return errors.Wrap(err, "cannot remove deleted files from inventory")
	}
	return nil
}

func (object *ObjectBase) Close() error {
	object.logger.Debug(fmt.Sprintf("Closing object '%s'", object.GetID()))
	if !(object.i.IsWriteable()) {
		return nil
	}

	if err := object.extensionManager.UpdateObjectAfter(object); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
	}

	if object.echo {
		if err := object.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot delete files")
		}
	}
	if !object.i.IsModified() {
		return nil
	}
	object.storageRoot.setModified()
	if err := object.i.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := object.StoreInventory(); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}
	if err := object.StoreExtensions(); err != nil {
		return errors.Wrap(err, "cannot store extensions")
	}
	return nil
}

func (object *ObjectBase) StartUpdate(msg string, UserName string, UserAddress string, echo bool) error {
	object.logger.Debugf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)
	object.echo = echo

	subfs, err := object.fs.SubFS("extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", object.fs, "extensions")
	}
	object.extensionManager.SetFS(subfs)

	if object.i.IsWriteable() {
		return errors.New("object already writeable")
	}
	if err := object.i.NewVersion(msg, UserName, UserAddress); err != nil {
		return errors.Wrap(err, "cannot create new object version")
	}
	if err := object.extensionManager.UpdateObjectBefore(object); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectBefore()")
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

func (object *ObjectBase) AddFolder(fsys fs.FS, checkDuplicate bool, area string) error {
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		path = filepath.ToSlash(path)
		// directory not interesting
		if info.IsDir() {
			return nil
		}
		/*
			realFilename, err := object.extensionManager.BuildObjectContentPath(object, path, area)
			if err != nil {
				return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
			}
		*/
		if err := object.AddFile(fsys, path, checkDuplicate, area); err != nil {
			return errors.Wrapf(err, "cannot add file '%s'", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (object *ObjectBase) AddReader(r io.ReadCloser, internalFilename string, area string) error {

	digestAlgorithms := object.i.GetFixityDigestAlgorithm()

	object.updateFiles = append(object.updateFiles, internalFilename)

	// file could be replaced by another file
	defer r.Close()

	var digest string
	if !slices.Contains(digestAlgorithms, object.i.GetDigestAlgorithm()) {
		digestAlgorithms = append(digestAlgorithms, object.i.GetDigestAlgorithm())
	}

	targetFilename := object.i.BuildRealname(internalFilename)
	writer, err := object.fs.Create(targetFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetFilename)
	}
	defer writer.Close()
	csw := checksum.NewChecksumWriter(digestAlgorithms)
	checksums, err := csw.Copy(writer, r)
	if err != nil {
		return errors.Wrapf(err, "cannot copy '%s' -> '%s'", internalFilename, targetFilename)
	}
	/*
		if digest != "" && digest != checksums[object.i.GetDigestAlgorithm()] {
			return fmt.Errorf("invalid checksum '%s'", digest)
		}
	*/
	if digest == "" {
		var ok bool
		digest, ok = checksums[object.i.GetDigestAlgorithm()]
		if !ok {
			return errors.Errorf("digest '%s' not generated", object.i.GetDigestAlgorithm())
		}
	} else {
		checksums[object.i.GetDigestAlgorithm()] = digest
	}
	if err := object.i.AddFile(internalFilename, targetFilename, checksums); err != nil {
		return errors.Wrapf(err, "cannot append '%s'/'%s' to inventory", internalFilename, internalFilename)
	}
	return nil
}

func (object *ObjectBase) AddFile(fsys fs.FS, path string, checkDuplicate bool, area string) error {
	//object.logger.Infof("[%s] adding '%s' -> '%s'", object.GetID(), sourceFilename, internalFilename)
	// paranoia
	path = filepath.ToSlash(path)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}
	internalFilename, err := object.extensionManager.BuildObjectContentPath(object, path, area)
	if err != nil {
		return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
	}

	digestAlgorithms := object.i.GetFixityDigestAlgorithm()

	object.updateFiles = append(object.updateFiles, internalFilename)

	file, err := fsys.Open(path)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", path)
	}
	// file could be replaced by another file
	defer func() {
		file.Close()
	}()
	var digest string
	newPath, err := object.extensionManager.BuildObjectExternalPath(object, path, area)
	if err != nil {
		return errors.Wrapf(err, "cannot map external path '%s'", path)
	}
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
				return errors.Wrapf(err, "cannot open file '%s'", path)
			}
		}
		// if file is already there we do nothing
		dup, err := object.i.AlreadyExists(newPath, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", internalFilename, digest)
		}
		if dup {
			object.logger.Infof("[%s] ignoring '%s'", object.GetID(), newPath)
			return nil
		}
		// file already ingested, but new virtual name
		if dups := object.i.GetDuplicates(digest); len(dups) > 0 {
			if err := object.i.CopyFile(newPath, digest); err != nil {
				return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, internalFilename)
			}
			return nil
		}
	} else {
		if !slices.Contains(digestAlgorithms, object.i.GetDigestAlgorithm()) {
			digestAlgorithms = append(digestAlgorithms, object.i.GetDigestAlgorithm())
		}
	}

	targetFilename := object.i.BuildRealname(internalFilename)
	writer, err := object.fs.Create(targetFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetFilename)
	}
	defer writer.Close()
	csw := checksum.NewChecksumWriter(digestAlgorithms)
	checksums, err := csw.Copy(writer, file)
	if err != nil {
		return errors.Wrapf(err, "cannot copy '%s' -> '%s'", path, targetFilename)
	}
	/*
		if digest != "" && digest != checksums[object.i.GetDigestAlgorithm()] {
			return fmt.Errorf("invalid checksum '%s'", digest)
		}
	*/
	if digest == "" {
		var ok bool
		digest, ok = checksums[object.i.GetDigestAlgorithm()]
		if !ok {
			return errors.Errorf("digest '%s' not generated", object.i.GetDigestAlgorithm())
		}
	} else {
		checksums[object.i.GetDigestAlgorithm()] = digest
	}
	if err := object.i.AddFile(newPath, targetFilename, checksums); err != nil {
		return errors.Wrapf(err, "cannot append '%s'/'%s' to inventory", path, internalFilename)
	}
	return nil
}

func (object *ObjectBase) DeleteFile(virtualFilename string, reader io.Reader, digest string) error {
	virtualFilename = filepath.ToSlash(virtualFilename)
	object.logger.Debugf("removing '%s' [%s]", virtualFilename, digest)

	if !object.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is already there we do nothing
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
	versionEntries, err := object.fs.ReadDir(version)
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder '%s'", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !allowedFilesRegexp.MatchString(ve.Name()) {
				object.addValidationError(E015, "extra file '%s' in version directory '%s'", ve.Name(), version)
			}
			// else {
			//	if ve.GetName() != "content" {
			//		object.addValidationError(E022, "forbidden subfolder '%s' in version directory '%s'", ve.GetName(), version)
			//	}
		}
	}
	return nil
}

func (object *ObjectBase) checkFilesAndVersions() error {
	// create list of version content directories
	versionContents := map[string]string{}
	versionStrings := object.i.GetVersionStrings()

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b string) bool {
		return object.i.VersionLessOrEqual(a, b) && a != b
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
		object.fs.WalkDir(
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
			fi, err := object.fs.Stat(versionContent)
			if err != nil {
				if !object.fs.IsNotExist(err) {
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
		versionEntries, err := object.fs.ReadDir(ver)
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
	entries, err := object.fs.ReadDir(".")
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
	checksumWriter := checksum.NewChecksumWriter(digestAlgorithms)
	versions := object.i.GetVersionStrings()
	for _, version := range versions {
		if err := object.fs.WalkDir(
			//fmt.Sprintf("%s/%s", version, object.i.GetContentDir()),
			version,
			func(path string, d fs.DirEntry, err error) error {
				//object.logger.Debug(path)
				if d.IsDir() {
					return nil
				}
				fp, err := object.fs.Open(path)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%s'", path)
				}
				defer fp.Close()
				css, err := checksumWriter.Copy(&checksum.NullWriter{}, fp)
				if err != nil {
					return errors.Wrapf(err, "cannot read and create checksums for file '%s'", path)
				}
				for d, cs := range css {
					if _, ok := result[d]; !ok {
						result[d] = map[string][]string{}
					}
					if _, ok := result[d][cs]; !ok {
						result[d][cs] = []string{}
					}
					result[d][cs] = append(result[d][cs], path)
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
	slices.SortFunc(versionStrings, func(a, b string) bool {
		return object.i.VersionLessOrEqual(a, b) && a != b
	})
	versionInventories := map[string]Inventory{}
	for _, ver := range versionStrings {
		vi, err := object.LoadInventory(ver)
		if err != nil {
			if object.fs.IsNotExist(err) {
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
