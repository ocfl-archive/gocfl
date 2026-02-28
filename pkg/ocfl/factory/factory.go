package factory

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

// Factory is the Unified Factory interface that combines multiple sub-factories.
// It is responsible for creating OCFL components across different packages.
// For more details, see the [Unified Factory Documentation].
//
// [Unified Factory Documentation]: https://github.com/ocfl-archive/gocfl/blob/main/pkg/ocfl/factory/README.md
type Factory interface {
	inventory.Factory
	object.Factory
	storageroot.Factory
	GetVersion() version.OCFLVersion
	SetVersion(version.OCFLVersion) error
	Copy() Factory
}
