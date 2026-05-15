package inventory

import (
	"encoding/json"
	"time"

	"emperror.dev/errors"
)

// InventorySpec represents the OCFL inventory specification version URL.
type InventorySpec string

const (
	// InventorySpec1_0 is the URL for the OCFL 1.0 inventory specification.
	InventorySpec1_0 InventorySpec = "https://ocfl.io/1.0/spec/#inventory"
	// InventorySpec1_1 is the URL for the OCFL 1.1 inventory specification.
	InventorySpec1_1 InventorySpec = "https://ocfl.io/1.1/spec/#inventory"
	// InventorySpec2_0 is the URL for the OCFL 2.0 inventory specification.
	InventorySpec2_0 InventorySpec = "https://ocfl.io/2.0/spec/#inventory"
)

// SpecIsLessOrEqual returns true if Specification s1 <= s2.
func SpecIsLessOrEqual(s1, s2 InventorySpec) bool {
	//return s1 == InventorySpec1_0 && s2 == InventorySpec1_1
	return s1 <= s2
}

// OCFLString is a helper type for string values in OCFL that can carry an error.
type OCFLString struct {
	string
	err error
}

// NewOCFLString creates a new OCFLString.
func NewOCFLString(str string) *OCFLString {
	return &OCFLString{
		string: str,
		err:    nil,
	}
}

func (s *OCFLString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		s.err = errors.Wrapf(err, "cannot unmarshal string '%s'", string(data))
		return nil
	}
	s.string = str
	return nil
}

func (s *OCFLString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.string)
}

func (s *OCFLString) String() string {
	return s.string
}

func (s *OCFLString) Equals(s2 *OCFLString) bool {
	if s.string != s2.String() {
		return false
	}
	return errors.Is(s.err, s2.err)
}

func (s *OCFLString) Err() error {
	if s == nil {
		return nil
	}
	return s.err
}

// OCFLTime is a helper type for time values in OCFL that can carry an error.
// It handles the RFC3339 format required by the OCFL specification.
type OCFLTime struct {
	time.Time
	err error
}

// NewOCFLTime creates a new OCFLTime from a time.Time object, truncated to seconds.
func NewOCFLTime(t time.Time) *OCFLTime {
	return &OCFLTime{
		Time: t.UTC().Truncate(time.Second),
	}
}

func (t *OCFLTime) MarshalJSON() ([]byte, error) {
	tstr := t.Format(time.RFC3339)
	return json.Marshal(tstr)
}
func (t *OCFLTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		t.err = errors.Wrapf(err, "cannot unmarshal time '%s'", string(data))
		return nil
	}
	tt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.err = errors.Wrapf(err, "cannot parse time %s", string(data))
		return nil
	}
	t.Time = tt

	return nil
}

func (t *OCFLTime) Err() error {
	if t == nil {
		return nil
	}
	return t.err
}

type _User struct {
	Address *OCFLString `json:"address,omitempty"`
	Name    *OCFLString `json:"name"`
}

// OCFLUser represents a user (author) in an OCFL version.
type OCFLUser struct {
	_User
	err error
}

// NewOCFLUser creates a new OCFLUser with the given name and address.
func NewOCFLUser(name, address string) *OCFLUser {
	user := &OCFLUser{
		_User: _User{
			Address: NewOCFLString(address),
			Name:    NewOCFLString(name),
		},
		err: nil,
	}
	return user
}

func (u *OCFLUser) UnmarshalJSON(data []byte) error {
	tu := &_User{}
	if err := json.Unmarshal(data, tu); err != nil {
		u.err = errors.Wrapf(err, "cannot unmarshal user '%s'", string(data))
		return nil
	}
	u._User.Address = tu.Address
	u._User.Name = tu.Name

	return nil
}

func (u *OCFLUser) Finalize() {
	if u.Name == nil {
		u.Name = NewOCFLString("")
	}
	if u.Address == nil {
		u.Address = NewOCFLString("")
	}
}

// OCFLManifest is a helper type for OCFL manifest maps.
type OCFLManifest struct {
	Manifest map[string][]string
	err      error
}

func (m *OCFLManifest) UnmarshalJSON(data []byte) error {
	m.Manifest = map[string][]string{}
	if err := json.Unmarshal(data, &m.Manifest); err != nil {
		m.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}

	return nil
}

func (m *OCFLManifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Manifest)
}
