package types

import (
	"context"
)

type Factory interface {
	NewVersions(ctx context.Context) Versions
	NewVersion(ctx context.Context) Version
	NewState(ctx context.Context) State
	NewUser(ctx context.Context) User
	NewManifest(ctx context.Context) Manifest
	NewFixity(ctx context.Context) Fixity
	NewInventory(ctx context.Context) Inventory
}
