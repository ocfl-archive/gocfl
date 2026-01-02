package types

import (
	"context"
)

type Factory interface {
	NewVersions() Versions
	NewVersion() Version
	NewState() State
	NewUser() User
	NewManifest() Manifest
	NewFixity() Fixity
	NewInventory(ctx context.Context, contentDir string) Inventory
}
