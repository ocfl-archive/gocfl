package factory

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type Factory interface {
	inventory.Factory
	object.Factory
	storageroot.Factory
	GetVersion() version.OCFLVersion
	SetVersion(version.OCFLVersion) error
	Copy() Factory
}
