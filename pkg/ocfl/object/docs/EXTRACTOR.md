# Extractor Module

The `Extractor` module retrieves content from specific versions of an OCFL object. It maps logical paths to their physical locations within the object structure.

- **Interface**: `Extractor` (`pkg/ocfl/object/object.go`)

## Key Methods

- `Extract(version *inventory.VersionNumber, withManifest bool, area string) error`: Extracts all files from a specific [Version](../../inventory/docs/VERSION.md) to a target filesystem.
- `GetFileReader(name string) (io.ReadCloser, int64, string, error)`: Retrieves an `io.ReadCloser` for a logical file path, along with its size and digest.
- `GetExtensionFileReader(extName, path string) (io.ReadCloser, int64, string, error)`: Accesses files stored within object extensions.
- `WithObject(o Object) Extractor`: Associates the extractor with an [Object](OBJECT.md) instance.
- `WithDestFS(destFS appendfs.FS) Extractor`: Sets the destination filesystem for extraction.

> [!NOTE]
> The source filesystem is associated with the `Object` via `WithReadFS(fsys)` before retrieving the extractor.

## Usage Example

The extractor is accessed via the [Object](OBJECT.md) interface:

```go
// Configure source and destination filesystems
obj.WithReadFS(sourceFS)
extractor := obj.GetExtractor().WithDestFS(destFS)
defer extractor.Close()

// Extract the head version (nil)
if err := extractor.Extract(nil, true, ""); err != nil {
    // handle extraction error
}
```

Higher-level functions in the `ocfl` package provide standardized wrappers for common extraction tasks.

---
- [Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Functional Modules Index](MODULES.md)
