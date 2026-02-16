package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_UserJSONMarshal20(t *testing.T) {
	genericUserJSONMarshal(version.Version2_0, t)
}

func Test_UserJSONUnmarshal20(t *testing.T) {
	genericUserJSONUnmarshal(version.Version2_0, t)
}

func Test_UserInvalidJSON20(t *testing.T) {
	genericUserInvalidJSON(version.Version2_0, t)
}

func Test_UserInvalidAddressCheck20(t *testing.T) {
	genericUserInvalidAddressCheck(version.Version2_0, t)
}
