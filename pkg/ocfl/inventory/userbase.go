package inventory

import (
	"fmt"
	"net/url"
	"regexp"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewUserBase(name, address string) *UserBase {
	return &UserBase{
		Address: NewOCFLString(address),
		Name:    NewOCFLString(name),
	}
}

type UserBase struct {
	Address *OCFLString
	Name    *OCFLString
}

var mailtoUriRegexp = regexp.MustCompile(`mailto:[^@]+@[^@]+`)

func (u *UserBase) Check(val validation.Validation, version string) error {
	if u.Address.Err() != nil {
		val.AddValidationError(validation.E054, "invalid user address in Version %s: %s", version, u.Address.Err().Error())
	}
	if u.Name.Err() != nil {
		val.AddValidationWarning(validation.E054, "invalid user name in Version %s: %s", version, u.Name.Err().Error())
	}
	uAddr := u.Address.String()
	if uAddr == "" {
		val.AddValidationWarning(validation.W008, "no user address in version %s", version)
	} else {
		if !mailtoUriRegexp.MatchString(uAddr) {
			u, err := url.Parse(uAddr)
			if err != nil {
				val.AddValidationWarning(validation.W009, "cannot parse user address '%s' in version '%s': %v", uAddr, version, err)
			} else {
				if u.Scheme == "" {
					val.AddValidationWarning(validation.W009, "cannot parse user address '%s' in version '%s'", uAddr, version)
				}
			}
		}
	}
	return nil
}

func (u *UserBase) SetAddress(address string) User {
	u.Address = NewOCFLString(address)
	return u
}

func (u *UserBase) SetName(name string) User {
	u.Name = NewOCFLString(name)
	return u
}

func (u *UserBase) Finalize() {
	if u.Name == nil {
		u.Name = NewOCFLString("")
	}
	if u.Address == nil {
		u.Address = NewOCFLString("")
	}
}

func (u *UserBase) Err() error {
	return errors.Combine(u.Name.Err(), u.Address.Err())
}

func (u *UserBase) Equals(other User) bool {
	otherU, ok := other.(*UserBase)
	if !ok {
		return false
	}
	return u.Address.Equals(otherU.Address) && u.Name.Equals(otherU.Name)
}

func (u *UserBase) GetAddress() (string, error) {
	return u.Address.String(), u.Address.Err()
}

func (u *UserBase) GetName() (string, error) {
	return u.Name.String(), u.Name.Err()
}

func (u *UserBase) String() string {
	return fmt.Sprintf("%s [%s]", u.Name.String(), u.Address.String())
}

var _ User = (*UserBase)(nil)
