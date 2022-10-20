package ocfl

import (
	"bytes"
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/object"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//const VERSION = "1.0"

//var objectConformanceDeclaration = fmt.Sprintf("0=ocfl_object_%s", VERSION)

type ObjectBase struct {
	ctx     context.Context
	fs      OCFLFS
	i       Inventory
	changed bool
	logger  *logging.Logger
	version OCFLVersion
	path    object.Path
}

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, id string, logger *logging.Logger) (*ObjectBase, error) {
	ocfl := &ObjectBase{ctx: ctx, fs: fs, version: defaultVersion, logger: logger}
	if id != "" {
		dPath, err := object.NewDefaultPath()
		if err != nil {
			return nil, errors.Wrap(err, "cannot initialize default path")
		}
		// create initial filesystem structure for new object
		if err := ocfl.New(id, dPath); err == nil {
			return ocfl, nil
		}
	}
	// load the object
	if err := ocfl.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	if id != "" && ocfl.GetID() != id {
		return nil, fmt.Errorf("id mismatch. %s != %s", id, ocfl.GetID())
	}
	return ocfl, nil
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

func (ocfl *ObjectBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	addValidationErrors(ocfl.ctx, GetValidationError(ocfl.version, errno).AppendDescription(format, a...))
}

func (ocfl *ObjectBase) LoadInventory() (Inventory, error) {
	return ocfl.LoadInventoryFolder(".")
}

// LoadInventory loads inventory from existing Object
func (ocfl *ObjectBase) LoadInventoryFolder(folder string) (Inventory, error) {
	inventory, err := NewInventory(ocfl.ctx, ocfl, "", ocfl.version, ocfl.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}

	// load inventory file
	filename := filepath.ToSlash(filepath.Join(folder, "inventory.json"))
	iFp, err := ocfl.fs.Open(filename)
	if ocfl.fs.IsNotExist(err) {
		ocfl.addValidationError(E063, "no inventory file in \"%s\"", ocfl.fs.String())
		return inventory, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", filename)
	}
	// read inventory into memory
	inventoryBytes, err := io.ReadAll(iFp)
	iFp.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %s", filename)
	}
	if err := json.Unmarshal(inventoryBytes, inventory); err != nil {
		return nil, errors.Wrap(err, "cannot marshal inventory.json")
	}
	digest := inventory.GetDigestAlgorithm()

	// check digest for inventory
	digestPath := fmt.Sprintf("%s.%s", filename, digest)
	digestBytes, err := fs.ReadFile(ocfl.fs, digestPath)
	if err != nil {
		ocfl.addValidationError(E058, "cannot read %s: %v", digestPath, err)
	} else {
		digestString := strings.TrimSpace(string(digestBytes))
		if !strings.HasSuffix(digestString, " inventory.json") {
			ocfl.addValidationError(E061, "no suffix \" inventory.json\" in %s", digestPath)
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
	return inventory, inventory.Init()
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
	checksumBytes := h.Sum(jsonBytes)
	checksumString := fmt.Sprintf("%x %s", checksumBytes, iFileName)
	iWriter, err := ocfl.fs.Create(iFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return errors.Wrap(err, "cannot write to inventory.json")
	}
	iFileName = fmt.Sprintf("%s/inventory.json", ocfl.i.GetVersion())
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
	csFileName = fmt.Sprintf("%s/inventory.json.%s", ocfl.i.GetVersion(), string(ocfl.i.GetDigestAlgorithm()))
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
	configFile := fmt.Sprintf("extensions/%s/config.json", ocfl.path.Name())
	extConfig, err := ocfl.fs.Create(configFile)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", configFile)
	}
	defer extConfig.Close()
	if err := ocfl.path.WriteConfig(extConfig); err != nil {
		return errors.Wrap(err, "cannot write config")
	}
	return nil
}
func (ocfl *ObjectBase) New(id string, path object.Path) error {
	ocfl.logger.Debugf("%s", id)

	ocfl.path = path
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

	ocfl.i, err = NewInventory(ocfl.ctx, ocfl, id, ocfl.version, ocfl.logger)
	return nil
}

func (ocfl *ObjectBase) Load() (err error) {
	// first check whether object already exists
	ocfl.version, err = GetObjectVersion(ocfl.ctx, ocfl.fs)
	if err != nil {
		return err
	}
	// read path from extension folder...
	exts, err := ocfl.fs.ReadDir("extensions")
	if err != nil {
		// if directory does not exists - no problem
		if err != fs.ErrNotExist {
			return errors.Wrap(err, "cannot read extensions folder")
		}
		exts = []fs.DirEntry{}
	}
	for _, extFolder := range exts {
		extConfig := fmt.Sprintf("extensions/%s/config.json", extFolder.Name())
		configReader, err := ocfl.fs.Open(extConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot open %s for reading", extConfig)
		}
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, configReader); err != nil {
			return errors.Wrapf(err, "cannot read %s", extConfig)
		}
		if ocfl.path, err = object.NewPath(buf.Bytes()); err != nil {
			ocfl.logger.Warningf("%s not a storage layout: %v", extConfig, err)
			continue
		}
	}
	if ocfl.path == nil {
		// ...or set to default
		if ocfl.path, err = object.NewDefaultPath(); err != nil {
			return errors.Wrap(err, "cannot initiate default storage layout")
		}
	}

	// load the inventory
	if ocfl.i, err = ocfl.LoadInventory(); err != nil {
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

func (ocfl *ObjectBase) AddFolder(fsys fs.FS) error {
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		// directory not interesting
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		checksum, err := checksum.Checksum(file, checksum.DigestSHA512)
		if err != nil {
			return errors.Wrapf(err, "cannot create checksum of %s", path)
		}
		if _, err := file.Seek(0, 0); err != nil {
			panic(err)
		}
		if err := ocfl.AddFile(strings.Trim(filepath.ToSlash(path), "/"), file, checksum); err != nil {
			return errors.Wrapf(err, "cannot add file %s", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (ocfl *ObjectBase) AddFile(virtualFilename string, reader io.Reader, digest string) error {
	virtualFilename = filepath.ToSlash(virtualFilename)
	ocfl.logger.Debugf("adding %s [%s]", virtualFilename, digest)

	if !ocfl.i.IsWriteable() {
		return errors.New("ocfl not writeable")
	}

	// if file is already there we do nothing
	dup, err := ocfl.i.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for %s [%s]", virtualFilename, digest)
	}
	if dup {
		ocfl.logger.Debugf("%s [%s] is a duplicate", virtualFilename, digest)
		return nil
	}
	var realFilename string
	if !ocfl.i.IsDuplicate(digest) {
		//		realFilename = ocfl.i.BuildRealname(virtualFilename)
		if realFilename, err = ocfl.path.ExecutePath(virtualFilename); err != nil {
			return errors.Wrapf(err, "cannot transform filename %s", virtualFilename)
		}
		realFilename = ocfl.i.BuildRealname(realFilename)
		writer, err := ocfl.fs.Create(realFilename)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", realFilename)
		}
		csw := checksum.NewChecksumWriter([]checksum.DigestAlgorithm{ocfl.i.GetDigestAlgorithm()})
		checksums, err := csw.Copy(writer, reader)
		if digest != "" && digest != checksums[ocfl.i.GetDigestAlgorithm()] {
			return fmt.Errorf("invalid checksum %s", digest)
		}
		digest = checksums[ocfl.i.GetDigestAlgorithm()]
		if err != nil {
			return errors.Wrapf(err, "cannot copy to %s", realFilename)
		}
	}
	if err := ocfl.i.AddFile(virtualFilename, realFilename, digest); err != nil {
		return errors.Wrapf(err, "cannot append %s/%s to inventory", realFilename, virtualFilename)
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
				ocfl.addValidationError(E015, "forbidden file \"%s\" in version directory \"%s\"", ve.Name(), version)
			}
			// else {
			//	if ve.Name() != "content" {
			//		ocfl.addValidationError(E022, "forbidden subfolder \"%s\" in version directory \"%s\"", ve.Name(), version)
			//	}
		}
	}
	return nil
}

func (ocfl *ObjectBase) checkFiles() error {
	versionContents := map[string]string{}
	versions := ocfl.i.GetVersions()
	for _, ver := range versions {
		versionContents[ver] = ocfl.i.GetContentDir()
	}
	files := ocfl.i.GetFiles()
	for _, filename := range files {
		f, err := ocfl.fs.Open(filename)
		if err != nil {
			if ocfl.fs.IsNotExist(err) {
				ocfl.addValidationError(E017, "file \"%s\" not found", filename)
			}
			return errors.Wrapf(err, "cannot open %s", filename)
		}
		f.Close()
		parts := strings.Split(filename, "/")
		if len(parts) < 3 {
			ocfl.addValidationError(E000, "invalid filename \"%s\" in manifest", filename)
		} else {
			versionContents[parts[0]] = parts[1]
		}
	}
	for ver, cont := range versionContents {
		cfiles := []string{}
		versionContent := ver + "/" + cont
		iv, err := ocfl.LoadInventoryFolder(ver)
		if err != nil {
			return errors.Wrapf(err, "cannot load inventory from folder \"%s\"", ver)
		}
		ivFiles := iv.GetFiles()
		ocfl.fs.WalkDir(
			versionContent,
			func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}
				cfiles = append(cfiles, filepath.ToSlash(path))
				return nil
			},
		)
		for _, f := range cfiles {
			if !slices.Contains(files, f) {
				ocfl.addValidationError(E023, "file \"%s/%s\" not in manifest", versionContent, f)
			}
			if !slices.Contains(ivFiles, f) {
				ocfl.addValidationError(E023, "file \"%s/%s\" not in manifest %s/inventory.json", versionContent, f, ver)
			}
		}
		// todo: deep check content
	}
	return nil
}

func (ocfl *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//ocfl.fs
	ocfl.logger.Infof("object %s with ocfl version %s found", ocfl.GetID(), ocfl.GetVersion())
	// check folders
	versions := ocfl.i.GetVersions()

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
				ocfl.addValidationError(E001, "invalid directory \"%s\" found", entry.Name())
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

	if err := ocfl.checkFiles(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

var objectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")
