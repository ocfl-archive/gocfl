package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_StateJSONMarshal20(t *testing.T) {
	genericStateJSONMarshal(version.Version2_0, t)
}

func Test_StateJSONUnmarshal20(t *testing.T) {
	genericStateJSONUnmarshal(version.Version2_0, t)
}

func Test_StateJSONUnmarshalError20(t *testing.T) {
	genericStateJSONUnmarshalError(version.Version2_0, t)
}

func Test_StateDelete20(t *testing.T) {
	genericStateDelete(version.Version2_0, t)
}

func Test_StateRename20(t *testing.T) {
	genericStateRename(version.Version2_0, t)
}

func Test_StateUniqueness20(t *testing.T) {
	genericStateUniqueness(version.Version2_0, t)
}
