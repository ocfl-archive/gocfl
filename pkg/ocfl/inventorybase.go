package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/uri"
	"github.com/op/go-logging"
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
	ctx                    context.Context
	folder                 string
	object                 Object
	modified               bool
	writeable              bool
	paddingLength          int
	versionValue           map[string]uint
	fixityDigestAlgorithms []checksum.DigestAlgorithm
	Id                     string                                           `json:"id"`
	Type                   InventorySpec                                    `json:"type"`
	DigestAlgorithm        checksum.DigestAlgorithm                         `json:"digestAlgorithm"`
	Head                   *OCFLString                                      `json:"head"`
	ContentDirectory       string                                           `json:"contentDirectory,omitempty"`
	Manifest               *OCFLManifest                                    `json:"manifest,omitempty"`
	Versions               *OCFLVersions                                    `json:"versions"`
	Fixity                 map[checksum.DigestAlgorithm]map[string][]string `json:"fixity,omitempty"`
	logger                 *logging.Logger
}

func newInventoryBase(ctx context.Context, object Object, folder string, objectType *url.URL, contentDir string, logger *logging.Logger) (*InventoryBase, error) {
	i := &InventoryBase{
		ctx:                    ctx,
		object:                 object,
		folder:                 folder,
		paddingLength:          0,
		fixityDigestAlgorithms: []checksum.DigestAlgorithm{},
		Type:                   InventorySpec(objectType.String()),
		Head:                   NewOCFLString(""),
		ContentDirectory:       contentDir,
		Manifest:               nil,
		Versions:               &OCFLVersions{Versions: map[string]*Version{}},
		Fixity:                 nil,
		logger:                 logger,
	}
	return i, nil
}

func (i *InventoryBase) isEqual(i2 *InventoryBase) bool {

	if !sliceContains(i.fixityDigestAlgorithms, i2.fixityDigestAlgorithms) || len(i.fixityDigestAlgorithms) != len(i2.fixityDigestAlgorithms) {
		return false
	}
	if i.Type != i2.Type {
		return false
	}
	if i.Head.string != i2.Head.string {
		return false
	}
	if i.ContentDirectory != i2.ContentDirectory {
		return false
	}
	if (i.Manifest == nil && i2.Manifest != nil) || (i.Manifest != nil && i2.Manifest == nil) {
		return false
	}
	if i.Manifest != nil {
		if len(i.Manifest.Manifest) != len(i.Manifest.Manifest) {
			return false
		}
		for key, vals := range i.Manifest.Manifest {
			vals2, ok := i2.Manifest.Manifest[key]
			if !ok {
				return false
			}
			if !sliceContains(vals, vals2) || len(vals) != len(vals2) {
				return false
			}
		}
	}
	if (i.Versions == nil && i2.Versions != nil) || (i.Versions != nil && i2.Versions == nil) {
		return false
	}
	if i.Versions != nil {
		if len(i.Versions.Versions) != len(i2.Versions.Versions) {
			return false
		}
		for key, version := range i.Versions.Versions {
			version2, ok := i2.Versions.Versions[key]
			if !ok {
				return false
			}
			if !version.EqualMeta(version2) {
				return false
			}
			if !version.EqualState(version2) {
				return false
			}
		}
	}
	return true
}

