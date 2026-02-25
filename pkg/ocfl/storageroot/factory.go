package storageroot

import (
	"context"
)

type Factory interface {
	NewStorageRoot(ctx context.Context) StorageRoot
	NewStorageRootInitializer(ctx context.Context) Initializer
	NewStorageRootLoader(ctx context.Context) Loader
}
