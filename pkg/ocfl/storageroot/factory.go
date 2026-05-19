package storageroot

import (
	"context"
)

// ConfigName represents the name of a configuration parameter for the storage root factory.
type ConfigName string

const (
	// StorageRootName is the configuration name for the storage root instance.
	StorageRootName ConfigName = "storageroot"
	// InitializerName is the configuration name for the initializer instance.
	InitializerName ConfigName = "initializer"
	// LoaderName is the configuration name for the loader instance.
	LoaderName ConfigName = "loader"
)

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
