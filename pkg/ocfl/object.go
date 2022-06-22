package ocfl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

const VERSION = "1.0"

var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

type OCFLObject struct {
	fs         OCFLFS
	pathPrefix string
	i          *Inventory
	changed    bool
	logger     *logging.Logger
}

// NewOCFL creates an empty OCFL structure
func NewOCFLObject(fs OCFLFS, pathPrefix, id string, logger *logging.Logger) (*OCFLObject, error) {
	if pathPrefix == "." {
		pathPrefix = ""
	}
	// no / prefix, but / suffix
	pathPrefix = strings.Trim(pathPrefix, "/") + "/"
	ocfl := &OCFLObject{fs: fs, logger: logger, pathPrefix: pathPrefix}

	if err := ocfl.Init(id); err != nil {
		return nil, emperror.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")
var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

// LoadInventory loads inventory from existing OCFLFS
func (ocfl *OCFLObject) LoadInventory(folder string) (*Inventory, error) {
	if folder == "." {
		folder = ""
	}
	ocfl.logger.Debugf("%s", folder)
	entries, err := ocfl.fs.ReadDir(ocfl.pathPrefix + folder)
	var inventoryDigest = map[checksum.DigestAlgorithm]string{}
	for _, entry := range entries {
		name := entry.Name()
		if matches := inventoryDigestRegexp.FindStringSubmatch(name); matches != nil {
			var digest checksum.DigestAlgorithm
			// OCFL supports only SHA512 (SHA256) for inventory digest
			switch matches[1] {
			case string(checksum.DigestSHA512):
				digest = checksum.DigestSHA512
			case string(checksum.DigestSHA256):
				digest = checksum.DigestSHA256
			default:
				return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - %s", name))
			}
			digestBytes, err := fs.ReadFile(ocfl.fs, ocfl.pathPrefix+name)
			if err != nil {
				return nil, emperror.Wrapf(err, "cannot read digest file %s", name)
			}
			inventoryDigest[digest] = strings.ToLower(strings.TrimSpace(string(digestBytes)))
			break
		} else {
			continue
		}
	}
	if len(inventoryDigest) == 0 {
		return nil, errors.New(fmt.Sprintf("cannot find digest file for %s", "inventory.json"))
	}
	iFp, err := ocfl.fs.Open(ocfl.pathPrefix + "inventory.json")
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot open %s", "inventory.json")
	}
	// read inventory into memory
	inventoryBytes, err := io.ReadAll(iFp)
	iFp.Close()
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot read %s", "inventory.json")
	}
	// checksum test
	for digest, hexString := range inventoryDigest {
		h, err := checksum.GetHash(digest)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - %s", string(digest)))
		}
		sumBytes := h.Sum(inventoryBytes)
		if hexString != fmt.Sprintf("%x", sumBytes) {
			return nil, errors.New(fmt.Sprintf("%s checksum of inventory.json failed", string(digest)))
		}
	}
	var inventory = &Inventory{logger: ocfl.logger}
	if err := json.Unmarshal(inventoryBytes, inventory); err != nil {
		return nil, emperror.Wrap(err, "cannot marshal inventory.json")
	}
	return inventory, nil
}

