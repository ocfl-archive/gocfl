package test

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"testing"
)

func Test_UserJSONMarshal11(t *testing.T) {
	genericUserJSONMarshal(version.Version1_1, t)
}

func Test_UserJSONUnmarshal11(t *testing.T) {
	genericUserJSONUnmarshal(version.Version1_1, t)
}

func Test_UserInvalidJSON11(t *testing.T) {
	genericUserInvalidJSON(version.Version1_1, t)
}

func Test_UserInvalidAddressCheck11(t *testing.T) {
	genericUserInvalidAddressCheck(version.Version1_1, t)
}
