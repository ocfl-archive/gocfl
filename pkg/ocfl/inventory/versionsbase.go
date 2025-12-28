package inventory

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"golang.org/x/exp/slices"
)

var versionZeroRegexp = regexp.MustCompile("^v0[0-9]+$")
var versionNoZeroRegexp = regexp.MustCompile("^v[1-9][0-9]*$")

type versionValue map[*VersionNumber]uint

func (v versionValue) GetInt(version *VersionNumber) (uint, bool) {
	vInt, ok := v[version]
	return vInt, ok
}

func (v versionValue) GetVersion(version uint) (*VersionNumber, bool) {
	for ver, verInt := range v {
		if verInt == version {
			return ver, true
		}
	}
	return nil, false
}

func NewVersionsBase(factory Factory) Versions {
	return &versionsBase{
		Versions:      map[*VersionNumber]Version{},
		versionValue:  versionValue{},
		latestVersion: NewVersionNumber(),
	}
}

type versionsBase struct {
	Versions      map[*VersionNumber]Version
	versionValue  versionValue
	err           error
	paddingLength int
	latestVersion *VersionNumber
}

func (v *versionsBase) AddVersion() error {
	//TODO implement me
	panic("implement me")
}

func (v *versionsBase) Delete(versionString *VersionNumber) (bool, error) {
	if _, ok := v.Versions[versionString]; !ok {
		return false, nil
	}
	delete(v.Versions, versionString)
	v.latestVersion = nil
	return true, nil
}

