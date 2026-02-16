package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_VersionJSONMarshal20(t *testing.T) {
	genericVersionJSONMarshal(version.Version2_0, t)
}

func Test_VersionJSONUnmarshal20(t *testing.T) {
	genericVersionJSONUnmarshal(version.Version2_0, t)
}

func Test_VersionJSONUnmarshalError20(t *testing.T) {
	genericVersionJSONUnmarshalError(version.Version2_0, t)
}

func Test_VersionState20(t *testing.T) {
	genericVersionState(version.Version2_0, t)
}

func Test_VersionMessage20(t *testing.T) {
	genericVersionMessage(version.Version2_0, t)
}

func Test_VersionCreated20(t *testing.T) {
	genericVersionCreated(version.Version2_0, t)
}
