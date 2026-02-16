package inventorytest

import (
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func Test_StateJSONMarshal11(t *testing.T) {
	genericStateJSONMarshal(version.Version1_1, t)
}

func Test_StateJSONUnmarshal11(t *testing.T) {
	genericStateJSONUnmarshal(version.Version1_1, t)
}

func Test_StateJSONUnmarshalError11(t *testing.T) {
	genericStateJSONUnmarshalError(version.Version1_1, t)
}

func Test_StateDelete11(t *testing.T) {
	genericStateDelete(version.Version1_1, t)
}

func Test_StateRename11(t *testing.T) {
	genericStateRename(version.Version1_1, t)
}

func Test_StateUniqueness11(t *testing.T) {
	genericStateUniqueness(version.Version1_1, t)
}
