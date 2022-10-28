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
	versionValue     map[string]uint                                  `json:"-"`
	Id               string                                           `json:"id"`
	Type             string                                           `json:"type"`
	DigestAlgorithm  checksum.DigestAlgorithm                         `json:"digestAlgorithm"`
	Head             string                                           `json:"head"`
	ContentDirectory string                                           `json:"contentDirectory,omitempty"`
	Manifest         map[string][]string                              `json:"manifest"`
	Versions         *OCFLVersions                                    `json:"versions"`
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
		Versions:         &OCFLVersions{Versions: map[string]*Version{}},
		Fixity:           nil,
		logger:           logger,
	}
	return i, nil
}
func (i *InventoryBase) Init() (err error) {
	i.versionValue = map[string]uint{}
	for version, _ := range i.Versions.Versions {
		vInt, err := strconv.Atoi(strings.TrimLeft(version, "v0"))
		if err != nil {
			i.addValidationError(E104, "invalid version format %s", version)
			continue
		}
		i.versionValue[version] = uint(vInt)
	}

	if err := i.check(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (i *InventoryBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	addValidationErrors(i.ctx, GetValidationError(i.object.GetVersion(), errno).AppendDescription(format, a...))
}
func (i *InventoryBase) GetID() string   { return i.Id }
func (i *InventoryBase) GetHead() string { return i.Head }

func (i *InventoryBase) GetContentDir() string {
	return i.ContentDirectory
}

func (i *InventoryBase) GetContentDirectory() string                  { return i.ContentDirectory }
func (i *InventoryBase) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }
func (i *InventoryBase) GetFixityDigestAlgorithm() []checksum.DigestAlgorithm {
	result := []checksum.DigestAlgorithm{}
	if i.Fixity == nil {
		return result
	}
	for digest, _ := range i.Fixity {
		if !slices.Contains(result, digest) {
			result = append(result, digest)
		}
	}
	return result
}
func (i *InventoryBase) IsWriteable() bool { return i.writeable }
func (i *InventoryBase) IsModified() bool  { return i.modified }

func (i *InventoryBase) GetVersionStrings() []string {
	var versions = []string{}
	for version, _ := range i.Versions.Versions {
		versions = append(versions, version)
	}
	return versions
}
func (i *InventoryBase) GetVersions() map[string]*Version {
	var versions = map[string]*Version{}
	for versionStr, version := range i.Versions.Versions {
		versions[versionStr] = version
	}
	return versions
}

func (i *InventoryBase) VersionLessOrEqual(v1, v2 string) bool {
	v1Int, ok := i.versionValue[v1]
	if !ok {
		return false
	}
	v2Int, ok := i.versionValue[v2]
	if !ok {
		return false
	}
	return v1Int <= v2Int
}

var versionZeroRegexp = regexp.MustCompile("^v0[0-9]+$")
var versionNoZeroRegexp = regexp.MustCompile("^v[1-9][0-9]*$")

func (i *InventoryBase) check() error {

	if err := i.checkVersions(); err != nil {
		return errors.WithStack(err)
	}
	if err := i.checkManifest(); err != nil {
		return errors.WithStack(err)
	}
	if err := i.checkFixity(); err != nil {
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
	if !slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA512, checksum.DigestSHA256}, i.DigestAlgorithm) {
		i.addValidationError(E025, "invalid digest algorithm \"%s\"", i.DigestAlgorithm)
	}

	if slices.Contains([]string{"", ".", ".."}, i.ContentDirectory) || strings.Contains(i.ContentDirectory, "/") {
		i.addValidationError(E017, "invalid content directory \"%s\"", i.ContentDirectory)
	}

	if i.Manifest == nil || len(i.Manifest) == 0 {
		i.addValidationError(E041, "no manifest in inventory")
	}
	if i.Versions == nil || len(i.Versions.Versions) == 0 {
		i.addValidationError(E041, "no versions in inventory")
	}

	return nil
}
func (i *InventoryBase) checkManifest() error {
	digests := []string{}
	allPaths := []string{}
	for digest, paths := range i.Manifest {
		digest = strings.ToLower(digest)
		if slices.Contains(digests, digest) {
			i.addValidationError(E096, "manifest digest '%s' is duplicate", digest)
		} else {
			digests = append(digests, digest)
		}
		for _, path := range paths {
			allPaths = append(allPaths, path)
			if path[0] == '/' || path[len(path)-1] == '/' {
				i.addValidationError(E100, "invalid path \"%s\" in manifest", path)
			}
			if path == "" {
				i.addValidationError(E099, "empty path in manifest")
			}
			path2 := path
			if path[0] == '/' {
				path2 = path[1:]
			}
			elements := strings.Split(path2, "/")
			for _, element := range elements {
				if slices.Contains([]string{"", ".", ".."}, element) {
					i.addValidationError(E099, "invalid path \"%s\" in manifest", path)
				}
			}

		}

	}
	slices.Sort(allPaths)
	for j := 0; j < len(allPaths)-1; j++ {
		if strings.HasPrefix(allPaths[j+1], allPaths[j]) {
			i.addValidationError(E101, "content path '%s' is prefix or equal to '%s' in manifest", allPaths[j], allPaths[j+1])
		}
	}

	return nil
}

