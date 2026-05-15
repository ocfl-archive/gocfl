package inventory

import (
	"context"
)

// Factory is the interface for creating inventory components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	// NewVersions creates a new Versions collection.
	NewVersions(ctx context.Context) Versions
	// NewVersion creates a new Version object.
	NewVersion(ctx context.Context) Version
	// NewState creates a new State object.
	NewState(ctx context.Context) State
	// NewUser creates a new User object.
	NewUser(ctx context.Context) User
	// NewManifest creates a new Manifest object.
	NewManifest(ctx context.Context) Manifest
	// NewFixity creates a new Fixity object.
	NewFixity(ctx context.Context) Fixity
	// NewInventory creates a new Inventory object.
	NewInventory(ctx context.Context) Inventory
}