func (v *versionsBase) AddFile(stateFilename string, digest string) (bool, error) {
	latestVersionString := v.LatestVersion()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.AddFile(stateFilename, digest)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionsBase) EchoDelete(existing []string, pathPrefix string) (bool, error) {
	latestVersionString := v.LatestVersion()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.EchoDelete(existing, pathPrefix)
	if err != nil {
		return modified, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionsBase) CopyFile(stateFilename, digest string) (bool, error) {
	latestVersionString := v.LatestVersion()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.CopyFile(stateFilename, digest)
	if err != nil {
		return false, errors.Wrapf(err, "failed to copy file with digest '%s' to '%s' from version %s", digest, stateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) RenameFile(oldStateFilename, newStateFilename string) (bool, error) {
	latestVersionString := v.LatestVersion()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.RenameFile(oldStateFilename, newStateFilename)
	if err != nil {
		return false, errors.Wrapf(err, "failed to rename file '%s' to '%s' from version %s", oldStateFilename, newStateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) DeleteFile(stateFilename string) (bool, error) {
	latestVersionString := v.LatestVersion()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.DeleteFile(stateFilename)
	if err != nil {
		return false, errors.Wrapf(err, "failed to delete file %s from version %s", stateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) FileExists(path, digest string) (bool, error) {
	// find all versions, which contain the path
	css := map[*VersionNumber]string{}
	for versionString, ver := range v.Iterate() {
		cs := ver.FileChecksum(path)
		if cs == "" {
			continue
		}
		css[versionString] = cs
	}
	if len(css) == 0 {
		return false, nil
	}
	var versions = make([]int, 0, len(css))
	for versionString := range css {
		versionInt, ok := v.versionValue[versionString]
		if !ok {
			return false, errors.Errorf("version %s does not exist", versionString)
		}
		versions = append(versions, int(versionInt))
	}

	// get the latest version with this path
	sort.Ints(versions)
	lastVersion := versions[len(versions)-1]

	lastVersionString, ok := v.versionValue.GetVersion(uint(lastVersion))
	if !ok {
		return false, errors.Errorf("version %d does not exist", lastVersion)
	}
	lastChecksum, ok := css[lastVersionString]
	if !ok {
		return false, errors.Errorf("checksum for version %s does not exist", lastVersionString)
	}

	// check whether the latest version is the correct file
	return lastChecksum == digest, nil
}

func (v *versionsBase) LatestVersion() *VersionNumber {
	if !v.latestVersion.IsValid() {
		if len(v.Versions) == 0 {
			return nil
		}
		// sort versions ascending
		var versions = []int{}
		for intVer := range maps.Values(v.versionValue) {
			versions = append(versions, int(intVer))
		}
		sort.Ints(versions)
		lastVersion := uint(versions[len(versions)-1])
		for versionString, intVer := range v.versionValue {
			if intVer == lastVersion {
				v.latestVersion = versionString
				break
			}
		}
	}
	return v.latestVersion
}

func (v *versionsBase) GetVersionStrings() iter.Seq[*VersionNumber] {
	return maps.Keys(v.Versions)
}

func (v *versionsBase) VersionLessOrEqual(v1, v2 *VersionNumber) bool {
	v1Int, ok := v.versionValue[v1]
	if !ok {
		return false
	}
	v2Int, ok := v.versionValue[v2]
	if !ok {
		return false
	}
	return v1Int <= v2Int
}

func (v *versionsBase) Finalize(val validation.Validation, factory Factory, inCreation bool) error {
	for ver, version := range v.Iterate() {
		vInt, err := strconv.Atoi(strings.TrimLeft(ver.String(), "v0"))
		if err != nil {
			val.AddValidationError(validation.E104, "invalid version format '%s'", ver)
			continue
		}
		v.versionValue[ver] = uint(vInt)
		if err := version.Finalize(val, factory, inCreation); err != nil {
			return errors.Wrapf(err, "failed to finalize inventory version '%s'", ver)
		}
	}
	return nil
}

func (v *versionsBase) Check(val validation.Validation, manifestDigests []string) error {
	if v.IsEmpty() {
		val.AddValidationError(validation.E008, "length of ver is 0")
		return nil
	}
	var versions = []int{}
	var paddingLength int = -1
	manifestDigestsLower := []string{}
	for _, manifestDigest := range manifestDigests {
		manifestDigestsLower = append(manifestDigestsLower, strings.ToLower(manifestDigest))
	}
	slices.Sort(manifestDigests)
	slices.Sort(manifestDigestsLower)
	for versionString, ver := range v.Iterate() {
		vInt, ok := v.versionValue[versionString]
		if !ok {
			//			i.AddValidationError(E104, "invalid ver format '%s'", ver)
			continue
		}
		versions = append(versions, int(vInt))
		if versionZeroRegexp.MatchString(versionString.String()) {
			if paddingLength == -1 {
				paddingLength = len(versionString.String()) - 2
			} else {
				if paddingLength != len(versionString.String())-2 {
					//i.AddValidationError(E011, "invalid ver padding '%s'", ver)
					val.AddValidationError(validation.E012, "invalid ver padding '%s'", versionString)
					val.AddValidationError(validation.E013, "invalid ver padding '%s'", versionString)
				}
			}
		} else {
			if versionNoZeroRegexp.MatchString(versionString.String()) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						val.AddValidationError(validation.E011, "invalid ver padding '%s'", versionString)
						val.AddValidationError(validation.E012, "invalid ver padding '%s'", versionString)
						val.AddValidationError(validation.E013, "invalid ver padding '%s'", versionString)
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				val.AddValidationError(validation.E104, "invalid version format '%s'", versionString)
			}
		}

		if err := ver.Check(val, manifestDigests, manifestDigestsLower); err != nil {
			return errors.Wrapf(err, "version %s validation check failed", versionString)
		}
	}
	slices.Sort(versions)
	for key, vInt := range versions {
		if key != vInt-1 {
			val.AddValidationError(validation.E010, "invalid ver sequence %v", versions)
			break
		}
	}
	v.paddingLength = paddingLength
	if paddingLength > 0 {
		val.AddValidationWarning(validation.W001, "padding length is %v", paddingLength)
	}

	return nil
}

func (v *versionsBase) SetVersion(versionString *VersionNumber, ver Version) Versions {
	verB, ok := ver.(*versionBase)
	if !ok {
		panic(fmt.Sprintf("cannot convert to versionBase '%s'", versionString))
	}
	v.Versions[versionString] = verB
	return v
}

func (v *versionsBase) IsEmpty() bool {
	return len(v.Versions) == 0
}

func (v *versionsBase) GetVersion(versionString *VersionNumber) Version {
	version, ok := v.Versions[versionString]
	if !ok {
		return nil
	}
	return version
}

func (v *versionsBase) Equals(other Versions) bool {
	otherBase, ok := other.(*versionsBase)
	if !ok {
		return false
	}
	if len(v.Versions) != len(otherBase.Versions) {
		return false
	}
	for key, version := range v.Iterate() {
		otherVersion := otherBase.GetVersion(key)
		if otherVersion.Err() != nil {
			return false
		}
		return version.Equals(otherVersion)
	}
	return false
}

func (v *versionsBase) String() string {
	var result string
	for k := range maps.Keys(v.Versions) {
		result += "; " + k.String()
	}
	return strings.TrimLeft(result, "; ")
}

func (v *versionsBase) Iterate() func(yield func(versionString *VersionNumber, version Version) bool) {
	return func(yield func(versionString *VersionNumber, version Version) bool) {
		for versionString, version := range v.Versions {
			if !yield(versionString, version) {
				return
			}
		}
	}
}

func (v *versionsBase) UnmarshalJSON(data []byte) error {
	v.Versions = map[*VersionNumber]Version{}
	if err := json.Unmarshal(data, &v.Versions); err != nil {
		v.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}

	return nil
}

func (v *versionsBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Versions)
}

var _ Versions = (*versionsBase)(nil)
