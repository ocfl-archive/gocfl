package ocfl

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type InventoryBase struct {
	modified         bool                                             `json:"-"`
	writeable        bool                                             `json:"-"`
	Id               string                                           `json:"id"`
	Type             string                                           `json:"type"`
	DigestAlgorithm  checksum.DigestAlgorithm                         `json:"digestAlgorithm"`
	Head             string                                           `json:"head"`
	ContentDirectory string                                           `json:"contentDirectory,omitempty"`
	Manifest         map[string][]string                              `json:"manifest"`
	Versions         map[string]*Version                              `json:"versions"`
	Fixity           map[checksum.DigestAlgorithm]map[string][]string `json:"fixity,omitempty"`
	logger           *logging.Logger
}

func NewInventoryBase(id string, objectType *url.URL, digestAlg checksum.DigestAlgorithm, contentDir string, logger *logging.Logger) (*InventoryBase, error) {
	i := &InventoryBase{
		Id:               id,
		Type:             objectType.String(),
		DigestAlgorithm:  digestAlg,
		Head:             "",
		ContentDirectory: contentDir,
		Manifest:         map[string][]string{},
		Versions:         map[string]*Version{},
		Fixity:           nil,
		logger:           logger,
	}
	return i, nil
}

func (i *InventoryBase) GetID() string                                { return i.Id }
func (i *InventoryBase) GetContentDirectory() string                  { return i.ContentDirectory }
func (i *InventoryBase) GetVersion() string                           { return i.Head }
func (i *InventoryBase) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }
func (i *InventoryBase) IsWriteable() bool                            { return i.writeable }
func (i *InventoryBase) IsModified() bool                             { return i.modified }
func (i *InventoryBase) BuildRealname(virtualFilename string) string {
	return fmt.Sprintf("%s/%s/%s", i.GetVersion(), i.GetContentDirectory(), FixFilename(filepath.ToSlash(virtualFilename)))
}

func (i *InventoryBase) NewVersion(msg, UserName, UserAddress string) error {
	if i.IsWriteable() {
		return errors.New(fmt.Sprintf("version %s already writeable", i.GetVersion()))
	}
	lastHead := i.Head
	if lastHead == "" {
		i.Head = "v1"
	} else {
		vStr := strings.TrimPrefix(strings.ToLower(i.Head), "v")
		v, err := strconv.Atoi(vStr)
		if err != nil {
			return errors.Wrapf(err, "cannot determine head of ObjectBase - %s", vStr)
		}
		i.Head = fmt.Sprintf("v%d", v+1)
	}
	i.Versions[i.Head] = &Version{
		Created: OCFLTime{time.Now()},
		Message: msg,
		State:   map[string][]string{},
		User: User{
			Name:    UserName,
			Address: UserAddress,
		},
	}
	// copy last state...
	if lastHead != "" {
		copyMapStringSlice(i.Versions[i.Head].State, i.Versions[lastHead].State)
	}
	i.writeable = true
	return nil
}

var vRegexp *regexp.Regexp = regexp.MustCompile("^v(\\d+)$")

