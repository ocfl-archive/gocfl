package inventory

import (
	"time"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/interfaces"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewVersionBase(factory interfaces.Factory) interfaces.Version {
	return &versionBase{
		Created: NewOCFLTime(time.Now()),
		Message: NewOCFLString("initial"),
		State:   factory.NewState(),
		User:    factory.NewUser(),
	}
}

type versionBase struct {
	version *interfaces.VersionNumber
	Created *OCFLTime        `json:"created"`
	Message *OCFLString      `json:"message"`
	State   interfaces.State `json:"state"`
	User    interfaces.User  `json:"user"`
}

func (v *versionBase) SetVersion(number *interfaces.VersionNumber) {
	v.version = number
}

func (v *versionBase) AddFile(stateFilename string, digest string) (bool, error) {
	return v.State.AddFile(stateFilename, digest)
}

func (v *versionBase) EchoDelete(existing []string, pathPrefix string) (bool, error) {
	modified, err := v.State.EchoDelete(existing, pathPrefix)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionBase) CopyFile(stateFilename, digest string) (bool, error) {
	modified, err := v.State.CopyFile(stateFilename, digest)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionBase) RenameFile(oldStateFilename, newStateFilename string) (bool, error) {
	modified, err := v.State.RenameFile(oldStateFilename, newStateFilename)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionBase) DeleteFile(stateFilename string) (bool, error) {
	modified, err := v.State.DeleteFile(stateFilename)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionBase) FileChecksum(path string) string {
	return v.State.FileChecksum(path)
}

func (v *versionBase) Check(val validation.Validation, manifestDigests, manifestDigestsLower []string) error {
	if v.Created.Err() != nil {
		val.AddValidationError(validation.E049, "invalid created format in version '%s': %v", v.version, v.Created.Err())
	}
	if err := v.User.Check(val, v.version); err != nil {
		return errors.Wrapf(err, "cannot check version %s", v.version)
	}
	if v.Message.Err() != nil {
		val.AddValidationError(validation.E094, "invalid format for message in version '%s': %v", v.version, v.Message.Err().Error())
	}
	if v.State == nil {
		return errors.Errorf("no state set for version '%s'", v.version)
	}
	if err := v.State.Check(val, v.version, manifestDigests, manifestDigestsLower); err != nil {
		return errors.Wrapf(err, "invalid state check for version '%s'", v.version)
	}
	return nil
}

func (v *versionBase) Err() error {
	if v == nil {
		return nil
	}
	return errors.Combine(
		v.User.Err(),
		v.State.Err(),
		v.Message.Err(),
		v.Created.Err(),
	)
}

func (v *versionBase) WithCreated(t time.Time) interfaces.Version {
	v.Created = NewOCFLTime(t)
	return v
}

func (v *versionBase) WithMessage(msg string) interfaces.Version {
	v.Message = NewOCFLString(msg)
	return v
}

func (v *versionBase) WithState(state interfaces.State) interfaces.Version {
	stateB, ok := state.(*stateBase)
	if !ok {
		panic("invalid state type")
		return v
	}
	v.State = stateB
	return v
}

func (v *versionBase) WithUser(user interfaces.User) interfaces.Version {
	userB, ok := user.(*userBase)
	if !ok {
		panic("invalid user type")
		return v
	}
	v.User = userB
	return v
}

func (v *versionBase) GetMessage() string {
	return v.Message.String()
}

func (v *versionBase) GetUser() interfaces.User {
	return v.User
}

func (v *versionBase) GetCreated() time.Time {
	return v.Created.Time
}

func (v *versionBase) GetState() interfaces.State {
	return v.State
}

func (v *versionBase) Finalize(val validation.Validation, factory interfaces.Factory, inCreation bool) error {
	if v.User == nil {
		_ = val.AddValidationWarning(validation.W007, "no user key in version '%s'", v.version)
		v.User = NewUserBase()
	}
	v.User.Finalize()
	if v.Message == nil {
		_ = val.AddValidationWarning(validation.W007, "no message key in version '%s'", v.version)
		v.Message = NewOCFLString("")
	}
	if v.State == nil {
		v.State = NewStateBase()
	}
	return nil
}

func (v *versionBase) Equals(other interfaces.Version) bool {
	if v == nil || other == nil {
		return false
	}
	otherVersion, ok := other.(*versionBase)
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

func (v *versionBase) String() string {
	return v.version.String()
}

func (v *versionBase) EqualMeta(v2 *versionBase) bool {
	if v == nil || v2 == nil {
		return false
	}
	if v.Created.Time.String() != v2.Created.Time.String() ||
		v.Message.string != v2.Message.string ||
		!v.User.Equals(v2.User) {
		return false
	}
	return true
}
func (v *versionBase) EqualState(v2 *versionBase) bool {
	if v == nil || v2 == nil {
		return false
	}
	if !v.State.Equals(v2.State) {
		return false
	}
	return true
}

var _ interfaces.Version = (*versionBase)(nil)