func (i *InventoryBase) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (err error) {
	i.Id = id
	i.DigestAlgorithm = digest
	i.fixityDigestAlgorithms = fixity
	return nil
}
func (i *InventoryBase) Finalize(inCreation bool) (err error) {
	if i.Manifest == nil {
		if !inCreation {
			i.addValidationError(E041, "no manifest in inventory")
		}
		i.Manifest = &OCFLManifest{Manifest: map[string][]string{}}
	}

	if i.Versions == nil {
		if !inCreation {
			i.addValidationError(E041, "no versions in inventory")
		}
		i.Versions = &OCFLVersions{Versions: map[string]*Version{}}
	}

	i.versionValue = map[string]uint{}
	for ver, version := range i.Versions.Versions {
		vInt, err := strconv.Atoi(strings.TrimLeft(ver, "v0"))
		if err != nil {
			i.addValidationError(E104, "invalid version format '%s'", ver)
			continue
		}
		i.versionValue[ver] = uint(vInt)
		if version.User == nil {
			i.addValidationWarning(W007, "no user key in version '%s'", ver)
			version.User = NewOCFLUser("", "")
		}
		version.User.Finalize()
		if version.Message == nil {
			i.addValidationWarning(W007, "no message key in version '%s'", ver)
			version.Message = NewOCFLString("")
		}
		if version.State == nil {
			version.State = &OCFLState{
				State: map[string][]string{},
				err:   nil,
			}
		}
	}
	for alg := range i.Fixity {
		if !slices.Contains(i.fixityDigestAlgorithms, alg) {
			i.fixityDigestAlgorithms = append(i.fixityDigestAlgorithms, alg)
		}
	}

	if !inCreation {
		if err := i.check(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (i *InventoryBase) addValidationError(errno ValidationErrorCode, format string, a ...any) {
	_ = addValidationErrors(i.ctx, GetValidationError(i.object.GetVersion(), errno).AppendDescription(format, a...).AppendDescription("(%s/inventory.json)", i.folder).AppendContext("object '%s' - '%s'", i.object.GetFS(), i.GetID()))
}
func (i *InventoryBase) addValidationWarning(errno ValidationErrorCode, format string, a ...any) {
	_ = addValidationWarnings(i.ctx, GetValidationError(i.object.GetVersion(), errno).AppendDescription(format, a...).AppendDescription("(%s/inventory.json)", i.folder).AppendContext("object '%s' - '%s'", i.object.GetFS(), i.GetID()))
}
func (i *InventoryBase) GetID() string          { return i.Id }
func (i *InventoryBase) GetHead() string        { return i.Head.string }
func (i *InventoryBase) GetSpec() InventorySpec { return i.Type }

func (i *InventoryBase) GetContentDir() string {
	if i.ContentDirectory == "" {
		return "content"
	}
	return i.ContentDirectory
}

func (i *InventoryBase) GetRealContentDir() string {
	return i.ContentDirectory
}

func (i *InventoryBase) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }
func (i *InventoryBase) GetFixityDigestAlgorithm() []checksum.DigestAlgorithm {
	return i.fixityDigestAlgorithms
}
func (i *InventoryBase) IsWriteable() bool { return i.writeable }
func (i *InventoryBase) IsModified() bool  { return i.modified }

func (i *InventoryBase) GetVersionStrings() []string {
	if len(i.Versions.Versions) == 0 {
		return []string{}
	}

	versionsInt := []int{}
	versionString := map[int]string{}
	for ver := range i.Versions.Versions {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return []string{}
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return []string{}
		}
		versionsInt = append(versionsInt, versionInt)
		versionString[versionInt] = ver
	}

	// sort versions ascending
	sort.Ints(versionsInt)
	var versions = []string{}
	for _, versionInt := range versionsInt {
		versions = append(versions, versionString[versionInt])
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

func (i *InventoryBase) GetStateFiles(version string, cs string) ([]string, error) {
	if version == "latest" || version == "" {
		version = i.GetHead()
	}
	ver, err := i.Versions.GetVersion(version)
	if err != nil {
		return nil, errors.Errorf("invalid version '%s'", version)
	}
	files, ok := ver.State.State[cs]
	if !ok {
		return nil, errors.Errorf("no state for in version %s [%s]", version, cs)
	}
	return files, nil
}

func (i *InventoryBase) IterateStateFiles(version string, fn StateFileCallback) error {
	if version == "latest" || version == "" {
		version = i.GetHead()
	}
	ver, err := i.Versions.GetVersion(version)
	if err != nil {
		return errors.Errorf("invalid version '%s'", version)
	}
	for digest, externalNames := range ver.State.State {
		internalNames, ok := i.Manifest.Manifest[digest]
		if !ok {
			return errors.Errorf("no manifest for [%s]%v", digest, externalNames)
		}
		if len(internalNames) == 0 {
			return errors.Errorf("invalid manifest for digest [%s]", digest)
		}
		if err := errors.WithStack(fn(internalNames, externalNames, digest)); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
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
	if i.Id != "" {
		if _, err := uri.Parse(i.Id); err != nil {
			i.addValidationWarning(W005, "cannot parse uri id '%s': %v", i.Id, err)
		} /* else {
			if u.Scheme == "" {
				i.addValidationWarning(W005, "id '%s' is not an uri", i.Id)
			}
		}
		*/
	}
	if i.Head.err != nil {
		i.addValidationError(E040, "invalid field \"head\" for object: %v", i.Head.err)
	} else {
		if i.Head.string == "" {
			i.addValidationError(E036, "invalid field \"head\" for object")
		}
	}
	if i.Type == "" {
		i.addValidationError(E036, "invalid field \"type\" for object")
	}
	if i.DigestAlgorithm == "" {
		i.addValidationError(E036, "invalid field \"digestAlgorithm\" for object")
	}

	if !slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA512, checksum.DigestSHA256}, i.DigestAlgorithm) {
		i.addValidationError(E025, "invalid digest algorithm '%s'", i.DigestAlgorithm)
	} else {
		if slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA256}, i.DigestAlgorithm) {
			i.addValidationError(W004, "digest algorithm '%s' not suggested", i.DigestAlgorithm)
		}
	}

	if i.ContentDirectory != "" {
		if slices.Contains([]string{"", ".", ".."}, i.ContentDirectory) || strings.Contains(i.ContentDirectory, "/") {
			i.addValidationError(E017, "invalid content directory '%s'", i.ContentDirectory)
		}
	}

	return nil
}
func (i *InventoryBase) checkManifest() error {
	i.logger.Debugf("[%s] checkManifest", i.GetID())
	defer i.logger.Debugf("[%s] checkManifest done", i.GetID())
	versionDigests := []string{}
	for _, version := range i.Versions.Versions {
		for digest := range version.State.State {
			versionDigests = append(versionDigests, digest)
		}
	}
	slices.Sort(versionDigests)

	digests := []string{}
	allPaths := []string{}
	for digest, paths := range i.Manifest.Manifest {
		//		digest = strings.ToLower(digest)
		if slices.Contains(digests, digest) {
			i.addValidationError(E096, "manifest digest '%s' is duplicate", digest)
		} else {
			digests = sliceInsertSorted(digests, digest)
			//digests = append(digests, digest)
			if !slices.Contains(versionDigests, digest) {
				i.addValidationError(E107, "digest '%s' does not appear in any version", digest)
			}
		}
		for _, path := range paths {
			//allPaths = sliceInsertSorted(allPaths, path)
			allPaths = append(allPaths, path)
			if path[0] == '/' || path[len(path)-1] == '/' {
				i.addValidationError(E100, "invalid path '%s' in manifest", path)
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
					i.addValidationError(E099, "invalid path '%s' in manifest", path)
				}
			}

		}

	}
	i.logger.Debugf("[%s] checkManifest prefix", i.GetID())
	slices.Sort(allPaths)
	for j := 0; j < len(allPaths)-1; j++ {
		prefix := strings.TrimRight(allPaths[j+1], "/") + "/"
		if strings.HasPrefix(allPaths[j], prefix) {
			i.addValidationError(E101, "content path '%s' is prefix or equal to '%s' in manifest", allPaths[j], prefix)
		}
	}
	return nil
}

