package inventory

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/uri"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"golang.org/x/exp/slices"
)

type InventoryBase struct {
	factory Factory
	ctx     context.Context
	folder  string
	//object                 ocfl.Object
	version       version.OCFLVersion
	modified      bool
	writeable     bool
	paddingLength int
	//versionValue           map[string]uint
	//fixityDigestAlgorithms []checksum.DigestAlgorithm
	Id               string                   `json:"id"`
	Type             InventorySpec            `json:"type"`
	DigestAlgorithm  checksum.DigestAlgorithm `json:"digestAlgorithm"`
	Head             *VersionNumber           `json:"head"`
	ContentDirectory string                   `json:"contentDirectory,omitempty"`
	Manifest         Manifest                 `json:"manifest,omitempty"`
	Versions         Versions                 `json:"versions"`
	Fixity           Fixity                   `json:"fixity,omitempty"`
	logger           zLogger.ZLogger
}

func newInventoryBase(ctx context.Context, factory Factory, ver version.OCFLVersion, folder string, objectType *url.URL, contentDir string, logger zLogger.ZLogger) (*InventoryBase, error) {
	i := &InventoryBase{
		ctx:     ctx,
		factory: factory,
		//object:                 object,
		version:       ver,
		folder:        folder,
		paddingLength: 0,
		//fixityDigestAlgorithms: []checksum.DigestAlgorithm{},
		Type:             InventorySpec(objectType.String()),
		Head:             NewVersionNumber(),
		ContentDirectory: contentDir,
		Manifest:         nil,
		Versions:         factory.NewVersions(),
		Fixity:           nil,
		logger:           logger,
	}
	return i, nil
}

func (i *InventoryBase) IsEqual(invent Inventory) bool {

	i2, ok := invent.(*InventoryBase)
	if !ok {
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
		if !i.Manifest.Equals(invent.GetManifest()) {
			return false
		}
	}
	if (i.Versions == nil && i2.Versions != nil) || (i.Versions != nil && i2.Versions == nil) {
		return false
	}
	if i.Versions != nil {
		if !i.Versions.Equals(i2.Versions) {
			return false
		}
	}
	return true
}

func (i *InventoryBase) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (err error) {
	i.Id = id
	i.DigestAlgorithm = digest
	i.Fixity = i.factory.NewFixity(fixity)
	return nil
}
func (i *InventoryBase) Finalize(inCreation bool) (err error) {
	if i.Manifest == nil {
		if !inCreation {
			i.AddValidationError(validation.E041, "no manifest in inventory")
		}
		i.Manifest = i.factory.NewManifest()
	}
	if err := i.Manifest.Finalize(i, i.factory, inCreation); err != nil {
		return errors.Wrap(err, "error finalizing manifest")
	}

	if i.Versions == nil {
		if !inCreation {
			i.AddValidationError(validation.E041, "no versions in inventory")
		}
		i.Versions = i.factory.NewVersions()
	}
	if err := i.Versions.Finalize(i, i.factory, inCreation); err != nil {
		return errors.Wrap(err, "error finalizing versions")
	}

	if err := i.Fixity.Finalize(inCreation); err != nil {
		return errors.Wrap(err, "error finalizing fixity")
	}

	if !inCreation {
		if err := i.check(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (i *InventoryBase) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	err := validation.GetValidationError(i.version, errno).AppendDescription(format, a...).AppendDescription("(%s/inventory.json)", i.folder).AppendContext("object '%s'", i.GetID())
	return errors.WithStack(validation.AddValidationErrors(i.ctx, err))
}
func (i *InventoryBase) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	err := validation.GetValidationError(i.version, errno).AppendDescription(format, a...).AppendDescription("(%s/inventory.json)", i.folder).AppendContext("object '%s'", i.GetID())
	return errors.WithStack(validation.AddValidationWarnings(i.ctx, err))
}
func (i *InventoryBase) GetID() string           { return i.Id }
func (i *InventoryBase) GetHead() *VersionNumber { return i.Head }
func (i *InventoryBase) GetSpec() InventorySpec  { return i.Type }

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
func (i *InventoryBase) GetFixityDigestAlgorithm() iter.Seq[checksum.DigestAlgorithm] {
	return i.Fixity.GetDigestAlgorithms()
}
func (i *InventoryBase) IsWriteable() bool { return i.writeable }
func (i *InventoryBase) IsModified() bool  { return i.modified }

func (i *InventoryBase) GetVersionNumbers() []*VersionNumber {
	versionsInt := []int{}
	versionString := map[int]*VersionNumber{}
	for ver := range i.Versions.Iterate() {
		matches := vRegexp.FindStringSubmatch(ver.String())
		if matches == nil {
			return []*VersionNumber{}
		}
		versionInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return []*VersionNumber{}
		}
		versionsInt = append(versionsInt, versionInt)
		versionString[versionInt] = ver
	}

	// sort versions ascending
	sort.Ints(versionsInt)
	var versions = []*VersionNumber{}
	for _, versionInt := range versionsInt {
		versions = append(versions, versionString[versionInt])
	}
	return versions
}
func (i *InventoryBase) GetVersions() map[*VersionNumber]Version {
	var versions = map[*VersionNumber]Version{}
	for versionStr, version := range i.Versions.Iterate() {
		versions[versionStr] = version
	}
	return versions
}

