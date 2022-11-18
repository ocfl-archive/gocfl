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
}

// newObjectBase creates an empty ObjectBase structure
func newObjectBase(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, storageroot StorageRoot, logger *logging.Logger) (*ObjectBase, error) {
	ocfl := &ObjectBase{
		ctx:         ctx,
		fs:          fs,
		version:     defaultVersion,
		storageRoot: storageroot,
		extensionManager: &ExtensionManager{
			extensions:        []Extension{},
			storagerootPath:   []StoragerootPath{},
			objectContentPath: []ObjectContentPath{},
		},
		logger: logger,
	}
	return ocfl, nil
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

func (ocfl *ObjectBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	addValidationErrors(ocfl.ctx, GetValidationError(ocfl.version, errno).AppendDescription(format, a...).AppendContext("object '%s' - '%s'", ocfl.fs, ocfl.GetID()))
}

func (ocfl *ObjectBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) {
	addValidationWarnings(ocfl.ctx, GetValidationError(ocfl.version, errno).AppendDescription(format, a...).AppendContext("object '%s' - '%s'", ocfl.fs, ocfl.GetID()))
}

func (ocfl *ObjectBase) getFS() OCFLFS {
	return ocfl.fs
}
func (ocfl *ObjectBase) CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error) {
	inventory, err := newInventory(ocfl.ctx, ocfl, "new", ocfl.GetVersion(), ocfl.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := inventory.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize empty inventory")
	}

	return inventory, nil
}

func (ocfl *ObjectBase) loadInventory(data []byte, folder string) (Inventory, error) {
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
	inventory, err := newInventory(ocfl.ctx, ocfl, folder, version, ocfl.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := json.Unmarshal(data, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			addValidationErrors(ocfl.ctx, GetValidationError(version, E033).AppendDescription("json syntax error: %v", err2).AppendContext("object %s", ocfl.fs))
			addValidationErrors(ocfl.ctx, GetValidationError(version, E034).AppendDescription("json syntax error: %v", err2).AppendContext("object %s", ocfl.fs))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(ocfl.ctx, GetValidationError(version, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]).AppendContext("object %s", ocfl.fs))
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - %s", string(data))
	}

	return inventory, inventory.Finalize()
}

// loadInventory loads inventory from existing Object
func (ocfl *ObjectBase) LoadInventory(folder string) (Inventory, error) {
	// load inventory file
	filename := filepath.ToSlash(filepath.Join(folder, "inventory.json"))
	iFp, err := ocfl.fs.Open(filename)
	if ocfl.fs.IsNotExist(err) {
		return nil, err
		//ocfl.addValidationError(E063, "no inventory file in '%s'", ocfl.fs.String())
	}
	if err != nil {
		return newInventory(ocfl.ctx, ocfl, folder, ocfl.version, ocfl.logger)
		//return nil, errors.Wrapf(err, "cannot open %s", filename)
	}
	// read inventory into memory
	inventoryBytes, err := io.ReadAll(iFp)
	iFp.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %s", filename)
	}
	inventory, err := ocfl.loadInventory(inventoryBytes, folder)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := inventory.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(ocfl.fs, sidecarPath)
	if err != nil {
		if ocfl.fs.IsNotExist(err) {
			ocfl.addValidationError(E058, "sidecar %s does not exist", sidecarPath)
		} else {
			ocfl.addValidationError(E060, "cannot read sidecar %s: %v", sidecarPath, err.Error())
		}
		//		ocfl.addValidationError(E058, "cannot read %s: %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		if !strings.HasSuffix(digestString, " inventory.json") {
			ocfl.addValidationError(E061, "no suffix \" inventory.json\" in %s", sidecarPath)
		} else {
			digestString = strings.TrimSpace(strings.TrimSuffix(digestString, " inventory.json"))
			h, err := checksum.GetHash(digest)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - %s", string(digest)))
			}
			h.Reset()
			h.Write(inventoryBytes)
			sumBytes := h.Sum(nil)
			inventoryDigestString := fmt.Sprintf("%x", sumBytes)
			if digestString != inventoryDigestString {
				ocfl.addValidationError(E060, "%s != %s", digestString, inventoryDigestString)
			}
		}
	}

	return inventory, inventory.Finalize()
}