func (i *InventoryBase) checkFixity() error {
	i.logger.Debugf("[%s] checkFixity", i.GetID())
	defer i.logger.Debugf("[%s] checkFixity done", i.GetID())
	for digestAlg, digestMap := range i.Fixity {
		digests := []string{}
		for digest, paths := range digestMap {
			lowerDigest := strings.ToLower(digest)
			if _, found := slices.BinarySearch(digests, lowerDigest); found {
				i.addValidationError(E097, "fixity '%s' digest '%s' is duplicate", digestAlg, digest)
			} else {
				digests = sliceInsertSorted(digests, lowerDigest)
				//digests = append(digests, lowerDigest)
			}
			// check content paths
			for _, path := range paths {
				if path[0] == '/' || path[len(path)-1] == '/' {
					i.addValidationError(E100, "invalid path '%s' in fixity", path)
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
						i.addValidationError(E099, "invalid path '%s' in fixity", path)
					}
				}
			}
		}
	}
	return nil
}

func (i *InventoryBase) checkVersions() error {
	i.logger.Debugf("[%s] checkVersions", i.GetID())
	defer i.logger.Debugf("[%s] checkVersions done", i.GetID())
	var paddingLength int = -1
	var versions = []int{}
	if len(i.Versions.Versions) == 0 {
		i.addValidationError(E008, "length of ver is 0")
	}
	manifestDigests := []string{}
	manifestDigestsLower := []string{}
	for mDigest := range i.Manifest.Manifest {
		manifestDigests = append(manifestDigests, mDigest)
		manifestDigestsLower = append(manifestDigestsLower, strings.ToLower(mDigest))
	}
	slices.Sort(manifestDigests)
	slices.Sort(manifestDigestsLower)

	for ver, version := range i.Versions.Versions {
		i.logger.Debugf("[%s] checkVersions '%s'", i.GetID(), ver)
		vInt, ok := i.versionValue[ver]
		if !ok {
			//			i.addValidationError(E104, "invalid ver format '%s'", ver)
			continue
		}
		versions = append(versions, int(vInt))
		if versionZeroRegexp.MatchString(ver) {
			if paddingLength == -1 {
				paddingLength = len(ver) - 2
			} else {
				if paddingLength != len(ver)-2 {
					//i.addValidationError(E011, "invalid ver padding '%s'", ver)
					i.addValidationError(E012, "invalid ver padding '%s'", ver)
					i.addValidationError(E013, "invalid ver padding '%s'", ver)
				}
			}
		} else {
			if versionNoZeroRegexp.MatchString(ver) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						i.addValidationError(E011, "invalid ver padding '%s'", ver)
						i.addValidationError(E012, "invalid ver padding '%s'", ver)
						i.addValidationError(E013, "invalid ver padding '%s'", ver)
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				i.addValidationError(E104, "invalid version format '%s'", ver)
			}
		}
		if version.Created.err != nil {
			i.addValidationError(E049, "invalid created format in version '%s': %v", ver, version.Created.err.Error())
		}
		if version.User.err != nil {
			i.addValidationError(E054, "invalid user in version '%s': %v", ver, version.User.err.Error())
		}
		if version.User.Name.err != nil {
			i.addValidationError(E054, "invalid user name in version '%s': %v", ver, version.User.Name.err.Error())
		}
		if version.User.Address.err != nil {
			i.addValidationError(E054, "invalid user address in version '%s': %v", ver, version.User.Address.err.Error())
		}
		if version.User.Address.string == "" {
			i.addValidationWarning(W008, "no user address in version '%s'", ver)
		} else {
			mailtoUriRegexp := regexp.MustCompile(`mailto:[^@]+@[^@]+`)
			if !mailtoUriRegexp.MatchString(version.User.Address.string) {
				u, err := url.Parse(version.User.Address.string)
				if err != nil {
					i.addValidationWarning(W009, "cannot parse user address '%s' in version '%s': %v", version.User.Address.string, ver, err)
				} else {
					if u.Scheme == "" {
						i.addValidationWarning(W009, "cannot parse user address '%s' in version '%s'", version.User.Address.string, ver)
					}
				}
			}
		}
		if version.Message.err != nil {
			i.addValidationError(E094, "invalid format for message in version '%s': %v", ver, version.Message.err)
		}

		if version.State.err != nil {
			i.addValidationError(E050, "invalid state format in version '%s': %v", ver, version.State.err.Error())
		}
		i.logger.Debugf("[%s] checkVersions %s state", i.GetID(), ver)
		for digest, paths := range version.State.State {
			// massive performance boost by using sorted manifest
			if _, found := slices.BinarySearch(manifestDigests, digest); !found {
				if _, found := slices.BinarySearch(manifestDigestsLower, strings.ToLower(digest)); found {
					i.addValidationError(E096, "wrong digest case in version '%s' - '%s'", ver, digest)
				} else {
					i.addValidationError(E050, "digest not in manifest of versions '%s' - '%s'", ver, digest)
				}
			}
			for _, path := range paths {
				if path[0] == '/' || path[len(path)-1] == '/' {
					i.addValidationError(E053, "invalid path '%s' in state for version '%s'", path, ver)
				}
				if path == "" {
					i.addValidationError(E051, "empty path in state for version '%s'", ver)
				}
				path2 := path
				if path[0] == '/' {
					path2 = path[1:]
				}
				elements := strings.Split(path2, "/")
				for _, element := range elements {
					if slices.Contains([]string{"", ".", ".."}, element) {
						i.addValidationError(E052, "invalid path '%s' in state for version '%s'", path, ver)
					}
				}
			}
		}
		i.logger.Debugf("[%s] checkVersions '%s' state done", i.GetID(), ver)
		i.logger.Debugf("[%s] checkVersions '%s' done", i.GetID(), ver)
	}
	slices.Sort(versions)
	for key, val := range versions {
		if key != val-1 {
			i.addValidationError(E010, "invalid ver sequence %v", versions)
			break
		}
	}
	i.paddingLength = paddingLength
	if paddingLength > 0 {
		i.addValidationWarning(W001, "padding length is %v", i.paddingLength)
	}

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
	if i.GetHead() != recentVersion && i.GetHead() != "" {
		i.addValidationError(E040, "manifest head '%s' is not recent ver '%s'", i.GetHead(), recentVersion)
	}

	// check that head exists in versions
	if i.Head.string != "" && !slices.Contains(i.GetVersionStrings(), i.Head.string) {
		i.addValidationError(E040, "manifest head '%s' does not exists in versions %v", i.Head.string, i.GetVersionStrings())
	}

	// check logical paths
	for _, version := range i.Versions.Versions {
		logPaths := []string{}
		for _, paths := range version.State.State {
			logPaths = append(logPaths, paths...)
		}
		slices.Sort(logPaths)
		for j := 0; j < len(logPaths)-1; j++ {
			prefix := strings.TrimSuffix(logPaths[j], "/") + "/"
			if strings.HasPrefix(logPaths[j+1], prefix) {
				i.addValidationError(E095, "logical path '%s' is prefix of '%s'", logPaths[j], logPaths[j+1])
			}
		}
	}

	return nil
}

