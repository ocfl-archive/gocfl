package inventory

import (
	"encoding/json"
	"maps"
	"strings"

	"emperror.dev/errors"
)

type VersionsBase struct {
	Versions map[string]*VersionBase
	err      error
}

func (v *VersionsBase) Get(versionString string) (Version, bool) {
	version, ok := v.Versions[versionString]
	return version, ok
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
		otherVersion, ok := otherBase.Get(key)
		if !ok {
			return false
		}

	}
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

func (v *VersionsBase) GetVersion(version string) (*VersionBase, error) {
	ver, ok := v.Versions[version]
	if !ok {
		return nil, errors.Errorf("invalid version '%s'", version)
	}
	return ver, nil
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
