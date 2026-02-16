package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_VersionsJSONMarshal11(t *testing.T) {
	genericVersionsJSONMarshal(version.Version1_1, t)
}

func Test_VersionsJSONUnmarshal11(t *testing.T) {
	genericVersionsJSONUnmarshal(version.Version1_1, t)
}

func Test_VersionsInvalidJSON11(t *testing.T) {
	genericVersionsInvalidJSON(version.Version1_1, t)
}

func Test_VersionsCheck11(t *testing.T) {
	genericVersionsCheck(version.Version1_1, t)
}
