package inventory

import (
	"context"
)

// Factory is the interface for creating inventory components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	NewVersions(ctx context.Context) Versions
	NewVersion(ctx context.Context) Version
	NewState(ctx context.Context) State
	NewUser(ctx context.Context) User
	NewManifest(ctx context.Context) Manifest
	NewFixity(ctx context.Context) Fixity
	NewInventory(ctx context.Context) Inventory
}