func (ocfl *ObjectBase) StoreInventory() error {
	ocfl.logger.Debug()

	// check whether ocfl filesystem is writeable
	if !ocfl.i.IsWriteable() {
		return errors.New("inventory not writeable - not updated")
	}

	// create inventory.json from inventory
	iFileName := "inventory.json"
	jsonBytes, err := json.MarshalIndent(ocfl.i, "", "   ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(ocfl.i.GetDigestAlgorithm())
	if err != nil {
		return errors.Wrapf(err, "invalid digest algorithm %s", string(ocfl.i.GetDigestAlgorithm()))
	}
	if _, err := h.Write(jsonBytes); err != nil {
		return errors.Wrapf(err, "cannot create checksum of manifest")
	}
	checksumBytes := h.Sum(nil)
	checksumString := fmt.Sprintf("%x %s", checksumBytes, iFileName)
	iWriter, err := ocfl.fs.Create(iFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "cannot write to inventory.json")
	}
	iFileName = fmt.Sprintf("%s/inventory.json", ocfl.i.GetHead())
	iWriter, err = ocfl.fs.Create(iFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "cannot write to inventory.json")
	}
	csFileName := fmt.Sprintf("inventory.json.%s", string(ocfl.i.GetDigestAlgorithm()))
	iCSWriter, err := ocfl.fs.Create(csFileName)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return errors.Wrapf(err, "cannot write to %s", csFileName)
	}
	csFileName = fmt.Sprintf("%s/inventory.json.%s", ocfl.i.GetHead(), string(ocfl.i.GetDigestAlgorithm()))
	iCSWriter, err = ocfl.fs.Create(csFileName)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return errors.Wrapf(err, "cannot write to %s", csFileName)
	}
	return nil
}

func (ocfl *ObjectBase) StoreExtensions() error {
	ocfl.logger.Debug()
	subfs, err := ocfl.fs.SubFS("extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder %s", ocfl.fs, "extensions")
	}

	if err := ocfl.extensionManager.StoreConfigs(subfs); err != nil {
		return errors.Wrap(err, "cannot store extension configs")
	}
	return nil
}
func (ocfl *ObjectBase) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensions []Extension) error {
	ocfl.logger.Debugf("%s", id)

	objectConformanceDeclaration := "ocfl_object_" + string(ocfl.version)
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration

	// first check whether ocfl is not empty
	fp, err := ocfl.fs.Open(objectConformanceDeclarationFile)
	if err == nil {
		// not empty, close it and return error
		if err := fp.Close(); err != nil {
			return errors.Wrapf(err, "cannot close %s", objectConformanceDeclarationFile)
		}
		return fmt.Errorf("cannot create object %s. %s already exists", id, objectConformanceDeclarationFile)
	}
	cnt, err := ocfl.fs.ReadDir(".")
	if err != nil && err != fs.ErrNotExist {
		return errors.Wrapf(err, "cannot read %s", ".")
	}
	if len(cnt) > 0 {
		return fmt.Errorf("%s is not empty", ".")
	}
	rfp, err := ocfl.fs.Create(objectConformanceDeclarationFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", objectConformanceDeclarationFile)
	}
	defer rfp.Close()
	if _, err := rfp.Write([]byte(objectConformanceDeclaration + "\n")); err != nil {
		return errors.Wrapf(err, "cannot write into %s", objectConformanceDeclarationFile)
	}

	for _, ext := range extensions {
		if err := ocfl.extensionManager.Add(ext); err != nil {
			return errors.Wrapf(err, "cannot add extension %s", ext.GetName())
		}
	}

	ocfl.i, err = ocfl.CreateInventory(id, digest, fixity)
	return nil
}

func (ocfl *ObjectBase) Load() (err error) {
	// first check whether object already exists
	//ocfl.version, err = GetObjectVersion(ocfl.ctx, ocfl.fs)
	//if err != nil {
	//	return err
	//}
	// read path from extension folder...
	exts, err := ocfl.fs.ReadDir("extensions")
	if err != nil {
		// if directory does not exist - no problem
		if err != fs.ErrNotExist {
			return errors.Wrap(err, "cannot read extensions folder")
		}
		exts = []fs.DirEntry{}
	}
	for _, extFolder := range exts {
		if !extFolder.IsDir() {
			ocfl.addValidationError(E067, "invalid file '%s' in extension dir", extFolder.Name())
			continue
		}
		extConfig := fmt.Sprintf("extensions/%s", extFolder.Name())
		subfs, err := ocfl.fs.SubFS(extConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot create subfs of %v for %s", ocfl.fs, extConfig)
		}
		if ext, err := ocfl.storageRoot.CreateExtension(subfs); err != nil {
			//return errors.Wrapf(err, "create extension of extensions/%s", extFolder.Name())
			ocfl.addValidationWarning(W013, "unknown extension in folder '%s'", extFolder.Name())
		} else {
			if err := ocfl.extensionManager.Add(ext); err != nil {
				return errors.Wrapf(err, "cannot add extension %s", extFolder.Name())
			}
		}
	}
	// load the inventory
	if ocfl.i, err = ocfl.LoadInventory("."); err != nil {
		return errors.Wrap(err, "cannot load inventory.json of root")
	}
	return nil
}