func (ocfl *OCFLObject) StoreInventory() error {
	ocfl.logger.Debug()

	if !ocfl.i.IsWriteable() {
		return errors.New("inventory not updated")
	}
	jsonBytes, err := json.MarshalIndent(ocfl.i, "", "   ")
	if err != nil {
		return emperror.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(ocfl.i.GetDigestAlgorithm())
	if err != nil {
		return emperror.Wrapf(err, "invalid digest algorithm %s", string(ocfl.i.GetDigestAlgorithm()))
	}
	checksumBytes := h.Sum(jsonBytes)
	checksumString := fmt.Sprintf("%x", checksumBytes)
	iFileName := "inventory.json"
	iWriter, err := ocfl.fs.Create(ocfl.pathPrefix + iFileName)
	if err != nil {
		return emperror.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return emperror.Wrap(err, "cannot write to inventory.json")
	}
	iFileName = fmt.Sprintf("%s/inventory.json", ocfl.i.GetVersion())
	iWriter, err = ocfl.fs.Create(ocfl.pathPrefix + iFileName)
	if err != nil {
		return emperror.Wrap(err, "cannot create inventory.json")
	}
	if _, err := iWriter.Write(jsonBytes); err != nil {
		return emperror.Wrap(err, "cannot write to inventory.json")
	}
	csFileName := fmt.Sprintf("inventory.json.%s", string(ocfl.i.GetDigestAlgorithm()))
	iCSWriter, err := ocfl.fs.Create(ocfl.pathPrefix + csFileName)
	if err != nil {
		return emperror.Wrapf(err, "cannot create %s", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return emperror.Wrapf(err, "cannot write to %s", csFileName)
	}
	csFileName = fmt.Sprintf("%s/inventory.json.%s", ocfl.i.GetVersion(), string(ocfl.i.GetDigestAlgorithm()))
	iCSWriter, err = ocfl.fs.Create(ocfl.pathPrefix + csFileName)
	if err != nil {
		return emperror.Wrapf(err, "cannot create %s", csFileName)
	}
	if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
		return emperror.Wrapf(err, "cannot write to %s", csFileName)
	}
	return nil
}

func (ocfl *OCFLObject) Init(id string) error {
	ocfl.logger.Debugf("%s", id)

	// first check whether ocfl is not empty
	fp, err := ocfl.fs.Open(ocfl.pathPrefix + rootConformanceDeclaration)
	if err != nil {
		if err != fs.ErrNotExist {
			return emperror.Wrap(err, "cannot initialize OCFL layout")
		}
		_, err := ocfl.fs.Create(ocfl.pathPrefix + rootConformanceDeclaration)
		if err != nil {
			return emperror.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
		/*
			if err := fp.Close(); err != nil {
				return emperror.Wrapf(err, "cannot close %s", rootConformanceDeclaration)
			}

		*/
		ocfl.i, err = NewInventory(id, ocfl.logger)
	} else {
		if err := fp.Close(); err != nil {
			return emperror.Wrapf(err, "cannot close %s", rootConformanceDeclaration)
		}
		// now load the inventory
		if ocfl.i, err = ocfl.LoadInventory("."); err != nil {
			return emperror.Wrap(err, "cannot load inventory.json of root")
		}
	}
	return nil
}

func (ocfl *OCFLObject) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return ocfl.i.GetDigestAlgorithm()
}
func (ocfl *OCFLObject) Close() error {
	ocfl.logger.Debug()
	if ocfl.i.IsWriteable() {
		if err := ocfl.i.Clean(); err != nil {
			return emperror.Wrap(err, "cannot clean inventory")
		}
		if err := ocfl.StoreInventory(); err != nil {
			return emperror.Wrap(err, "cannot store inventory")
		}
	}
	return nil
}

func (ocfl *OCFLObject) StartUpdate(msg string, UserName string, UserAddress string) error {
	ocfl.logger.Debugf("%s / %s / %s", msg, UserName, UserAddress)

	if ocfl.i.IsWriteable() {
		return errors.New("ocfl already writeable")
	}
	if err := ocfl.i.NewVersion(msg, UserName, UserAddress); err != nil {
		return emperror.Wrap(err, "cannot create new ocfl version")
	}
	return nil
}

func (ocfl *OCFLObject) AddFile(virtualFilename string, reader io.Reader, digest string) error {
	virtualFilename = filepath.ToSlash(virtualFilename)
	ocfl.logger.Debugf("%s [%s]", virtualFilename, digest)

	if !ocfl.i.IsWriteable() {
		return errors.New("ocfl not writeable")
	}

	// if file is already there we do nothing
	dup, err := ocfl.i.IsDuplicate(virtualFilename, digest)
	if err != nil {
		return emperror.Wrapf(err, "cannot check duplicate for %s [%s]", virtualFilename, digest)
	}
	if dup {
		ocfl.logger.Debugf("%s [%s] is a duplicate", virtualFilename, digest)
		return nil
	}
	realFilename := ocfl.i.BuildRealname(virtualFilename)
	writer, err := ocfl.fs.Create(ocfl.pathPrefix + realFilename)
	if err != nil {
		return emperror.Wrapf(err, "cannot create %s", realFilename)
	}
	csw := checksum.NewChecksumWriter([]checksum.DigestAlgorithm{ocfl.i.GetDigestAlgorithm()})
	checksums, err := csw.Copy(writer, reader)
	if err != nil {
		return emperror.Wrapf(err, "cannot copy to %s", realFilename)
	}
	if err := ocfl.i.AddFile(virtualFilename, realFilename, checksums[ocfl.i.GetDigestAlgorithm()]); err != nil {
		return emperror.Wrapf(err, "cannot append %s/%s to inventory", realFilename, virtualFilename)
	}
	return nil
}