func (i *InventoryBase) checkFixity() error {
	for digestAlg, digestMap := range i.Fixity {
		digests := []string{}
		for digest, paths := range digestMap {
			digest = strings.ToLower(digest)
			if slices.Contains(digests, digest) {
				i.addValidationError(E097, "fixity %s digest '%s' is duplicate", digestAlg, digest)
			} else {
				digests = append(digests, digest)
			}
			// check content paths
			for _, path := range paths {
				if path[0] == '/' || path[len(path)-1] == '/' {
					i.addValidationError(E100, "invalid path \"%s\" in fixity", path)
				}
				if path == "" {
					i.addValidationError(E099, "empty path in fixity")
				}
				path2 := path
				if path[0] == '/' {
					path2 = path[1:]
				}
				elements := strings.Split(path2, "/")
				for _, element := range elements {
					if slices.Contains([]string{"", ".", ".."}, element) {
						i.addValidationError(E099, "invalid path \"%s\" in fixity", path)
					}
				}
			}
		}
	}
	return nil
}

func (i *InventoryBase) checkVersions() error {
	var paddingLength int = -1
	var versions = []int{}
	if len(i.Versions.Versions) == 0 {
		i.addValidationError(E008, "length of ver is 0")
	}
	for ver, version := range i.Versions.Versions {
		vInt, ok := i.versionValue[ver]
		if !ok {
			//			i.addValidationError(E104, "invalid ver format %s", ver)
			continue
		}
		versions = append(versions, int(vInt))
		if versionZeroRegexp.MatchString(ver) {
			if paddingLength == -1 {
				paddingLength = len(ver) - 2
			} else {
				if paddingLength != len(ver)-2 {
					//i.addValidationError(E011, "invalid ver padding %s", ver)
					i.addValidationError(E012, "invalid ver padding %s", ver)
					i.addValidationError(E013, "invalid ver padding %s", ver)
				}
			}
		} else {
			if versionNoZeroRegexp.MatchString(ver) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						i.addValidationError(E011, "invalid ver padding %s", ver)
						i.addValidationError(E012, "invalid ver padding %s", ver)
						i.addValidationError(E013, "invalid ver padding %s", ver)
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				i.addValidationError(E104, "invalid ver format %s", ver)
			}
		}
		if version.Created.err != nil {
			i.addValidationError(E049, "invalid created format in version %s: %v", ver, version.Created.err.Error())
		}
		if version.User.err != nil {
			i.addValidationError(E054, "invalid user in version %s: %v", ver, version.Created.err.Error())
		}
		if version.User.Name.err != nil {
			i.addValidationError(E054, "invalid user.name in version %s: %v", ver, version.Created.err.Error())
		}
		if version.User.Address.err != nil {
			i.addValidationError(E054, "invalid user.address in version %s: %v", ver, version.Created.err.Error())
		}
		if version.State.err != nil {
			i.addValidationError(E050, "invalid state format in version %s: %v", ver, version.State.err.Error())
		}
		for digest, paths := range version.State.State {
			digestLowerUpper := []string{strings.ToLower(digest), strings.ToUpper(digest)}
			found := false
			for mDigest, _ := range i.Manifest {
				if mDigest == digest {
					found = true
				} else {
					if slices.Contains(digestLowerUpper, mDigest) {
						i.addValidationError(E050, "wrong digest case in version %s - '%s' != '%s'", ver, digest, mDigest)
						found = true
						break
					}
				}
			}
			if !found {
				i.addValidationError(E050, "digest not in manifest of versions %s - '%s '", ver, digest)
			}
			for _, path := range paths {
				if path[0] == '/' || path[len(path)-1] == '/' {
					i.addValidationError(E053, "invalid path \"%s\" in state for version %s", path, ver)
				}
				if path == "" {
					i.addValidationError(E051, "empty path in state for version %s", ver)
				}
				path2 := path
				if path[0] == '/' {
					path2 = path[1:]
				}
				elements := strings.Split(path2, "/")
				for _, element := range elements {
					if slices.Contains([]string{"", ".", ".."}, element) {
						i.addValidationError(E052, "invalid path \"%s\" in state for version %s", path, ver)
					}
				}
			}
		}
	}
	slices.Sort(versions)
	for key, val := range versions {
		if key != val-1 {
			i.addValidationError(E010, "invalid ver sequence %v", versions)
			break
		}
	}
	i.paddingLength = paddingLength

	// check head is recent ver
	var recentVersion string
	for _, ver := range i.GetVersionStrings() {
		if recentVersion == "" {
			recentVersion = ver
		} else {
			if !i.VersionLessOrEqual(ver, recentVersion) {
				recentVersion = ver
			}
		}
	}
	if i.GetHead() != recentVersion {
		i.addValidationError(E040, "manifest head %s is not recent ver %s", i.GetHead(), recentVersion)
	}

	// check that head exists in versions
	if !slices.Contains(i.GetVersionStrings(), i.Head) {
		i.addValidationError(E040, "manifest head %s does not exists in versions %v", i.Head, i.GetVersionStrings())
	}

	// check logical paths
	logPaths := []string{}
	for _, version := range i.Versions.Versions {
		for _, paths := range version.State.State {
			logPaths = append(logPaths, paths...)
		}
	}
	slices.Sort(logPaths)
	for j := 0; j < len(logPaths)-1; j++ {
		if strings.HasPrefix(logPaths[j+1], logPaths[j]) {
			i.addValidationError(E095, "logical path '%s' is prefix of '%s'", logPaths[j], logPaths[j+1])
		}
	}

	return nil
}

