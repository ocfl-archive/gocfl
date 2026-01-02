package inventory

import (
	"fmt"
	"net/url"
	"regexp"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewUserBase() *userBase {
	return &userBase{
		Address: types.NewOCFLString(""),
		Name:    types.NewOCFLString(""),
	}
}

type userBase struct {
	Address *types.OCFLString
	Name    *types.OCFLString
}

var mailtoUriRegexp = regexp.MustCompile(`mailto:[^@]+@[^@]+`)

func (u *userBase) Check(val validation.Validation, version *types.VersionNumber) error {
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

func (u *userBase) WithAddress(address string) types.User {
	u.Address = types.NewOCFLString(address)
	return u
}

func (u *userBase) WithName(name string) types.User {
	u.Name = types.NewOCFLString(name)
	return u
}

func (u *userBase) Finalize() {
	if u.Name == nil {
		u.Name = types.NewOCFLString("")
	}
	if u.Address == nil {
		u.Address = types.NewOCFLString("")
	}
}

func (u *userBase) Err() error {
	if u == nil {
		return nil
	}
	return errors.Combine(u.Name.Err(), u.Address.Err())
}

func (u *userBase) Equals(other types.User) bool {
	otherU, ok := other.(*userBase)
	if !ok {
		return false
	}
	return u.Address.Equals(otherU.Address) && u.Name.Equals(otherU.Name)
}

func (u *userBase) GetAddress() string {
	return u.Address.String()
}

func (u *userBase) GetName() string {
	return u.Name.String()
}

func (u *userBase) String() string {
	return fmt.Sprintf("%s [%s]", u.Name.String(), u.Address.String())
}

var _ types.User = (*userBase)(nil)
