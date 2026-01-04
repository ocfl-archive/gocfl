package factory

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
)

type Factory interface {
	inventory.Factory
	object.Factory
}
