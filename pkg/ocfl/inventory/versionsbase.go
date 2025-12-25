package inventory

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"golang.org/x/exp/slices"
)

func NewVersionsBase() *VersionsBase {
	return &VersionsBase{
		Versions:     map[string]*VersionBase{},
		versionValue: map[string]uint{},
	}
}

type VersionsBase struct {
	Versions     map[string]*VersionBase
	versionValue map[string]uint
	err          error
}

func (v *VersionsBase) Finalize(val validation.Validation, factory Factory, inCreation bool) error {
	for ver, version := range v.Iterate() {
		vInt, err := strconv.Atoi(strings.TrimLeft(ver, "v0"))
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

func (v *VersionsBase) Check(val validation.Validation, manifestDigests []string) error {
	manifestDigestsLower := []string{}
	for _, manifestDigest := range manifestDigests {
		manifestDigestsLower = append(manifestDigestsLower, strings.ToLower(manifestDigest))
	}
	slices.Sort(manifestDigests)
	slices.Sort(manifestDigestsLower)
	for versionString, ver := range v.Iterate() {
		if err := ver.Check(val, manifestDigests, manifestDigestsLower); err != nil {
			return errors.Wrapf(err, "version %s validation check failed", versionString)
		}
	}
	return nil
}

func (v *VersionsBase) SetVersion(versionString string, ver Version) Versions {
	verB, ok := ver.(*VersionBase)
	if !ok {
		panic(fmt.Sprintf("cannot convert to VersionBase '%s'", versionString))
	}
	v.Versions[versionString] = verB
	return v
}

func (v *VersionsBase) IsEmpty() bool {
	return len(v.Versions) == 0
}

func (v *VersionsBase) GetVersion(versionString string) Version {
	version, ok := v.Versions[versionString]
	if !ok {
		return nil
	}
	return version
}

func (v *VersionsBase) Equals(other Versions) bool {
	otherBase, ok := other.(*VersionsBase)
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

func (v *VersionsBase) String() string {
	var result string
	for k := range maps.Keys(v.Versions) {
		result += "; " + k
	}
	return strings.TrimLeft(result, "; ")
}

func (v *VersionsBase) Iterate() func(yield func(versionString string, version Version) bool) {
	return func(yield func(versionString string, version Version) bool) {
		for versionString, version := range v.Versions {
			if !yield(versionString, version) {
				return
			}
		}
	}
}

func (v *VersionsBase) UnmarshalJSON(data []byte) error {
	v.Versions = map[string]*VersionBase{}
	if err := json.Unmarshal(data, &v.Versions); err != nil {
		v.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}

	return nil
}

func (v *VersionsBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Versions)
}

var _ Versions = (*VersionsBase)(nil)