func (i *InventoryBase) GetStateFiles(version *VersionNumber, cs string) ([]string, error) {
	if !version.IsValid() {
		version = i.GetHead()
	}
	ver := i.Versions.GetVersion(version)
	if ver == nil {
		return nil, errors.Errorf("invalid version '%s'", version)
	}
	state := ver.GetState()
	if state == nil {
		return nil, errors.Errorf("failed to get state of '%s'", version)
	}
	files, err := state.GetFiles(cs)
	if err != nil {
		if errors.Is(err, DigestNotFound) {
			return nil, errors.Wrapf(err, "no state for in version %s [%s]", version, cs)
		}
		return nil, errors.Wrapf(err, "failed to get state for in version %s [%s]", version, cs)
	}
	return files, nil
}

func (i *InventoryBase) IterateStateFiles(version *VersionNumber, fn StateFileCallback) error {
	if !version.IsValid() {
		version = i.GetHead()
	}
	ver := i.Versions.GetVersion(version)
	if ver == nil {
		return errors.Errorf("invalid version '%s'", version)
	}
	state := ver.GetState()
	if state == nil {
		return errors.Errorf("cannot get state for '%s'", version)
	}
	for digest, externalNames := range state.IterateFiles() {
		internalNames, err := i.Manifest.GetFiles(digest)
		if err != nil {
			return errors.Wrapf(err, "no manifest for [%s]%v", digest, externalNames)
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
		i.AddValidationError(validation.E036, "invalid field \"id\" for object")
	}
	if i.Id != "" {
		if _, err := uri.Parse(i.Id); err != nil {
			i.AddValidationWarning(validation.W005, "cannot parse uri id '%s': %v", i.Id, err)
		} /* else {
			if u.Scheme == "" {
				i.AddValidationWarning(W005, "id '%s' is not an uri", i.Id)
			}
		}
		*/
	}
	if !i.Head.IsValid() {
		i.AddValidationError(validation.E040, "invalid field \"head\" for object")
	}
	if i.Type == "" {
		i.AddValidationError(validation.E036, "invalid field \"type\" for object")
	}
	if i.DigestAlgorithm == "" {
		i.AddValidationError(validation.E036, "invalid field \"digestAlgorithm\" for object")
	}

	if !slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA512, checksum.DigestSHA256}, i.DigestAlgorithm) {
		i.AddValidationError(validation.E025, "invalid digest algorithm '%s'", i.DigestAlgorithm)
	} else {
		if slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA256}, i.DigestAlgorithm) {
			i.AddValidationError(validation.W004, "digest algorithm '%s' not suggested", i.DigestAlgorithm)
		}
	}

	if i.ContentDirectory != "" {
		if slices.Contains([]string{"", ".", ".."}, i.ContentDirectory) || strings.Contains(i.ContentDirectory, "/") {
			i.AddValidationError(validation.E017, "invalid content directory '%s'", i.ContentDirectory)
		}
	}

	return nil
}
func (i *InventoryBase) checkManifest() error {
	i.logger.Debug().Msgf("[%s] checkManifest", i.GetID())
	defer i.logger.Debug().Msgf("[%s] checkManifest done", i.GetID())
	versionDigests := []string{}
	for versionString, version := range i.Versions.Iterate() {
		state := version.GetState()
		if state == nil {
			return errors.Errorf("cannot get state for version '%s'", versionString)
		}
		for digest, _ := range state.IterateFiles() {
			versionDigests = append(versionDigests, digest)
		}
	}
	slices.Sort(versionDigests)

	digests := []string{}
	allPaths := []string{}
	for digest, paths := range i.Manifest.IterateFiles() {
		//		digest = strings.ToLower(digest)
		if slices.Contains(digests, digest) {
			i.AddValidationError(validation.E096, "manifest digest '%s' is duplicate", digest)
		} else {
			digests = util.SliceInsertSorted(digests, digest)
			//digests = append(digests, digest)
			if !slices.Contains(versionDigests, digest) {
				i.AddValidationError(validation.E107, "digest '%s' does not appear in any version", digest)
			}
		}
		for _, path := range paths {
			//allPaths = sliceInsertSorted(allPaths, path)
			allPaths = append(allPaths, path)
			if path[0] == '/' || path[len(path)-1] == '/' {
				i.AddValidationError(validation.E100, "invalid path '%s' in manifest", path)
			}
			if path == "" {
				i.AddValidationError(validation.E099, "empty path in manifest")
			}
			path2 := path
			if path[0] == '/' {
				path2 = path[1:]
			}
			elements := strings.Split(path2, "/")
			for _, element := range elements {
				if slices.Contains([]string{"", ".", ".."}, element) {
					i.AddValidationError(validation.E099, "invalid path '%s' in manifest", path)
				}
			}

		}

	}
	i.logger.Debug().Msgf("[%s] checkManifest prefix", i.GetID())
	slices.Sort(allPaths)
	for j := 0; j < len(allPaths)-1; j++ {
		prefix := strings.TrimRight(allPaths[j+1], "/") + "/"
		if strings.HasPrefix(allPaths[j], prefix) {
			i.AddValidationError(validation.E101, "content path '%s' is prefix or equal to '%s' in manifest", allPaths[j], prefix)
		}
	}
	return nil
}

