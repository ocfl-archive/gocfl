// Package version defines the supported OCFL versions and provides utilities
// for version comparison and specification access.
package version

import (
	_ "embed"
	"regexp"
	"slices"
)

// OCFLVersion represents an OCFL specification version string.
type OCFLVersion string

// String returns the string representation of the OCFL version.
func (v OCFLVersion) String() string {
	return string(v)
}

const (
	// Version1_1 represents OCFL version 1.1.
	Version1_1 OCFLVersion = "1.1"
	// Version1_0 represents OCFL version 1.0.
	Version1_0 OCFLVersion = "1.0"
	// Version2_0 represents OCFL version 2.0 (future version).
	Version2_0 OCFLVersion = "2.0"
)

// ValidVersions contains all supported OCFL versions.
var ValidVersions = []OCFLVersion{
	Version1_1,
	Version1_0,
	Version2_0,
}

// FloatVersions maps OCFL versions to their float64 representation for comparison.
var FloatVersions = map[OCFLVersion]float64{
	Version1_1: 1.1,
	Version1_0: 1.0,
	Version2_0: 2.0,
}

// Default is the default OCFL version used when no version is specified.
const Default = Version1_1

// OCFLSpec1_0 contains the full text of the OCFL 1.0 specification.
//
//go:embed ocfl_spec_1.0.md
var OCFLSpec1_0 string

// OCFLSpec1_1 contains the full text of the OCFL 1.1 specification.
//
//go:embed ocfl_spec_1.1.md
var OCFLSpec1_1 string

// Spec maps OCFL versions to their respective specification documents.
var Spec = map[OCFLVersion]string{
	Version1_0: OCFLSpec1_0,
	Version1_1: OCFLSpec1_1,
}

// OCFLStorageRootVersionNamasteRegexp is the regular expression for the Namaste file in the storage root.
var OCFLStorageRootVersionNamasteRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

// ObjectVersionRegexp is the regular expression for the Namaste file in an OCFL object.
var ObjectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

// ValidVersion checks if the given version is supported.
func ValidVersion(ver OCFLVersion) bool {
	return slices.Contains(ValidVersions, ver)
}

// Less compares two OCFL versions and returns true if ver is less than ver2.
func Less(ver OCFLVersion, ver2 OCFLVersion) bool {
	if !ValidVersion(ver) {
		return false
	}
	if !ValidVersion(ver2) {
		return false
	}
	return FloatVersions[ver] < FloatVersions[ver2]
}