func (i *InventoryBase) GetFiles() map[string][]string {
	var result = map[string][]string{}
	versions := []string{}
	for _, files := range i.Manifest {
		for _, filename := range files {
			parts := strings.Split(filename, "/")
			if len(parts) < 3 {
				i.addValidationError(E000, "invalid filepath in manifest \"%s\"", filename)
			}
			version := parts[0]
			//fn := parts[2]
			if parts[1] != i.GetContentDir() {
				i.addValidationError(E019, "invalid content directory \"%s\" in \"%s\"", parts[1], filename)
			}
			if _, ok := result[version]; !ok {
				versions = append(versions, version)
				result[version] = []string{}
			}
			result[version] = append(result[version], filename)
		}
	}
	iVersions := i.GetVersionStrings()
	if !sliceContains(iVersions, versions) {
		i.addValidationError(E023, "versions %v do not contains versions from manifest %v", iVersions, versions)
	}
	return result
}

func (i *InventoryBase) GetManifest() map[string][]string {
	return i.Manifest
}

func (i *InventoryBase) GetFixity() map[checksum.DigestAlgorithm]map[string][]string {
	if i.Fixity == nil {
		return map[checksum.DigestAlgorithm]map[string][]string{}
	}
	return i.Fixity
}
func (i *InventoryBase) GetFilesFlat() []string {
	filesV := i.GetFiles()
	result := []string{}
	for _, files := range filesV {
		for _, file := range files {
			result = append(result, file)
		}
	}
	return result
}

