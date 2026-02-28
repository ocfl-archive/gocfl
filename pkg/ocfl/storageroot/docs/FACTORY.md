# Factory Interface

The `Factory` interface (`pkg/ocfl/storageroot/factory.go`) is responsible for instantiating `StorageRoot` objects and their operational modules. This pattern allows for the creation of version-specific implementations while keeping the rest of the library decoupled.

> [!IMPORTANT]
> The `storageroot.Factory` interface should **not** be called directly by client code. Instead, use the **Unified Factory** provided by the `factory` package, which implements this interface along with others.

## Factory Interface

The factory provides methods for creating core storage root components:

- `NewStorageRoot(ctx context.Context) StorageRoot`: Creates a new [StorageRoot](STORAGEROOT.md) instance.
- `NewStorageRootInitializer(ctx context.Context) Initializer`: Creates an [Initializer](INITIALIZER.md) for a new storage root.
- `NewStorageRootLoader(ctx context.Context) Loader`: Creates a [Loader](LOADER.md) for an existing storage root.

## Usage

The storage root factory is typically used via the **Unified Factory** to ensure that the correct version-specific implementation is selected.

### Example via Unified Factory

```go
import (
    "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
    "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
)

// Assume 'f' is an instance of factory.Factory (e.g. from factoryimpl.NewFactoryBase)
func HandleStorageRoot(ctx context.Context, f factory.Factory) error {
    // Create a new storage root instance via the factory
    sr := f.NewStorageRoot(ctx)
    
    // Use the storage root...
    return nil
}
```

The factory implementation is responsible for determining the appropriate OCFL version and configuration when creating these instances.

---
- [Back to Storage Root Overview](../README.md)
- [OCFL Factory](../../factory/README.md)