func (i *InventoryBase) checkFixity() error {
	i.logger.Debug().Msgf("[%s] checkFixity", i.GetID())
	defer i.logger.Debug().Msgf("[%s] checkFixity done", i.GetID())
	return errors.Wrap(i.Fixity.Check(i, nil), "error checking fixity")
}

func (i *InventoryBase) checkVersions() error {
	i.logger.Debug().Msgf("[%s] checkVersions", i.GetID())
	defer i.logger.Debug().Msgf("[%s] checkVersions done", i.GetID())
	manifestDigests := []string{}
	for mDigest, _ := range i.Manifest.IterateFiles() {
		manifestDigests = append(manifestDigests, mDigest)
	}

	if err := i.Versions.Check(i, manifestDigests); err != nil {
		return errors.Wrap(err, "cannot check versions")
	}
	// check head is recent ver
	var recentVersion *VersionNumber
	for ver := range i.Versions.GetVersionNumbers() {
		if !recentVersion.IsValid() {
			recentVersion = ver
		} else {
			if recentVersion.Less(ver) {
				recentVersion = ver
			}
		}
	}
	if !i.GetHead().Equal(recentVersion) && i.GetHead().IsValid() {
		i.AddValidationError(validation.E040, "manifest head '%s' is not recent ver '%s'", i.GetHead(), recentVersion)
	}

	// check that head exists in versions
	if !i.Head.IsValid() {
		i.AddValidationError(validation.E040, "manifest head '%s' does not exists in versions %v", i.Head.string, i.GetVersionNumbers())
	}

	return nil
}