func (ocfl *ObjectBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return ocfl.i.GetDigestAlgorithm()
}
func (ocfl *ObjectBase) Close() error {
	ocfl.logger.Debug()
	if ocfl.i.IsWriteable() {
		if err := ocfl.i.Clean(); err != nil {
			return errors.Wrap(err, "cannot clean inventory")
		}
		if err := ocfl.StoreInventory(); err != nil {
			return errors.Wrap(err, "cannot store inventory")
		}
		if err := ocfl.StoreExtensions(); err != nil {
			return errors.Wrap(err, "cannot store extensions")
		}
	}
	return nil
}

func (ocfl *ObjectBase) StartUpdate(msg string, UserName string, UserAddress string) error {
	ocfl.logger.Debugf("%s / %s / %s", msg, UserName, UserAddress)

	if ocfl.i.IsWriteable() {
		return errors.New("ocfl already writeable")
	}
	if err := ocfl.i.NewVersion(msg, UserName, UserAddress); err != nil {
		return errors.Wrap(err, "cannot create new ocfl version")
	}
	return nil
}

func (ocfl *ObjectBase) AddFolder(fsys fs.FS, checkDuplicate bool) error {
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		path = filepath.ToSlash(path)
		// directory not interesting
		if info.IsDir() {
			return nil
		}
		realFilename, err := ocfl.extensionManager.BuildObjectContentPath(ocfl, path)
		if err != nil {
			return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
		}
		if err := ocfl.AddFile(fsys, path, realFilename, checkDuplicate); err != nil {
			return errors.Wrapf(err, "cannot add file %s", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (ocfl *ObjectBase) AddFile(fsys fs.FS, sourceFilename string, internalFilename string, checkDuplicate bool) error {
	ocfl.logger.Debugf("adding %s -> %s", sourceFilename, internalFilename)
	// paranoia
	internalFilename = filepath.ToSlash(internalFilename)

	if !ocfl.i.IsWriteable() {
		return errors.New("ocfl not writeable")
	}

	digestAlgorithms := ocfl.i.GetFixityDigestAlgorithm()

	file, err := fsys.Open(sourceFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", sourceFilename)
	}
	// file could be replaced by another file
	defer func() {
		file.Close()
	}()
	var digest string
	if checkDuplicate {
		// do the checksum
		digest, err = checksum.Checksum(file, ocfl.i.GetDigestAlgorithm())
		if err != nil {
			return errors.Wrapf(err, "cannot create digest of %s", sourceFilename)
		}
		// set filepointer to beginning
		if seeker, ok := file.(io.Seeker); ok {
			// if we have a seeker, we just seek
			if _, err := seeker.Seek(0, 0); err != nil {
				panic(err)
			}
		} else {
			// otherwise reopen it
			file, err = fsys.Open(sourceFilename)
			if err != nil {
				return errors.Wrapf(err, "cannot open file '%s'", sourceFilename)
			}
		}
		// if file is already there we do nothing
		dup, err := ocfl.i.AlreadyExists(sourceFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check duplicate for %s [%s]", internalFilename, digest)
		}
		if dup {
			ocfl.logger.Debugf("%s [%s] is a duplicate", internalFilename, digest)
			return nil
		}
		// file already ingested, but new virtual name
		if dups := ocfl.i.GetDuplicates(digest); len(dups) > 0 {
			if err := ocfl.i.RenameFile(sourceFilename, digest); err != nil {
				return errors.Wrapf(err, "cannot append %s to inventory as %s", sourceFilename, internalFilename)
			}
			return nil
		}
	} else {
		if !slices.Contains(digestAlgorithms, ocfl.i.GetDigestAlgorithm()) {
			digestAlgorithms = append(digestAlgorithms, ocfl.i.GetDigestAlgorithm())
		}
	}

	targetFilename := ocfl.i.BuildRealname(internalFilename)
	writer, err := ocfl.fs.Create(targetFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", targetFilename)
	}
	defer writer.Close()
	csw := checksum.NewChecksumWriter(digestAlgorithms)
	checksums, err := csw.Copy(writer, file)
	if err != nil {
		return errors.Wrapf(err, "cannot copy %s -> %s", sourceFilename, targetFilename)
	}
	/*
		if digest != "" && digest != checksums[ocfl.i.GetDigestAlgorithm()] {
			return fmt.Errorf("invalid checksum %s", digest)
		}
	*/
	if digest == "" {
		var ok bool
		digest, ok = checksums[ocfl.i.GetDigestAlgorithm()]
		if !ok {
			return errors.Errorf("digest %s not generated", ocfl.i.GetDigestAlgorithm())
		}
	} else {
		checksums[ocfl.i.GetDigestAlgorithm()] = digest
	}
	if err := ocfl.i.AddFile(sourceFilename, targetFilename, checksums); err != nil {
		return errors.Wrapf(err, "cannot append %s/%s to inventory", sourceFilename, internalFilename)
	}
	return nil
}

func (ocfl *ObjectBase) DeleteFile(virtualFilename string, reader io.Reader, digest string) error {
	virtualFilename = filepath.ToSlash(virtualFilename)
	ocfl.logger.Debugf("removing %s [%s]", virtualFilename, digest)

	if !ocfl.i.IsWriteable() {
		return errors.New("ocfl not writeable")
	}

	// if file is already there we do nothing
	dup, err := ocfl.i.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for %s [%s]", virtualFilename, digest)
	}
	if !dup {
		ocfl.logger.Debugf("%s [%s] not in archive - ignoring", virtualFilename, digest)
		return nil
	}
	if err := ocfl.i.DeleteFile(virtualFilename); err != nil {
		return errors.Wrapf(err, "cannot delete %s", virtualFilename)
	}
	return nil

}

func (ocfl *ObjectBase) GetID() string {
	if ocfl.i == nil {
		return ""
	}
	return ocfl.i.GetID()
}

func (ocfl *ObjectBase) GetVersion() OCFLVersion {
	return ocfl.version
}

var allowedFilesRegexp = regexp.MustCompile("^(inventory.json(\\.sha512|\\.sha384|\\.sha256|\\.sha1|\\.md5)?|0=ocfl_object_[0-9]+\\.[0-9]+)$")

func (ocfl *ObjectBase) checkVersionFolder(version string) error {
	versionEntries, err := ocfl.fs.ReadDir(version)
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder %s", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !allowedFilesRegexp.MatchString(ve.Name()) {
				ocfl.addValidationError(E015, "extra file '%s' in version directory '%s'", ve.Name(), version)
			}
			// else {
			//	if ve.GetName() != "content" {
			//		ocfl.addValidationError(E022, "forbidden subfolder '%s' in version directory '%s'", ve.GetName(), version)
			//	}
		}
	}
	return nil
}

