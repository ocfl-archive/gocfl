package storageroot

import (
	"context"
)

type Factory interface {
	NewStorageRoot(ctx context.Context) StorageRoot
}
