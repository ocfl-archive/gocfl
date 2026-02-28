package storageroot

import (
	"context"
)

// Factory is the interface for creating storage root components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	NewStorageRoot(ctx context.Context) StorageRoot
	NewStorageRootInitializer(ctx context.Context) Initializer
	NewStorageRootLoader(ctx context.Context) Loader
}
