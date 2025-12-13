package inventory

import (
	"slices"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type VersionBase struct {
	version string
	Created *OCFLTime   `json:"created"`
	Message *OCFLString `json:"message"`
	State   *StateBase  `json:"state"`
	User    *OCFLUser   `json:"user"`
}

func (v *VersionBase) Finalize(inCreation bool, val validation.Validation) error {
	if v.User == nil {
		_ = val.AddValidationWarning(validation.W007, "no user key in version '%s'", v.version)
		v.User = NewOCFLUser("", "")
	}
	v.User.Finalize()
	if v.Message == nil {
		_ = val.AddValidationWarning(validation.W007, "no message key in version '%s'", v.version)
		v.Message = NewOCFLString("")
	}
	if v.State == nil {
		v.State = &StateBase{
			State: map[string][]string{},
			err:   nil,
		}
	}
	return nil
}

func (v *VersionBase) Equals(other Version) bool {
	otherVersion, ok := other.(*VersionBase)
	if !ok {
		return false
	}
	if !v.EqualMeta(otherVersion) {
		return false
	}
	if !v.EqualState(otherVersion) {
		return false
	}
	return true
}

func (v *VersionBase) String() string {
	return v.version
}

func (v *VersionBase) EqualMeta(v2 *VersionBase) bool {
	if v2 == nil {
		return false
	}
	if v.Created.Time.String() != v2.Created.Time.String() ||
		v.Message.string != v2.Message.string ||
		v.User.Name.String() != v2.User.Name.String() ||
		v.User.Address.String() != v2.User.Address.String() {
		return false
	}
	return true
}
func (v *VersionBase) EqualState(v2 *VersionBase) bool {
	if v2 == nil {
		return false
	}
	/*
		if v.Created.Time.String() != v2.Created.Time.String() ||
			v.Message.string != v2.Message.string ||
			v.User.GetName.string != v2.User.GetName.string ||
			v.User.Address.String() != v2.User.Address.String() {
			return false
		}
	*/
	if len(v.State.State) != len(v2.State.State) {
		return false
	}
	files := []string{}
	for _, vals := range v.State.State {
		files = append(files, vals...)
	}
	slices.Sort(files)
	files = slices.Compact(files)
	files2 := []string{}
	for _, vals := range v2.State.State {
		files2 = append(files2, vals...)
	}
	slices.Sort(files2)
	files2 = slices.Compact(files2)
	if len(files) != len(files2) {
		return false
	}
	if !util.SliceContains(files, files2) {
		return false
	}
	return true
}

var _ Version = (*VersionBase)(nil)
