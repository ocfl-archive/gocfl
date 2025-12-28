package inventory

import (
	"encoding/json"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

func NewVersionNumber() *VersionNumber {
	return &VersionNumber{}
}

type VersionNumber struct {
	string
	int
}

func (v *VersionNumber) WithString(s string) *VersionNumber {
	v.string = s
	v.int, _ = strconv.Atoi(strings.TrimLeft(s, "v0"))
	return v
}

func (v *VersionNumber) String() string {
	return v.string
}
func (v *VersionNumber) Int() int {
	return v.int
}
func (v *VersionNumber) IsValid() bool {
	if v == nil {
		return false
	}
	return v.int > 0
}

func (v *VersionNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.string)
}

func (v *VersionNumber) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	v.string = s
	v.int, _ = strconv.Atoi(strings.TrimLeft(s, "v0"))
	return nil
}
