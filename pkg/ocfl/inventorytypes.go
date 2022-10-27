package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"time"
)

type OCFLState struct {
	State map[string][]string
	err   error
}

func (s *OCFLState) UnmarshalJSON(data []byte) error {
	s.State = map[string][]string{}
	if err := json.Unmarshal(data, &s.State); err != nil {
		s.err = errors.Wrapf(err, "cannot unmarshal state %s", string(data))
		return nil
	}
	return nil
}

type OCFLString struct {
	string
	err error
}

func (s *OCFLString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		s.err = errors.Wrapf(err, "cannot unmarshal string '%s'", string(data))
		return nil
	}
	s.string = string(data)
	return nil
}

type OCFLTime struct {
	time.Time
	err error
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

type User struct {
	Address OCFLString `json:"address,omitempty"`
	Name    OCFLString `json:"name"`
}
type OCFLUser struct {
	User
	err error
}

func (u *OCFLUser) UnmarshalJSON(data []byte) error {
	tu := &User{}
	if err := json.Unmarshal(data, tu); err != nil {
		u.err = errors.Wrapf(err, "cannot unmarshal user '%s'", string(data))
		return nil
	}
	u.User.Address = tu.Address
	u.User.Name = tu.Name

	return nil
}

type Version struct {
	Created OCFLTime   `json:"created"`
	Message OCFLString `json:"message"`
	State   OCFLState  `json:"state"`
	User    OCFLUser   `json:"user"`
}

func (v *Version) Equal(v2 *Version) bool {
	if v2 == nil {
		return false
	}
	if v.Created.Time != v2.Created.Time ||
		v.Message.string != v2.Message.string ||
		v.User.Name != v2.User.Name ||
		v.User.Address != v.User.Address {
		return false
	}
	if len(v.State.State) != len(v2.State.State) {
		return false
	}
	for key, vals := range v.State.State {
		vals2, ok := v2.State.State[key]
		if !ok {
			return false
		}
		if len(vals) != len(vals2) {
			return false
		}
		if !sliceContains(vals, vals2) {
			return false
		}
	}
	return true
}

type OCFLVersions struct {
	Versions map[string]*Version
	err      error
}

func (v *OCFLVersions) UnmarshalJSON(data []byte) error {
	v.Versions = map[string]*Version{}
	if err := json.Unmarshal(data, &v.Versions); err != nil {
		v.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}

	return nil
}
