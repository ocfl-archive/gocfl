package inventoryimpl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewUserBase(logger ocfllogger.OCFLLogger) *userBase {
	return &userBase{
		Address: inventory.NewOCFLString(""),
		Name:    inventory.NewOCFLString(""),
		logger:  logger.With("component", "inventory-user"),
		errs:    []error{},
	}
}

type userBase struct {
	Address *inventory.OCFLString `json:"address"`
	Name    *inventory.OCFLString `json:"name"`
	logger  ocfllogger.OCFLLogger
	errs    []error
}

var mailtoUriRegexp = regexp.MustCompile(`mailto:[^@]+@[^@]+`)

func (u *userBase) Check(version *inventory.VersionNumber) error {
	if u.Address.Err() != nil {
		u.logger.ValidationError(validation.E054, "invalid user address in Version %s: %s", version, u.Address.Err().Error())
	}
	if u.Name.Err() != nil {
		u.logger.ValidationError(validation.E054, "invalid user name in Version %s: %s", version, u.Name.Err().Error())
	}
	uAddr := u.Address.String()
	if uAddr == "" {
		if u.Name.String() == "" {
			u.logger.ValidationError(validation.W007, "no user Version %s", version)
		} else {
			u.logger.ValidationError(validation.W008, "no user address in Version %s", version)
		}
	} else {
		if !mailtoUriRegexp.MatchString(uAddr) {
			url, err := url.Parse(uAddr)
			if err != nil {
				u.logger.ValidationError(validation.W009, "cannot parse user address '%s' in Version %s: %s", uAddr, version, err.Error())
			} else {
				if url.Scheme == "" {
					u.logger.ValidationError(validation.W009, "cannot parse user address '%s' in Version %s", uAddr, version)
				}
			}
		}
	}
	return nil
}

func (u *userBase) WithAddress(address string) inventory.User {
	u.Address = inventory.NewOCFLString(address)
	return u
}

func (u *userBase) WithName(name string) inventory.User {
	u.Name = inventory.NewOCFLString(name)
	return u
}

func (u *userBase) Finalize() {
	if u.Name == nil {
		u.logger.ValidationError(validation.E054, "no user name in Version")
		u.Name = inventory.NewOCFLString("")
	}
	if u.Address == nil {
		u.logger.ValidationError(validation.W008, "no address in Version")
		u.Address = inventory.NewOCFLString("")
	}
	if u.Name.Err() != nil {
		u.logger.ValidationError(validation.E054, "invalid user name in Version: %s", u.Name.Err().Error())
	}
	if u.Address.Err() != nil {
		u.logger.ValidationError(validation.W008, "invalid user address in Version: %s", u.Address.Err().Error())
	}
	uerr := errors.Combine(u.errs...)
	if uerr != nil {
		u.logger.ValidationError(validation.E054, "error decoding user: %s", uerr.Error())
	}
}

func (u *userBase) Err() error {
	if u == nil {
		return nil
	}
	return errors.Combine(u.Name.Err(), u.Address.Err(), errors.Combine(u.errs...))
}

func (u *userBase) Equals(other inventory.User) bool {
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

func (u *userBase) UnmarshalJSON(data []byte) error {
	type userBase2 userBase
	if err := json.Unmarshal(data, (*userBase2)(u)); err != nil {
		u.errs = append(u.errs, errors.Wrapf(err, "cannot unmarshal string '%s'", string(data)))
		return nil
	}
	return nil
}

var _ inventory.User = (*userBase)(nil)
