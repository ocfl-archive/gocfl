package storageroot

import (
	"context"
)

type ConfigName string

const StorageRootName ConfigName = "storageroot"
const InitializerName ConfigName = "initializer"
const LoaderName ConfigName = "loader"

// Factory is the interface for creating storage root components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	// NewStorageRoot creates a new StorageRoot instance.
	NewStorageRoot(ctx context.Context) StorageRoot
	// NewStorageRootInitializer creates a new Initializer instance.
	NewStorageRootInitializer(ctx context.Context) Initializer
	// NewStorageRootLoader creates a new Loader instance.
	NewStorageRootLoader(ctx context.Context) Loader
}
