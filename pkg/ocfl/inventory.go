package ocfl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	InventoryType    = "https://ocfl.io/1.0/spec/#inventory"
	DigestAlg        = checksum.DigestSHA512
	ContentDirectory = "content"
)

type OCFLTime struct{ time.Time }

func (t *OCFLTime) MarshalJSON() ([]byte, error) {
	tstr := t.Format(time.RFC3339)
	return json.Marshal(tstr)
}

func (t *OCFLTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return emperror.Wrapf(err, "cannot unmarshal string of %s", string(data))
	}
	tt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return emperror.Wrapf(err, "cannot parse %s", string(data))
	}
	t.Time = tt
	return nil
}

type User struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Version struct {
	Created OCFLTime            `json:"created"`
	Message string              `json:"message"`
	State   map[string][]string `json:"state"`
	User    User                `json:"user"`
}

type Inventory struct {
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

func NewInventory(id string, logger *logging.Logger) (*Inventory, error) {
	i := &Inventory{
		Id:               id,
		Type:             InventoryType,
		DigestAlgorithm:  DigestAlg,
		Head:             "",
		ContentDirectory: ContentDirectory,
		Manifest:         map[string][]string{},
		Versions:         map[string]*Version{},
		Fixity:           nil,
		logger:           logger,
	}
	return i, nil
}

func (i *Inventory) GetID() string                                { return i.Id }
func (i *Inventory) GetContentDirectory() string                  { return i.ContentDirectory }
func (i *Inventory) GetVersion() string                           { return i.Head }
func (i *Inventory) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }
func (i *Inventory) IsWriteable() bool                            { return i.writeable }
func (i *Inventory) IsModified() bool                             { return i.modified }
func (i *Inventory) BuildRealname(virtualFilename string) string {
	return fmt.Sprintf("%s/%s/%s", i.GetVersion(), i.GetContentDirectory(), FixFilename(filepath.ToSlash(virtualFilename)))
}

func (i *Inventory) NewVersion(msg, UserName, UserAddress string) error {
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
			return emperror.Wrapf(err, "cannot determine head of Object - %s", vStr)
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

func (i *Inventory) getLastVersion() (string, error) {
	versions := []int{}
	for ver, _ := range i.Versions {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return "", errors.New(fmt.Sprintf("invalid version in inventory - %s", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", emperror.Wrapf(err, "cannot convert version number to int - %s", matches[1])
		}
		versions = append(versions, versionInt)
	}

	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	return fmt.Sprintf("v%d", lastVersion), nil
}
func (i *Inventory) IsDuplicate(checksum string) bool {
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
func (i *Inventory) AlreadyExists(virtualFilename, checksum string) (bool, error) {
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
			return false, emperror.Wrapf(err, "cannot convert version number to int - %s", matches[1])
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

func (i *Inventory) IsUpdate(virtualFilename, checksum string) (bool, error) {
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
			return false, emperror.Wrapf(err, "cannot convert version number to int - %s", matches[1])
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

func (i *Inventory) DeleteFile(virtualFilename string) error {
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
func (i *Inventory) Rename(oldVirtualFilename, newVirtualFilename string) error {
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

func (i *Inventory) AddFile(virtualFilename string, realFilename string, checksum string) error {
	i.logger.Debugf("%s - %s [%s]", virtualFilename, realFilename, checksum)
	checksum = strings.ToLower(checksum) // paranoia
	if _, ok := i.Manifest[checksum]; !ok {
		i.Manifest[checksum] = []string{}
	}
	dup, err := i.AlreadyExists(virtualFilename, checksum)
	if err != nil {
		return emperror.Wrapf(err, "cannot check for duplicate of %s [%s]", virtualFilename, checksum)
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
		return emperror.Wrapf(err, "cannot check for update of %s [%s]", virtualFilename, checksum)
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
func (i *Inventory) Clean() error {
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
		return emperror.Wrap(err, "cannot get last version")
	}
	i.Head = lastVersion
	return nil
}