func (i *InventoryBase) CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error {
	i.logger.Debugf("[%s] checkFiles", i.GetID())
	defer i.logger.Debugf("[%s] checkFiles done", i.GetID())
	csFiles, ok := fileManifest[i.GetDigestAlgorithm()]
	if !ok {
		if len(fileManifest) == 0 {
			return nil
		}
		return errors.Errorf("checksum for '%s' not created", i.GetDigestAlgorithm())
	}
	for digest, files := range i.GetManifest() {
		csFilenames, ok := csFiles[strings.ToLower(digest)]
		if !ok {
			i.addValidationError(E092, "digest '%s' for file(s) %v not found in content", digest, files)
			continue
		}
		for _, file := range files {
			if !slices.Contains(csFilenames, file) {
				i.addValidationError(E092, "invalid digest for file '%s'", file)
			}
		}
	}
	//check fixity
	for digestAlg, fixity := range i.GetFixity() {
		csFiles, ok = fileManifest[digestAlg]
		if !ok {
			return errors.Errorf("checksum for '%s' not created", digestAlg)
		}
		for digest, files := range fixity {
			csFilenames, ok := csFiles[digest]
			if !ok {
				csFilenames, ok = csFiles[strings.ToLower(digest)]
				if !ok {
					i.addValidationError(E093, "fixity digest '%s' for file(s) %v not found in content", digest, files)
					continue
				}
			}
			for _, file := range files {
				if !slices.Contains(csFilenames, file) {
					i.addValidationError(E093, "invalid fixity digest for file '%s'", file)
				}
			}
		}

	}
	return nil
}