func (i *InventoryBase) CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error {
	i.logger.Debug().Msgf("[%s] checkFiles", i.GetID())
	defer i.logger.Debug().Msgf("[%s] checkFiles done", i.GetID())
	csFiles, ok := fileManifest[i.GetDigestAlgorithm()]
	if !ok {
		if len(fileManifest) == 0 {
			return nil
		}
		return errors.Errorf("checksum for '%s' not created", i.GetDigestAlgorithm())
	}
	if err := i.Manifest.Check(i, csFiles); err != nil {
		return errors.Wrap(err, "manifest check failed")
	}
	if err := i.Fixity.Check(i, fileManifest); err != nil {
		return errors.Wrap(err, "fixity check failed")
	}
	return nil
}

/*
func (i *InventoryBase) GetFiles() map[*VersionNumber][]string {
	var result = map[*VersionNumber][]string{}
	versions := []*VersionNumber{}
	for _, files := range i.Manifest.IterateFiles() {
		for _, filename := range files {
			parts := strings.Split(filename, "/")
			if len(parts) < 3 {
				i.AddValidationError(validation.E000, "invalid filepath in manifest '%s'", filename)
			}
			version := NewVersionNumber().WithString(parts[0])
			//fn := parts[2]
			if parts[1] != i.GetContentDir() {
				//i.AddValidationError(E015, "extra file/directory '%s' in manifest", parts[1])
				//i.AddValidationError(E019, "invalid content directory '%s' in '%s'", parts[1], filename)
			}
			if _, ok := result[version]; !ok {
				versions = append(versions, version)
				result[version] = []string{}
			}
			result[version] = append(result[version], filename)
		}
	}
	iVersions := i.GetVersionNumbers()
	if !util.SliceContains(iVersions, versions) {
		slices.Sort(iVersions)
		i.AddValidationError(validation.E023, "versions %v do not contains versions from manifest %v", iVersions, versions)
	}
	return result
}

*/

func (i *InventoryBase) GetManifest() Manifest {
	return i.Manifest
}

func (i *InventoryBase) GetFixity() Fixity {
	/*
		if i.Fixity == nil {
			return map[checksum.DigestAlgorithm]map[string][]string{}
		}

	*/
	return i.Fixity
}

func (i *InventoryBase) BuildManifestName(stateFilename string) string {
	return i.BuildManifestNameVersion(stateFilename, i.GetHead())
}

func (i *InventoryBase) BuildManifestNameVersion(stateFilename string, version *VersionNumber) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(version.String(), i.GetContentDir(), stateFilename)))
}

