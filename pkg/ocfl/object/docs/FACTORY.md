# Object Factory Interface

The `Factory` interface (`pkg/ocfl/object/factory.go`) is responsible for instantiating `Object` instances and their operational modules. This pattern allows for the creation of version-specific implementations while keeping the rest of the library decoupled.

> [!IMPORTANT]
> The `object.Factory` interface should **not** be called directly by client code. Instead, use the **Unified Factory** provided by the `factory` package, which implements this interface along with others.

## Factory Interface

The factory provides methods for creating core object components:

- `NewObject(ctx context.Context) Object`: Creates a new [Object](OBJECT.md) instance.
- `NewLoader(ctx context.Context) Loader`: Creates a [Loader](LOADER.md) for loading an existing object.
- `NewInitializer(ctx context.Context) Initializer`: Creates an [Initializer](INITIALIZER.md) for creating a new object.
- `NewValidator(ctx context.Context) Validator`: Creates a [Validator](VALIDATOR.md) for validating an object.
- `NewExtractor(ctx context.Context) Extractor`: Creates an [Extractor](EXTRACTOR.md) for extracting content.
- `NewVersionWriter(ctx context.Context) VersionWriter`: Creates a [VersionWriter](VERSION_WRITER.md) for adding new versions.

## Usage

The object factory is typically used via the **Unified Factory** to ensure that the correct version-specific implementation is selected.

> [!TIP]
> While factories provide low-level control, the **initocfl** package provides high-level functions like `LoadObject` and `InitObject` that handle factory instantiation and common setup for you.

---
- [Back to Object Overview](README.md)
- [OCFL Factory](../../factory/README.md)