func (i *InventoryBase) GetFiles() map[string][]string {
	var result = map[string][]string{}
	versions := []string{}
	for _, files := range i.Manifest.Manifest {
		for _, filename := range files {
			parts := strings.Split(filename, "/")
			if len(parts) < 3 {
				i.addValidationError(E000, "invalid filepath in manifest '%s'", filename)
			}
			version := parts[0]
			//fn := parts[2]
			if parts[1] != i.GetContentDir() {
				//i.addValidationError(E015, "extra file/directory '%s' in manifest", parts[1])
				//i.addValidationError(E019, "invalid content directory '%s' in '%s'", parts[1], filename)
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
		slices.Sort(iVersions)
		i.addValidationError(E023, "versions %v do not contains versions from manifest %v", iVersions, versions)
	}
	return result
}

func (i *InventoryBase) GetManifest() map[string][]string {
	return i.Manifest.Manifest
}

func (i *InventoryBase) GetFixity() Fixity {
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

func (i *InventoryBase) BuildManifestName(stateFilename string) string {
	return i.BuildManifestNameVersion(stateFilename, i.GetHead())
}

func (i *InventoryBase) BuildManifestNameVersion(stateFilename string, version string) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(version, i.GetContentDir(), stateFilename)))
}

func (i *InventoryBase) NewVersion(msg, UserName, UserAddress string) error {
	/*
		if i.IsWriteable() {
			return errors.New(fmt.Sprintf("version '%s' already writeable", i.GetHead()))
		}
	*/
	lastHead := i.Head.string
	if lastHead == "" {
		if i.paddingLength <= 0 {
			i.Head.string = "v1"
			i.Head.err = nil
		} else {
			i.Head.string = fmt.Sprintf(fmt.Sprintf("v0%%0%dd", i.paddingLength), 1)
		}
	} else {
		vStr := strings.TrimLeft(strings.ToLower(i.Head.string), "v0")
		v, err := strconv.Atoi(vStr)
		if err != nil {
			return errors.Wrapf(err, "cannot determine head of ObjectBase - '%s'", vStr)
		}

		if i.paddingLength <= 0 {
			i.Head.string = fmt.Sprintf("v%d", v+1)
		} else {
			i.Head.string = fmt.Sprintf(fmt.Sprintf("v0%%0%dd", i.paddingLength), v+1)
		}
	}
	i.Versions.Versions[i.Head.string] = &Version{
		Created: &OCFLTime{time.Now(), nil},
		Message: NewOCFLString(msg),
		State:   &OCFLState{State: map[string][]string{}},
		User:    NewOCFLUser(UserName, UserAddress),
	}
	// copy last state...
	if lastHead != "" {
		copyMapStringSlice(i.Versions.Versions[i.Head.string].State.State, i.Versions.Versions[lastHead].State.State)
	}
	i.writeable = true
	return nil
}

