package inventory

import (
	"encoding/json"
	"fmt"
	"iter"
	"regexp"
	"sort"
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
		versions:    map[int]Version{},
		versionInts: map[string]int{},
		factory:     factory,
	}
}

type versionsBase struct {
	versions      map[int]Version
	versionInts   map[string]int
	err           error
	paddingLength int
	factory       Factory
}

func (v *versionsBase) Err() error {
	if v == nil {
		return nil
	}
	var errs = []error{}
	for _, version := range v.Iterate() {
		errs = append(errs, version.Err())
	}
	return errors.Combine(errs...)
}

func (v *versionsBase) Delete(versionNumber *VersionNumber) (bool, error) {
	if _, ok := v.versions[versionNumber.Int()]; !ok {
		return false, nil
	}
	delete(v.versions, versionNumber.Int())
	delete(v.versionInts, versionNumber.String())
	return true, nil
}

func (v *versionsBase) AddFile(stateFilename string, digest string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
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
	latestVersionString := v.LatestVersionNumber()
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
	latestVersionString := v.LatestVersionNumber()
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
	latestVersionString := v.LatestVersionNumber()
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
	latestVersionString := v.LatestVersionNumber()
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
	css := map[int]string{}
	for versionString, ver := range v.Iterate() {
		cs := ver.FileChecksum(path)
		if cs == "" {
			continue
		}
		css[versionString.Int()] = cs
	}
	if len(css) == 0 {
		return false, nil
	}
	latestVersion := v.LatestVersionNumber()

	lastChecksum, ok := css[latestVersion.Int()]
	if !ok {
		return false, errors.Errorf("checksum for version %s does not exist", lastChecksum)
	}

	// check whether the latest version is the correct file
	return lastChecksum == digest, nil
}

func (v *versionsBase) GetVersionNumbers() iter.Seq[*VersionNumber] {
	return func(yield func(*VersionNumber) bool) {
		for versionString, versionInt := range v.versionInts {
			if !yield(NewVersionNumber().Init(versionInt, versionString)) {
				return
			}
		}
	}
}

func (v *versionsBase) Finalize(val validation.Validation, factory Factory, inCreation bool) error {
	for ver, version := range v.Iterate() {
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
	var versionsSeq = []int{}
	var paddingLength int = -1
	manifestDigestsLower := []string{}
	for _, manifestDigest := range manifestDigests {
		manifestDigestsLower = append(manifestDigestsLower, strings.ToLower(manifestDigest))
	}
	slices.Sort(manifestDigests)
	slices.Sort(manifestDigestsLower)
	for versionNumber, ver := range v.Iterate() {
		versionsSeq = append(versionsSeq, versionNumber.Int())
		if versionZeroRegexp.MatchString(versionNumber.String()) {
			if paddingLength == -1 {
				paddingLength = len(versionNumber.String()) - 2
			} else {
				if paddingLength != len(versionNumber.String())-2 {
					//i.AddValidationError(E011, "invalid ver padding '%s'", ver)
					val.AddValidationError(validation.E012, "invalid ver padding '%s'", versionNumber)
					val.AddValidationError(validation.E013, "invalid ver padding '%s'", versionNumber)
				}
			}
		} else {
			if versionNoZeroRegexp.MatchString(versionNumber.String()) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						val.AddValidationError(validation.E011, "invalid ver padding '%s'", versionNumber)
						val.AddValidationError(validation.E012, "invalid ver padding '%s'", versionNumber)
						val.AddValidationError(validation.E013, "invalid ver padding '%s'", versionNumber)
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				val.AddValidationError(validation.E104, "invalid version format '%s'", versionNumber)
			}
		}

		if err := ver.Check(val, manifestDigests, manifestDigestsLower); err != nil {
			return errors.Wrapf(err, "version %s validation check failed", versionNumber)
		}
	}
	slices.Sort(versionsSeq)
	for key, vInt := range versionsSeq {
		if key != vInt-1 {
			val.AddValidationError(validation.E010, "invalid ver sequence %v", versionsSeq)
			break
		}
	}
	v.paddingLength = paddingLength
	if paddingLength > 0 {
		val.AddValidationWarning(validation.W001, "padding length is %v", paddingLength)
	}

	return nil
}

func (v *versionsBase) SetVersion(versionNumber *VersionNumber, ver Version) Versions {
	verB, ok := ver.(*versionBase)
	if !ok {
		panic(fmt.Sprintf("cannot convert to versionBase '%s'", versionNumber))
	}
	v.versions[versionNumber.Int()] = verB
	v.versionInts[versionNumber.String()] = versionNumber.Int()
	return v
}

func (v *versionsBase) IsEmpty() bool {
	return len(v.versions) == 0
}

func (v *versionsBase) GetVersion(versionNumber *VersionNumber) Version {
	if versionNumber.IsLatest() {
		versionNumber = v.LatestVersionNumber()
	}
	version, ok := v.versions[versionNumber.Int()]
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
	if len(v.versions) != len(otherBase.versions) {
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
	for k := range v.GetVersionNumbers() {
		result += "; " + k.String()
	}
	return strings.TrimLeft(result, "; ")
}

func (v *versionsBase) Iterate() func(yield func(versionNumber *VersionNumber, version Version) bool) {
	return func(yield func(versionString *VersionNumber, version Version) bool) {
		for versionNumber := range v.GetVersionNumbers() {
			if !yield(versionNumber, v.versions[versionNumber.Int()]) {
				return
			}
		}
	}
}

func (v *versionsBase) UnmarshalJSON(data []byte) error {
	var newVersions = map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &newVersions); err != nil {
		v.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}
	v.versions = map[int]Version{}
	for key, bytes := range newVersions {
		versionNumber := NewVersionNumber().WithString(key)
		ver := v.factory.NewVersion()
		if err := json.Unmarshal(bytes, &ver); err != nil {
			v.err = errors.Wrapf(err, "cannot unmarshal version '%s': '%s'", versionNumber.String(), string(bytes))
			return nil
		}
		ver.SetVersion(versionNumber)
		v.versions[versionNumber.Int()] = ver
		v.versionInts[versionNumber.String()] = versionNumber.Int()
	}
	return nil
}

func (v *versionsBase) MarshalJSON() ([]byte, error) {
	var newVersions = map[string]Version{}
	for versionNumber := range v.GetVersionNumbers() {
		newVersions[versionNumber.String()] = v.versions[versionNumber.Int()]
	}

	return json.Marshal(newVersions)
}

func (v *versionsBase) GetVersionNumber(vInt int) *VersionNumber {
	for versionNumber := range v.GetVersionNumbers() {
		if versionNumber.Int() == vInt {
			return versionNumber
		}
	}
	return nil
}

func (v *versionsBase) LatestVersionNumber() *VersionNumber {
	var versions = make([]int, 0, len(v.versions))
	for versionInt, _ := range v.versions {
		versions = append(versions, versionInt)
	}

	// get the latest version with this path
	sort.Ints(versions)
	lastVersionInt := versions[len(versions)-1]
	return v.GetVersionNumber(lastVersionInt)
}

var _ Versions = (*versionsBase)(nil)
