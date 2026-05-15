# Extractor Module

The `Extractor` module is used to retrieve content from specific versions of an OCFL object. It handles the mapping from logical paths to their physical locations in the object structure.

- **Interface**: `Extractor` (`pkg/ocfl/object/object.go`)

## Main Methods

The `Extractor` interface provides methods for retrieving and extracting object content:

- `Extract(version *inventory.VersionNumber, withManifest bool, area string) error`: Extracts all files from a specific [Version](../../inventory/docs/VERSION.md) of the object to a target filesystem.
- `GetFileReader(name string) (io.ReadCloser, int64, string, error)`: Retrieves an `io.ReadCloser` for a specific logical file, along with its size and digest.
- `GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error)`: Provides access to files stored within object extensions.
- `WithObject(o Object) Extractor`: Associates the extractor with an [Object](OBJECT.md) instance.
- `WithDestFS(destFS appendfs.FS) Extractor`: Configures the destination filesystem for extraction.

> [!NOTE]
> The source filesystem is now passed to the `Object` via `WithReadFS(fsys)` before calling `GetExtractor()`.

## Usage Example

The extractor is accessed via the [Object](OBJECT.md) interface:

```go
// Set source and destination filesystems
obj.WithReadFS(sourceFS)
extractor := obj.GetExtractor().WithDestFS(destFS)

// Extract the head version (nil)
if err := extractor.Extract(nil, true, ""); err != nil {
    // handle extraction error
}
```

Higher-level orchestration functions in [pkg/ocfl/ocflactions](../actions.go) like `Extract` wrap this logic for common use cases.

---
- [Back to Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