var vRegexp *regexp.Regexp = regexp.MustCompile("^v(\\d+)$")

func (i *InventoryBase) getLastVersion() string {
	if len(i.Versions.Versions) == 0 {
		return ""
	}
	versions := []int{}
	versionString := map[int]string{}
	for ver := range i.Versions.Versions {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return ""
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return ""
		}
		versions = append(versions, versionInt)
		versionString[versionInt] = ver
	}

	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	return versionString[lastVersion]
}
func (i *InventoryBase) GetDuplicates(checksum string) []string {
	// not necessary but fast...
	if checksum == "" {
		return nil
	}
	for cs, files := range i.Manifest.Manifest {
		if cs == checksum {
			return files
		}
	}
	return nil
}
func (i *InventoryBase) AlreadyExists(stateFilename, checksum string) (bool, error) {
	i.logger.Debugf("'%s' [%s]", stateFilename, checksum)
	if checksum == "" {
		i.logger.Debugf("'%s' - duplicate %v", stateFilename, false)
		return false, nil
	}

	// first get checksum of last version of a file
	css := map[string]string{}
	for ver, version := range i.Versions.Versions {
		for cs, filenames := range version.State.State {
			found := false
			for _, filename := range filenames {
				if filename == stateFilename {
					css[ver] = cs
					found = true
				}
			}
			if found {
				break
			}
		}
	}
	if len(css) == 0 {
		i.logger.Debugf("'%s' - duplicate %v", stateFilename, false)
		return false, nil
	}
	versions := []int{}

	for ver := range css {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return false, errors.New(fmt.Sprintf("invalid version in inventory - '%s'", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return false, errors.Wrapf(err, "cannot convert version number to int - '%s'", matches[1])
		}
		versions = append(versions, versionInt)
	}
	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	lastChecksum, ok := css[fmt.Sprintf("v%d", lastVersion)]
	if !ok {
		return false, errors.New(fmt.Sprintf("could not get checksum for v%d", lastVersion))
	}
	i.logger.Debugf("'%s' - duplicate %v", stateFilename, lastChecksum == checksum)
	return lastChecksum == checksum, nil
}

func (i *InventoryBase) IsUpdate(virtualFilename, checksum string) (bool, error) {
	i.logger.Debugf("'%s' [%s]", virtualFilename, checksum)
	if checksum == "" {
		i.logger.Debugf("'%s' - update %v", virtualFilename, false)
		return false, nil
	}

	// first get checksum of last version of a file
	css := map[string]string{}
	for ver, version := range i.Versions.Versions {
		for cs, filenames := range version.State.State {
			found := false
			for _, filename := range filenames {
				if filename == virtualFilename {
					css[ver] = cs
					found = true
				}
			}
			if found {
				break
			}
		}
	}
	if len(css) == 0 {
		i.logger.Debugf("'%s' - update %v", virtualFilename, false)
		return false, nil
	}
	versions := []int{}

	for ver := range css {
		matches := vRegexp.FindStringSubmatch(ver)
		if matches == nil {
			return false, errors.New(fmt.Sprintf("invalid version in inventory - '%s'", ver))
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return false, errors.Wrapf(err, "cannot convert version number to int - '%s'", matches[1])
		}
		versions = append(versions, versionInt)
	}
	// sort versions ascending
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]
	lastChecksum, ok := css[fmt.Sprintf("v%d", lastVersion)]
	if !ok {
		return false, errors.New(fmt.Sprintf("could not get checksum for v%d", lastVersion))
	}
	i.logger.Debugf("'%s' - update %v", virtualFilename, lastChecksum != checksum)
	return lastChecksum != checksum, nil
}