func (i *InventoryBase) BuildRealname(virtualFilename string) string {
	//	return fmt.Sprintf("%s/%s/%s", i.GetHead(), i.GetContentDirectory(), FixFilename(filepath.ToSlash(virtualFilename)))
	return fmt.Sprintf("%s/%s/%s", i.GetHead(), i.GetContentDirectory(), filepath.ToSlash(virtualFilename))
}

func (i *InventoryBase) NewVersion(msg, UserName, UserAddress string) error {
	if i.IsWriteable() {
		return errors.New(fmt.Sprintf("version %s already writeable", i.GetHead()))
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
	i.Versions.Versions[i.Head] = &Version{
		Created: OCFLTime{time.Now(), nil},
		Message: OCFLString{msg, nil},
		State:   OCFLState{},
		User: OCFLUser{
			User: User{
				Name:    OCFLString{string: UserName},
				Address: OCFLString{string: UserAddress},
			},
		},
	}
	// copy last state...
	if lastHead != "" {
		copyMapStringSlice(i.Versions.Versions[i.Head].State.State, i.Versions.Versions[lastHead].State.State)
	}
	i.writeable = true
	return nil
}

var vRegexp *regexp.Regexp = regexp.MustCompile("^v(\\d+)$")

func (i *InventoryBase) getLastVersion() (string, error) {
	versions := []int{}
	for ver, _ := range i.Versions.Versions {
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
	for ver, version := range i.Versions.Versions {
		for checksum, filenames := range version.State.State {
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
	for ver, version := range i.Versions.Versions {
		for checksum, filenames := range version.State.State {
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
	for key, vals := range i.Versions.Versions[i.GetHead()].State.State {
		newState[key] = []string{}
		for _, val := range vals {
			if val == virtualFilename {
				found = true
			} else {
				newState[key] = append(newState[key], val)
			}
		}
	}
	i.Versions.Versions[i.GetHead()].State.State = newState
	i.modified = found
	return nil
}
func (i *InventoryBase) Rename(oldVirtualFilename, newVirtualFilename string) error {
	var newState = map[string][]string{}
	var found = false
	for key, vals := range i.Versions.Versions[i.GetHead()].State.State {
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
	i.Versions.Versions[i.GetHead()].State.State = newState
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

	if _, ok := i.Versions.Versions[i.Head].State.State[checksum]; !ok {
		i.Versions.Versions[i.Head].State.State[checksum] = []string{}
	}

	upd, err := i.IsUpdate(virtualFilename, checksum)
	if err != nil {
		return errors.Wrapf(err, "cannot check for update of %s [%s]", virtualFilename, checksum)
	}
	if upd {
		i.logger.Debugf("%s is an update - removing old version", virtualFilename)
		i.DeleteFile(virtualFilename)
	}

	i.Versions.Versions[i.Head].State.State[checksum] = append(i.Versions.Versions[i.Head].State.State[checksum], virtualFilename)

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
	if i.GetHead() == "v1" {
		return nil
	}
	i.logger.Debugf("deleting %v", i.GetHead())
	delete(i.Versions.Versions, i.GetHead())
	lastVersion, err := i.getLastVersion()
	if err != nil {
		return errors.Wrap(err, "cannot get last version")
	}
	i.Head = lastVersion
	return nil
}
