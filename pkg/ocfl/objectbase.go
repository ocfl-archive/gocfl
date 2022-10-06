package ocfl

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/object"
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
	fs      OCFLFS
	i       Inventory
	changed bool
	logger  *logging.Logger
	version OCFLVersion
	path    object.Path
}

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(fs OCFLFS, defaultVersion OCFLVersion, id string, logger *logging.Logger) (*ObjectBase, error) {
	ocfl := &ObjectBase{fs: fs, version: defaultVersion, logger: logger}
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

// LoadInventory loads inventory from existing Object
func (ocfl *ObjectBase) LoadInventory() (Inventory, error) {
	// load inventory file
	iFp, err := ocfl.fs.Open("inventory.json")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", "inventory.json")
	}
	// read inventory into memory
	inventoryBytes, err := io.ReadAll(iFp)
	iFp.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %s", "inventory.json")
	}
	inventory, err := NewInventory("", ocfl.version, ocfl.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	if err := json.Unmarshal(inventoryBytes, inventory); err != nil {
		return nil, errors.Wrap(err, "cannot marshal inventory.json")
	}
	digest := inventory.GetDigestAlgorithm()

	// check digest for inventory
	digestPath := fmt.Sprintf("inventory.json.%s", digest)
	digestBytes, err := fs.ReadFile(ocfl.fs, digestPath)
	if err != nil {
		return nil, MultiError(err, GetValidationError(ocfl.version, E058))
	}
	digestString := strings.TrimSpace(string(digestBytes))
	if !strings.HasSuffix(digestString, " inventory.json") {
		return nil, GetValidationError(ocfl.version, E061)
	}
	digestString = strings.TrimSuffix(digestString, " inventory.json")
	h, err := checksum.GetHash(digest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - %s", string(digest)))
	}
	sumBytes := h.Sum(inventoryBytes)
	inventoryDigestString := fmt.Sprintf("%x", sumBytes)
	if digestString != inventoryDigestString {
		return nil, MultiError(fmt.Errorf("%s != %s", digestString, inventoryDigestString), GetValidationError(ocfl.version, E060))
	}

	return inventory, nil
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

	ocfl.i, err = NewInventory(id, ocfl.version, ocfl.logger)
	return nil
}

func (ocfl *ObjectBase) Load() error {
	// first check whether object already exists
	ver, err := ocfl.getVersion()
	if err != nil {
		return err
	}
	if ver != ocfl.version {
		return errors.New(fmt.Sprintf("version mismatch: %s != %s", ver, ocfl.version))
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

func (ocfl *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//ocfl.fs
	var multiError = []error{}
	version, err := ocfl.getVersion()
	if err != nil {
		multiError = append(multiError, errors.Wrap(err, "cannot get version"))
	}
	ocfl.logger.Infof("object with ocfl version %s found", version)
	return nil
}

var objectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

func (ocfl *ObjectBase) getVersion() (version OCFLVersion, err error) {
	files, err := ocfl.fs.ReadDir(".")
	if err != nil {
		return "", errors.Wrap(err, "cannot get files")
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := objectVersionRegexp.FindStringSubmatch(file.Name())
		if matches != nil {
			if version != "" {
				return "", errVersionMultiple
			}
			version = OCFLVersion(matches[1])
			r, err := ocfl.fs.Open(filepath.Join(file.Name()))
			if err != nil {
				return "", errors.Wrapf(err, "cannot open %s", file.Name())
			}
			cnt, err := io.ReadAll(r)
			if err != nil {
				r.Close()
				return "", errors.Wrapf(err, "cannot read %s", file.Name())
			}
			r.Close()
			if string(cnt) != fmt.Sprintf("ocfl_object_%s\n", version) {
				return "", GetValidationError(ocfl.version, E007)
			}
		}
	}
	if version == "" {
		return "", GetValidationError(ocfl.version, E003)
	}
	ocfl.version = version
	return version, nil
}