func (i *InventoryBase) echoDelete(existing []string, pathPrefix string) error {
	var deleteFiles = []string{}
	version, ok := i.Versions.Versions[i.GetHead()]
	if !ok {
		return errors.Errorf("cannot get version '%s'", i.Head.string)
	}
	statefiles := []string{}
	for _, state := range version.State.State {
		for _, filename := range state {
			statefiles = append(statefiles, filename)
		}
	}
	for _, filename := range statefiles {
		if !strings.HasPrefix(filename, pathPrefix) {
			continue
		}
		if _, found := slices.BinarySearch(existing, filename); !found {
			deleteFiles = append(deleteFiles, filename)
		}

	}
	for _, filename := range deleteFiles {
		//		i.logger.Infof("removing '%s' from state", filename)
		if err := i.DeleteFile(filename); err != nil {
			return errors.Wrapf(err, "cannot delete '%s'", filename)
		}
	}
	return nil
}

func (i *InventoryBase) DeleteFile(stateFilename string) error {
	var newState = map[string][]string{}
	var found = false
	for key, state := range i.Versions.Versions[i.GetHead()].State.State {
		newState[key] = []string{}
		for _, val := range state {
			if val == stateFilename {
				found = true
			} else {
				newState[key] = append(newState[key], val)
			}
		}
	}
	i.Versions.Versions[i.GetHead()].State.State = newState
	if found {
		i.modified = found
		i.logger.Infof("[%s] removing '%s' from state", i.GetID(), stateFilename)
	}
	return nil
}

