package version

import (
	"regexp"
	"slices"
)

type OCFLVersion string

func (v OCFLVersion) String() string {
	return string(v)
}

const Version1_1 OCFLVersion = "1.1"
const Version1_0 OCFLVersion = "1.0"
const Version2_0 OCFLVersion = "2.0"

var ValidVersions = []OCFLVersion{
	Version1_1,
	Version1_0,
	Version2_0,
}

const Default = Version1_1

var OCFLStorageRootVersionNamasteRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")
var ObjectVersionRegexp = regexp.MustCompile("^0=ocfl_object_([0-9]+\\.[0-9]+)$")

func ValidVersion(ver OCFLVersion) bool {
	return slices.Contains(ValidVersions, ver)
}
