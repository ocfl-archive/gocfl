package inventory

import (
	"time"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type VersionBase struct {
	version string
	Created *OCFLTime   `json:"created"`
	Message *OCFLString `json:"message"`
	State   State       `json:"state"`
	User    User        `json:"user"`
}

func (v *VersionBase) Check(val validation.Validation, manifestDigests, manifestDigestsLower []string) error {
	if v.Created.Err() != nil {
		val.AddValidationError(validation.E049, "invalid created format in version '%s': %v", v.version, v.Created.Err())
	}
	if err := v.User.Check(val, v.version); err != nil {
		return errors.Wrapf(err, "cannot check version %s", v.version)
	}
	if v.Message.Err() != nil {
		val.AddValidationError(validation.E094, "invalid format for message in version '%s': %v", v.version, v.Message.Err().Error())
	}
	return nil
}

func (v *VersionBase) Err() error {
	return errors.Combine(
		v.User.Err(),
		v.State.Err(),
		v.Message.Err(),
		v.Created.Err(),
	)
}

func (v *VersionBase) SetCreated(t time.Time) Version {
	v.Created = NewOCFLTime(t)
	return v
}

func (v *VersionBase) SetMessage(msg string) Version {
	v.Message = NewOCFLString(msg)
	return v
}

func (v *VersionBase) SetState(state State) Version {
	v.State = state
	return v
}

func (v *VersionBase) SetUser(user User) Version {
	v.User = user
	return v
}

func (v *VersionBase) GetMessage() string {
	return v.Message.String()
}

func (v *VersionBase) GetUser() User {
	return v.User
}

func (v *VersionBase) GetCreated() time.Time {
	return v.Created.Time
}

func (v *VersionBase) GetState() State {
	return v.State
}

func (v *VersionBase) Finalize(val validation.Validation, factory Factory, inCreation bool) error {
	if v.User == nil {
		_ = val.AddValidationWarning(validation.W007, "no user key in version '%s'", v.version)
		v.User = factory.NewUser()
	}
	v.User.Finalize()
	if v.Message == nil {
		_ = val.AddValidationWarning(validation.W007, "no message key in version '%s'", v.version)
		v.Message = NewOCFLString("")
	}
	if v.State == nil {
		v.State = factory.NewState()
	}
	return nil
}

func (v *VersionBase) Equals(other Version) bool {
	if v == nil || other == nil {
		return false
	}
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
	if v == nil || v2 == nil {
		return false
	}
	if v.Created.Time.String() != v2.Created.Time.String() ||
		v.Message.string != v2.Message.string ||
		v.User.Equals(v2.User) {
		return false
	}
	return true
}
func (v *VersionBase) EqualState(v2 *VersionBase) bool {
	if v == nil || v2 == nil {
		return false
	}
	if !v.State.Equals(v2.State) {
		return false
	}
	return true
}

var _ Version = (*VersionBase)(nil)