func (i *InventoryBase) getLastVersion() (string, error) {
	versions := []int{}
	for ver, _ := range i.Versions {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return "", errors.New(fmt.Sprintf("invalid version in inventory - %s", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", errors.Wrapf(err, "cannot convert version number to int - %s", matches[1])
		}
		versions = append(versions, versionInt)
	}

	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	return fmt.Sprintf("v%d", lastVersion), nil
}
func (i *InventoryBase) IsDuplicate(checksum string) bool {
	// not necessary but fast...
	if checksum == "" {
		return false
	}
	for cs, _ := range i.Manifest {
		if cs == checksum {
			return true
		}
	}
	return false
}
func (i *InventoryBase) AlreadyExists(virtualFilename, checksum string) (bool, error) {
	i.logger.Debugf("%s [%s]", virtualFilename, checksum)
	if checksum == "" {
		i.logger.Debugf("%s - duplicate %v", virtualFilename, false)
		return false, nil
	}

	// first get checksum of last version of a file
	cs := map[string]string{}
	for ver, version := range i.Versions {
		for checksum, filenames := range version.State {
			found := false
			for _, filename := range filenames {
				if filename == virtualFilename {
					cs[ver] = checksum
					found = true
				}
			}
			if found {
				break
			}
		}
	}
	if len(cs) == 0 {
		i.logger.Debugf("%s - duplicate %v", virtualFilename, false)
		return false, nil
	}
	versions := []int{}

	for ver, _ := range cs {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return false, errors.New(fmt.Sprintf("invalid version in inventory - %s", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return false, errors.Wrapf(err, "cannot convert version number to int - %s", matches[1])
		}
		versions = append(versions, versionInt)
	}
	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	lastChecksum, ok := cs[fmt.Sprintf("v%d", lastVersion)]
	if !ok {
		return false, errors.New(fmt.Sprintf("could not get checksum for v%d", lastVersion))
	}
	i.logger.Debugf("%s - duplicate %v", virtualFilename, lastChecksum == checksum)
	return lastChecksum == checksum, nil
}

func (i *InventoryBase) IsUpdate(virtualFilename, checksum string) (bool, error) {
	i.logger.Debugf("%s [%s]", virtualFilename, checksum)
	if checksum == "" {
		i.logger.Debugf("%s - update %v", virtualFilename, false)
		return false, nil
	}

	// first get checksum of last version of a file
	cs := map[string]string{}
	for ver, version := range i.Versions {
		for checksum, filenames := range version.State {
			found := false
			for _, filename := range filenames {
				if filename == virtualFilename {
					cs[ver] = checksum
					found = true
				}
			}
			if found {
				break
			}
		}
	}
	if len(cs) == 0 {
		i.logger.Debugf("%s - update %v", virtualFilename, false)
		return false, nil
	}
	versions := []int{}

	for ver, _ := range cs {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return false, errors.New(fmt.Sprintf("invalid version in inventory - %s", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return false, errors.Wrapf(err, "cannot convert version number to int - %s", matches[1])
		}
		versions = append(versions, versionInt)
	}
	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	lastChecksum, ok := cs[fmt.Sprintf("v%d", lastVersion)]
	if !ok {
		return false, errors.New(fmt.Sprintf("could not get checksum for v%d", lastVersion))
	}
	i.logger.Debugf("%s - update %v", virtualFilename, lastChecksum != checksum)
	return lastChecksum != checksum, nil
}

func (i *InventoryBase) DeleteFile(virtualFilename string) error {
	var newState = map[string][]string{}
	var found = false
	for key, vals := range i.Versions[i.GetVersion()].State {
		newState[key] = []string{}
		for _, val := range vals {
			if val == virtualFilename {
				found = true
			} else {
				newState[key] = append(newState[key], val)
			}
		}
	}
	i.Versions[i.GetVersion()].State = newState
	i.modified = found
	return nil
}
func (i *InventoryBase) Rename(oldVirtualFilename, newVirtualFilename string) error {
	var newState = map[string][]string{}
	var found = false
	for key, vals := range i.Versions[i.GetVersion()].State {
		newState[key] = []string{}
		for _, val := range vals {
			if val == oldVirtualFilename {
				found = true
				newState[key] = append(newState[key], newVirtualFilename)
			} else {
				newState[key] = append(newState[key], val)
			}
		}
	}
	i.Versions[i.GetVersion()].State = newState
	i.modified = found
	return nil
}

func (i *InventoryBase) AddFile(virtualFilename string, realFilename string, checksum string) error {
	i.logger.Debugf("%s - %s [%s]", virtualFilename, realFilename, checksum)
	checksum = strings.ToLower(checksum) // paranoia
	if _, ok := i.Manifest[checksum]; !ok {
		i.Manifest[checksum] = []string{}
	}
	dup, err := i.AlreadyExists(virtualFilename, checksum)
	if err != nil {
		return errors.Wrapf(err, "cannot check for duplicate of %s [%s]", virtualFilename, checksum)
	}
	if dup {
		i.logger.Debugf("%s is a duplicate - ignoring", virtualFilename)
		return nil
	}

	if realFilename != "" {
		i.Manifest[checksum] = append(i.Manifest[checksum], realFilename)
	}

	if _, ok := i.Versions[i.Head].State[checksum]; !ok {
		i.Versions[i.Head].State[checksum] = []string{}
	}

	upd, err := i.IsUpdate(virtualFilename, checksum)
	if err != nil {
		return errors.Wrapf(err, "cannot check for update of %s [%s]", virtualFilename, checksum)
	}
	if upd {
		i.logger.Debugf("%s is an update - removing old version", virtualFilename)
		i.DeleteFile(virtualFilename)
	}

	i.Versions[i.Head].State[checksum] = append(i.Versions[i.Head].State[checksum], virtualFilename)

	i.modified = true

	return nil
}

// clear unmodified version
func (i *InventoryBase) Clean() error {
	i.logger.Debug()
	// read only means nothing to do
	if i.IsModified() {
		return nil
	}
	// only one version. could be empty
	if i.GetVersion() == "v1" {
		return nil
	}
	i.logger.Debugf("deleting %v", i.GetVersion())
	delete(i.Versions, i.GetVersion())
	lastVersion, err := i.getLastVersion()
	if err != nil {
		return errors.Wrap(err, "cannot get last version")
	}
	i.Head = lastVersion
	return nil
}
