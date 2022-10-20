package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"golang.org/x/exp/slices"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type InventoryBase struct {
	ctx              context.Context
	object           Object                                           `json:"-"`
	modified         bool                                             `json:"-"`
	writeable        bool                                             `json:"-"`
	paddingLength    int                                              `json:"-"`
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

func NewInventoryBase(
	ctx context.Context,
	object Object,
	id string,
	objectType *url.URL,
	digestAlg checksum.DigestAlgorithm,
	contentDir string,
	logger *logging.Logger) (*InventoryBase, error) {
	i := &InventoryBase{
		ctx:              ctx,
		object:           object,
		Id:               id,
		paddingLength:    0,
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
func (i *InventoryBase) Init() (err error) {
	if err := i.check(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (i *InventoryBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	addValidationErrors(i.ctx, GetValidationError(i.object.GetVersion(), errno).AppendDescription(format, a...))
}
func (i *InventoryBase) GetID() string { return i.Id }

func (i *InventoryBase) GetContentDir() string {
	return i.ContentDirectory
}

func (i *InventoryBase) GetContentDirectory() string                  { return i.ContentDirectory }
func (i *InventoryBase) GetVersion() string                           { return i.Head }
func (i *InventoryBase) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }
func (i *InventoryBase) IsWriteable() bool                            { return i.writeable }
func (i *InventoryBase) IsModified() bool                             { return i.modified }

func (i *InventoryBase) GetVersions() []string {
	var versions = []string{}
	for version, _ := range i.Versions {
		versions = append(versions, version)
	}
	return versions
}

var versionZeroRegexp = regexp.MustCompile("^v0[0-9]+$")
var versionNoZeroRegexp = regexp.MustCompile("^v[1-9][0-9]*$")

func (i *InventoryBase) check() error {

	if err := i.checkVersions(); err != nil {
		return errors.WithStack(err)
	}
	if i.Id == "" {
		i.addValidationError(E036, "invalid field \"id\" for object")
	}
	if i.Head == "" {
		i.addValidationError(E036, "invalid field \"head\" for object")
	}
	if i.Type == "" {
		i.addValidationError(E036, "invalid field \"type\" for object")
	}
	if i.DigestAlgorithm == "" {
		i.addValidationError(E036, "invalid field \"digestAlgorithm\" for object")
	}
	if slices.Contains([]string{"", ".", ".."}, i.ContentDirectory) || strings.Contains(i.ContentDirectory, "/") {
		i.addValidationError(E017, "invalid content directory \"%s\"", i.ContentDirectory)
	}
	return nil
}

func (i *InventoryBase) checkVersions() error {
	var paddingLength int = -1
	var versions = []int{}
	if len(i.Versions) == 0 {
		i.addValidationError(E008, "length of version is 0")
	}
	for version, _ := range i.Versions {
		vInt, err := strconv.Atoi(strings.TrimLeft(version, "v0"))
		if err != nil {
			i.addValidationError(E104, "invalid version format %s", version)
			continue
		}
		versions = append(versions, vInt)
		if versionZeroRegexp.MatchString(version) {
			if paddingLength == -1 {
				paddingLength = len(version) - 2
			} else {
				if paddingLength != len(version)-2 {
					//i.addValidationError(E011, "invalid version padding %s", version)
					i.addValidationError(E012, "invalid version padding %s", version)
					i.addValidationError(E013, "invalid version padding %s", version)
				}
			}
		} else {
			if versionNoZeroRegexp.MatchString(version) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						i.addValidationError(E011, "invalid version padding %s", version)
						i.addValidationError(E012, "invalid version padding %s", version)
						i.addValidationError(E013, "invalid version padding %s", version)
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				i.addValidationError(E104, "invalid version format %s", version)
			}
		}
	}
	slices.Sort(versions)
	for key, val := range versions {
		if key != val-1 {
			i.addValidationError(E010, "invalid version sequence %v", versions)
			break
		}
	}
	i.paddingLength = paddingLength
	return nil
}

func (i *InventoryBase) GetFiles() []string {
	var result = []string{}
	for _, files := range i.Manifest {
		for _, filename := range files {
			parts := strings.Split(filename, "/")
			if len(parts) < 3 {
				i.addValidationError(E000, "invalid filepath in manifest \"%s\"", filename)
			}
			if parts[1] != i.GetContentDir() {
				i.addValidationError(E019, "invalid content directory \"%s\" in \"%s\"", parts[1], filename)
			}
			result = append(result, filename)
		}
	}
	return result
}

func (i *InventoryBase) BuildRealname(virtualFilename string) string {
	//	return fmt.Sprintf("%s/%s/%s", i.GetVersion(), i.GetContentDirectory(), FixFilename(filepath.ToSlash(virtualFilename)))
	return fmt.Sprintf("%s/%s/%s", i.GetVersion(), i.GetContentDirectory(), filepath.ToSlash(virtualFilename))
}

func (i *InventoryBase) NewVersion(msg, UserName, UserAddress string) error {
	if i.IsWriteable() {
		return errors.New(fmt.Sprintf("version %s already writeable", i.GetVersion()))
	}
	lastHead := i.Head
	if lastHead == "" {
		if i.paddingLength <= 0 {
			i.Head = "v1"
		} else {
			i.Head = fmt.Sprintf(fmt.Sprintf("v0%%0%dd", i.paddingLength), 1)
		}
	} else {
		vStr := strings.TrimLeft(strings.ToLower(i.Head), "v0")
		v, err := strconv.Atoi(vStr)
		if err != nil {
			return errors.Wrapf(err, "cannot determine head of ObjectBase - %s", vStr)
		}

		if i.paddingLength <= 0 {
			i.Head = fmt.Sprintf("v%d", v+1)
		} else {
			i.Head = fmt.Sprintf(fmt.Sprintf("v0%%0%dd", i.paddingLength), v+1)
		}
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
