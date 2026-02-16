package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_VersionsJSONMarshal20(t *testing.T) {
	genericVersionsJSONMarshal(version.Version2_0, t)
}

func Test_VersionsJSONUnmarshal20(t *testing.T) {
	genericVersionsJSONUnmarshal(version.Version2_0, t)
}

func Test_VersionsInvalidJSON20(t *testing.T) {
	genericVersionsInvalidJSON(version.Version2_0, t)
}

func Test_VersionsCheck20(t *testing.T) {
	genericVersionsCheck(version.Version2_0, t)
}