func (ocfl *ObjectBase) checkFilesAndVersions() error {
	// create list of version content directories
	versionContents := map[string]string{}
	versionStrings := ocfl.i.GetVersionStrings()

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b string) bool {
		return ocfl.i.VersionLessOrEqual(a, b) && a != b
	})

	for _, ver := range versionStrings {
		versionContents[ver] = ocfl.i.GetContentDir()
	}

	// load object content files
	objectContentFiles := map[string][]string{}
	objectContentFilesFlat := []string{}
	objectFilesFlat := []string{}
	for ver, cont := range versionContents {
		// load all object version content files
		versionContent := ver + "/" + cont
		inventoryFile := ver + "/inventory.json"
		if _, ok := objectContentFiles[ver]; !ok {
			objectContentFiles[ver] = []string{}
		}
		ocfl.fs.WalkDir(
			ver,
			func(path string, d fs.DirEntry, err error) error {
				path = filepath.ToSlash(path)
				if d.IsDir() {
					if !strings.HasPrefix(path, versionContent) && path != ver && !strings.HasPrefix(ver+"/"+ocfl.i.GetContentDir(), path) {
						ocfl.addValidationWarning(W002, "extra dir '%s' in version %s", path, ver)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[ver] = append(objectContentFiles[ver], path)
						objectContentFilesFlat = append(objectContentFilesFlat, path)
					} else {
						if !strings.HasPrefix(path, inventoryFile) {
							ocfl.addValidationWarning(W002, "extra file '%s' in version %s", path, ver)
						}
					}
				}
				return nil
			},
		)
		if len(objectContentFiles[ver]) == 0 {
			fi, err := ocfl.fs.Stat(versionContent)
			if err != nil {
				if !ocfl.fs.IsNotExist(err) {
					return errors.Wrapf(err, "cannot stat '%s'", versionContent)
				}
			} else {
				if fi.IsDir() {
					ocfl.addValidationWarning(W003, "empty content folder '%s'", versionContent)
				}
			}
		}
	}
	// load all inventories
	versionInventories, err := ocfl.getVersionInventories()
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}

	csDigestFiles, err := ocfl.createContentManifest()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ocfl.i.CheckFiles(csDigestFiles); err != nil {
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
			ocfl.addValidationError(E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", inv.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := inv.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", ver)
		}
		digestAlg := inv.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{"content"}
		versionEntries, err := ocfl.fs.ReadDir(ver)
		if err != nil {
			ocfl.addValidationError(E010, "cannot read version folder '%s'", ver)
			continue
			//			return errors.Wrapf(err, "cannot read dir '%s'", ver)
		}
		for _, entry := range versionEntries {
			if entry.IsDir() {
				if !slices.Contains(allowedDirs, entry.Name()) {
					ocfl.addValidationWarning(W002, "extra dir '%s' in version directory '%s'", entry.Name(), ver)
				}
			} else {
				if !slices.Contains(allowedFiles, entry.Name()) {
					ocfl.addValidationError(E015, "extra file '%s' in version directory '%s'", entry.Name(), ver)
				}
			}
		}
	}

	for key := 0; key < len(versionStrings)-1; key++ {
		v1 := versionStrings[key]
		vi1, ok := versionInventories[v1]
		if !ok {
			ocfl.addValidationWarning(W010, "no inventory for version %s", versionStrings[key])
			continue
			// return errors.Errorf("no inventory for version %s", versionStrings[key])
		}
		v2 := versionStrings[key+1]
		vi2, ok := versionInventories[v2]
		if !ok {
			ocfl.addValidationWarning(W000, "no inventory for version %s", versionStrings[key+1])
			continue
		}
		if !SpecIsLessOrEqual(vi1.GetSpec(), vi2.GetSpec()) {
			ocfl.addValidationError(E103, "spec in version %s (%s) greater than spec in version %s (%s)", v1, vi1.GetSpec(), v2, vi2.GetSpec())
		}
	}

	id := ocfl.i.GetID()
	digestAlg := ocfl.i.GetDigestAlgorithm()
	versions := ocfl.i.GetVersions()
	for ver, i := range versionInventories {
		// check for id consistency
		if id != i.GetID() {
			ocfl.addValidationError(E037, "invalid id - root inventory id %s != version %s inventory id %s", id, ver, i.GetID())
		}
		if i.GetHead() != "" && i.GetHead() != ver {
			ocfl.addValidationError(E040, "wrong head %s in manifest for version '%s'", i.GetHead(), ver)
		}

		if i.GetDigestAlgorithm() != digestAlg {
			ocfl.addValidationError(W004, "different digest algorithm '%s' in version '%s'", i.GetDigestAlgorithm(), ver)
		}

		for verVer, verVersion := range i.GetVersions() {
			testV, ok := versions[verVer]
			if !ok {
				ocfl.addValidationError(E066, "version '%s' in version folder '%s' not in object root manifest", ver, verVer)
			}
			if !testV.EqualState(verVersion) {
				ocfl.addValidationError(E066, "version '%s' in version folder '%s' not equal to version in object root manifest", ver, verVer)
			}
			if !testV.EqualMeta(verVersion) {
				ocfl.addValidationError(W011, "version '%s' in version folder '%s' has different metadata as version in object root manifest", ver, verVer)
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
				ocfl.addValidationError(E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}

	rootManifestFiles := ocfl.i.GetFilesFlat()
	for _, manifestFile := range rootManifestFiles {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			ocfl.addValidationError(E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
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
		if ocfl.i.VersionLessOrEqual(latestVersion, objectContentVersion) {
			latestVersion = objectContentVersion
		}
		// check version inventories
		for inventoryVersion, versionInventory := range versionInventories {
			if versionInventory.VersionLessOrEqual(objectContentVersion, inventoryVersion) {
				versionManifestFiles := versionInventory.GetFilesFlat()
				for _, objectContentVersionFile := range objectContentVersionFiles {
					// check all inventories which are less in version
					if !slices.Contains(versionManifestFiles, objectContentVersionFile) {
						ocfl.addValidationError(E023, "file '%s' not in manifest version %s", objectContentVersionFile, inventoryVersion)
					}
				}
			}
		}
		rootVersion := ocfl.i.GetHead()
		if ocfl.i.VersionLessOrEqual(objectContentVersion, rootVersion) {
			rootManifestFiles := ocfl.i.GetFilesFlat()
			for _, objectContentVersionFile := range objectContentVersionFiles {
				// check all inventories which are less in version
				if !slices.Contains(rootManifestFiles, objectContentVersionFile) {
					ocfl.addValidationError(E023, "file '%s' not in manifest version %s", objectContentVersionFile, rootVersion)
				}
			}
		}
	}

	return nil
}

func (ocfl *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//ocfl.fs
	ocfl.logger.Infof("object %s with ocfl version %s found", ocfl.GetID(), ocfl.GetVersion())
	// check folders
	versions := ocfl.i.GetVersionStrings()

	// check for allowed files and directories
	allowedDirs := append(versions, "logs", "extensions")
	versionCounter := 0
	entries, err := ocfl.fs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !slices.Contains(allowedDirs, entry.Name()) {
				ocfl.addValidationError(E001, "invalid directory '%s' found", entry.Name())
				// could it be a version folder?
				if _, err := strconv.Atoi(strings.TrimLeft(entry.Name(), "v0")); err == nil {
					if err2 := ocfl.checkVersionFolder(entry.Name()); err2 == nil {
						ocfl.addValidationError(E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			if slices.Contains(versions, entry.Name()) {
				err := ocfl.checkVersionFolder(entry.Name())
				if err != nil {
					return errors.WithStack(err)
				}
				versionCounter++
			}
		} else {
			if !allowedFilesRegexp.MatchString(entry.Name()) {
				ocfl.addValidationError(E001, "invalid file %s found", entry.Name())
			}
		}
	}

	if versionCounter != len(versions) {
		ocfl.addValidationError(E010, "number of versions in inventory (%v) does not fit versions in filesystem (%v)", versionCounter, len(versions))
	}

	if err := ocfl.checkFilesAndVersions(); err != nil {
		return errors.WithStack(err)
	}

	dAlgs := []checksum.DigestAlgorithm{ocfl.i.GetDigestAlgorithm()}
	dAlgs = append(dAlgs, ocfl.i.GetFixityDigestAlgorithm()...)
	return nil
}

// create checksums of all content files
func (ocfl *ObjectBase) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	// get all possible digest algs
	digestAlgorithms, err := ocfl.getAllDigests()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get digests")
	}

	result := map[checksum.DigestAlgorithm]map[string][]string{}
	checksumWriter := checksum.NewChecksumWriter(digestAlgorithms)
	versions := ocfl.i.GetVersionStrings()
	for _, version := range versions {
		if err := ocfl.fs.WalkDir(
			//fmt.Sprintf("%s/%s", version, ocfl.i.GetContentDir()),
			version,
			func(path string, d fs.DirEntry, err error) error {
				//ocfl.logger.Debug(path)
				if d.IsDir() {
					return nil
				}
				fp, err := ocfl.fs.Open(path)
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
			return nil, errors.Wrapf(err, "cannot walk content dir %s", ocfl.i.GetContentDir())
		}
	}
	return result, nil
}

var objectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

// helper functions

func (ocfl *ObjectBase) getVersionInventories() (map[string]Inventory, error) {
	if ocfl.versionInventories != nil {
		return ocfl.versionInventories, nil
	}

	versionStrings := ocfl.i.GetVersionStrings()

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b string) bool {
		return ocfl.i.VersionLessOrEqual(a, b) && a != b
	})
	versionInventories := map[string]Inventory{}
	for _, ver := range versionStrings {
		vi, err := ocfl.LoadInventory(ver)
		if err != nil {
			if ocfl.fs.IsNotExist(err) {
				ocfl.addValidationWarning(W010, "no inventory for version %s", ver)
				continue
			}
			return nil, errors.Wrapf(err, "cannot load inventory from folder '%s'", ver)
		}
		versionInventories[ver] = vi
	}
	ocfl.versionInventories = versionInventories
	return ocfl.versionInventories, nil
}

func (ocfl *ObjectBase) getAllDigests() ([]checksum.DigestAlgorithm, error) {
	versionInventories, err := ocfl.getVersionInventories()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get version inventories")
	}
	allDigestAlgs := []checksum.DigestAlgorithm{ocfl.i.GetDigestAlgorithm()}
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
