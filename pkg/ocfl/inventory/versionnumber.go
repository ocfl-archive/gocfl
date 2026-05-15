package inventory

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

var VersionZeroRegexp = regexp.MustCompile("^v0[0-9]+$")
var VersionNoZeroRegexp = regexp.MustCompile("^v[1-9][0-9]*$")

func NewVersionNumber() *VersionNumber {
	return &VersionNumber{}
}

// VersionNumber represents an OCFL version number (e.g. "v1", "v0002").
type VersionNumber struct {
	string // version string representation
	int    // version integer value
}

// GetPaddingLength returns the length of the zero padding, or -1 if no padding is present.
func (v *VersionNumber) GetPaddingLength() int {
	if VersionZeroRegexp.MatchString(v.string) {
		return len(v.string) - 2
	}
	return -1
}

// Equal returns true if two version numbers are numerically equal.
func (v *VersionNumber) Equal(v2 *VersionNumber) bool {
	return v.int == v2.int
}

// Less returns true if this version number is numerically less than v2.
func (v *VersionNumber) Less(v2 *VersionNumber) bool {
	return v.int < v2.int
}

// WithLatest sets the version number representation to "latest".
func (v *VersionNumber) WithLatest() *VersionNumber {
	v.string = "latest"
	return v
}

func (v *VersionNumber) WithString(versionString string) *VersionNumber {
	v.string = versionString
	v.int, _ = strconv.Atoi(strings.TrimLeft(versionString, "v0"))
	return v
}

func (v *VersionNumber) String() string {
	if v == nil {
		return ""
	}
	return v.string
}
func (v *VersionNumber) Int() int {
	if v == nil {
		return 0
	}
	return v.int
}

// IsLatest returns true if the version number represents the "latest" version.
func (v *VersionNumber) IsLatest() bool {
	if v == nil {
		return false
	}
	return v.string == "latest"
}

// IsValid returns true if the version number is valid (numerically greater than zero).
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

func (v *VersionNumber) Init(versionInt int, versionString string) *VersionNumber {
	v.string = versionString
	v.int = versionInt
	return v
}