/*
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
*/

func (i *InventoryBase) CopyFile(dest string, digest string) error {
	i.logger.Infof("[%s] copying '%s' -> '%s'", i.GetID(), digest, dest)

	if _, ok := i.Manifest.Manifest[digest]; !ok {
		return errors.Errorf("cannot find file with digest '%s'", digest)
	}
	// nothing to do if already there...
	if slices.Contains(i.Versions.Versions[i.Head.string].State.State[digest], dest) {
		return nil
	}
	i.Versions.Versions[i.Head.string].State.State[digest] = append(i.Versions.Versions[i.Head.string].State.State[digest], dest)
	i.modified = true
	return nil
}

func (i *InventoryBase) AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error {
	i.logger.Debugf("[%s] adding '%s' -> '%s'", i.GetID(), stateFilenames, manifestFilename)
	digest, ok := checksums[i.GetDigestAlgorithm()]
	if !ok {
		return errors.Errorf("no digest for '%s' in checksums", i.GetDigestAlgorithm())
	}
	digest = strings.ToLower(digest) // paranoia

	for alg, fixityDigest := range checksums {
		if alg == i.GetDigestAlgorithm() {
			continue
		}
		if i.Fixity == nil {
			i.Fixity = map[checksum.DigestAlgorithm]map[string][]string{}
		}
		if _, ok := i.Fixity[alg]; !ok {
			i.Fixity[alg] = map[string][]string{}
		}
		if _, ok := i.Fixity[alg][fixityDigest]; !ok {
			i.Fixity[alg][fixityDigest] = []string{}
		}
		if !slices.Contains(i.Fixity[alg][fixityDigest], manifestFilename) {
			i.Fixity[alg][fixityDigest] = append(i.Fixity[alg][fixityDigest], manifestFilename)
			i.modified = true
		}
	}

	for _, virtualFilename := range stateFilenames {
		dup, err := i.AlreadyExists(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add for duplicate of '%s' [%s]", stateFilenames, digest)
		}
		if dup {
			i.logger.Debugf("'%s' is a duplicate", stateFilenames)
			// return nil
		}

		if manifestFilename != "" {
			if _, ok := i.Manifest.Manifest[digest]; !ok {
				i.Manifest.Manifest[digest] = []string{}
			}
			i.Manifest.Manifest[digest] = append(i.Manifest.Manifest[digest], manifestFilename)
		}

		if _, ok := i.Versions.Versions[i.Head.string].State.State[digest]; !ok {
			i.Versions.Versions[i.Head.string].State.State[digest] = []string{}
		}

		upd, err := i.IsUpdate(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check for update of '%s' [%s]", stateFilenames, digest)
		}
		if upd {
			i.logger.Debugf("'%s' is an update - removing old version", stateFilenames)
			if err := i.DeleteFile(virtualFilename); err != nil {
				return errors.Wrapf(err, "cannot delete old version of '%s' [%s]", stateFilenames, digest)
			}
			i.modified = true
		}

		if !dup {
			i.Versions.Versions[i.Head.string].State.State[digest] = append(i.Versions.Versions[i.Head.string].State.State[digest], virtualFilename)
			i.modified = true
		}
	}

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
	lastVersion := i.getLastVersion()
	if lastVersion == "" {
		return errors.New("cannot get last version")
	}
	i.Head.string = lastVersion
	i.Head.err = nil
	return nil
}
