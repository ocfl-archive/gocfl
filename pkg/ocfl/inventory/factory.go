package inventory

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
	NewInventory(ctx context.Context, objectFolder string, contentDir string) Inventory
}