func (i *InventoryBase) NewVersion(msg, UserName, UserAddress string) error {
	/*
		if i.IsWriteable() {
			return errors.New(fmt.Sprintf("version '%s' already writeable", i.GetHead()))
		}
	*/
	lastHead := i.Head
	if !lastHead.IsValid() {
		if i.paddingLength <= 0 {
			i.Head.WithString("v1")
		} else {
			i.Head.WithString(fmt.Sprintf(fmt.Sprintf("v0%%0%dd", i.paddingLength), 1))
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
	ver := i.factory.NewVersion()
	state := i.factory.NewState()
	user := i.factory.NewUser().WithAddress(UserAddress).WithName(UserName)
	ver.WithCreated(time.Now()).WithMessage(msg).WithState(state).WithUser(user)
	if !lastHead.IsValid() {
		lastVersion := i.Versions.GetVersion(lastHead)
		if lastVersion == nil {
			return errors.Errorf("cannot determine version - '%s'", lastHead)
		}
		lastState := lastVersion.GetState()
		if lastState == nil {
			return errors.Errorf("cannot determine state - '%s'", lastHead)
		}
		newState := i.factory.NewState()
		newState.CopyFrom(lastState)
		ver.WithState(newState)
	}
	i.Versions.SetVersion(i.Head, ver)
	i.writeable = true
	return nil
}

var vRegexp *regexp.Regexp = regexp.MustCompile("^v(\\d+)$")

func (i *InventoryBase) AlreadyExists(stateFilename, checksum string) (bool, error) {
	return i.Versions.FileExists(stateFilename, checksum)
}

func (i *InventoryBase) IsUpdate(virtualFilename, checksum string) (bool, error) {
	found, err := i.Versions.FileExists(virtualFilename, checksum)
	return !found, errors.WithStack(err)
}

func (i *InventoryBase) EchoDelete(existing []string, pathPrefix string) error {
	modified, err := i.Versions.EchoDelete(existing, pathPrefix)
	if err != nil {
		return errors.WithStack(err)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) DeleteFile(stateFilename string) error {
	modified, err := i.Versions.DeleteFile(stateFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", stateFilename)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) RenameFile(stateSource, stateDest string) error {
	modified, err := i.Versions.RenameFile(stateSource, stateDest)
	if err != nil {
		return errors.Wrapf(err, "cannot rename '%s' to '%s'", stateSource, stateDest)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) CopyFile(dest string, digest string) error {
	i.logger.Info().Msgf("[%s] copying '%s' -> '%s'", i.GetID(), digest, dest)

	modified, err := i.Versions.CopyFile(dest, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot copy '%s' to '%s'", digest, dest)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error {
	i.logger.Debug().Msgf("[%s] adding '%s' -> '%s'", i.GetID(), stateFilenames, manifestFilename)
	digest, ok := checksums[i.GetDigestAlgorithm()]
	if !ok {
		return errors.Errorf("no digest for '%s' in checksums", i.GetDigestAlgorithm())
	}
	digest = strings.ToLower(digest) // paranoia

	fixitydigests := map[checksum.DigestAlgorithm]string{}
	for alg, cs := range checksums {
		if alg == i.GetDigestAlgorithm() {
			continue
		}
		fixitydigests[alg] = cs
	}
	modified, err := i.Fixity.AddFile(manifestFilename, fixitydigests)
	if err != nil {
		return errors.Wrapf(err, "cannot add fixity '%s' to '%s'", digest, manifestFilename)
	}
	i.modified = i.modified || modified

	if manifestFilename != "" {
		modified, err := i.Manifest.AddFile(manifestFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add manifest '%s' to '%s'", digest, manifestFilename)
		}
		i.modified = i.modified || modified
	}

	for _, virtualFilename := range stateFilenames {
		dup, err := i.AlreadyExists(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add for duplicate of '%s' [%s]", stateFilenames, digest)
		}
		if dup {
			i.logger.Debug().Msgf("'%s' is a duplicate", stateFilenames)
			// return nil
		}

		modfied, err := i.Versions.AddFile(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add state '%s' to '%s'", digest, virtualFilename)
		}
		i.modified = i.modified || modfied

		upd, err := i.IsUpdate(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check for update of '%s' [%s]", stateFilenames, digest)
		}
		if upd {
			i.logger.Debug().Msgf("'%s' is an update - removing old version", stateFilenames)
			if err := i.DeleteFile(virtualFilename); err != nil {
				return errors.Wrapf(err, "cannot delete old version of '%s' [%s]", stateFilenames, digest)
			}
			i.modified = true
		}

		if !dup {
			modified, err := i.Versions.AddFile(virtualFilename, digest)
			if err != nil {
				return errors.Wrapf(err, "cannot add version of '%s' [%s]", stateFilenames, digest)
			}
			i.modified = i.modified || modified
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
	if i.GetHead().String() == "v1" {
		return nil
	}
	i.logger.Debug().Msgf("deleting %v", i.GetHead())
	i.Versions.Delete(i.GetHead())
	lastVersion := i.Versions.LatestVersion()
	if !lastVersion.IsValid() {
		return errors.New("cannot get last version")
	}
	i.Head = lastVersion
	return nil
}

var _ Inventory = (*InventoryBase)(nil)
var _ validation.Validation = (*InventoryBase)(nil)
