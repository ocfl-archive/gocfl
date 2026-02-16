package inventorytest

import (
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func Test_VersionJSONMarshal11(t *testing.T) {
	genericVersionJSONMarshal(version.Version1_1, t)
}

func Test_VersionJSONUnmarshal11(t *testing.T) {
	genericVersionJSONUnmarshal(version.Version1_1, t)
}

func Test_VersionJSONUnmarshalError11(t *testing.T) {
	genericVersionJSONUnmarshalError(version.Version1_1, t)
}

func Test_VersionState11(t *testing.T) {
	genericVersionState(version.Version1_1, t)
}

func Test_VersionMessage11(t *testing.T) {
	genericVersionMessage(version.Version1_1, t)
}

func Test_VersionCreated11(t *testing.T) {
	genericVersionCreated(version.Version1_1, t)
}
