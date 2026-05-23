# VersionWriter Module

The `VersionWriter` module manages the creation of new OCFL versions. It handles logical file paths, deduplication via digests, and finalizes updates by writing the updated [Inventory](../../inventory/README.md).

- **Interface**: `VersionWriter` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1 Version Directories](https://ocfl.io/1.1/spec/#version-directories)

## Key Methods

### Adding Content
- `AddFile(sourceFS, path, checkDuplicate, area, noExt, isDir) error`: Adds a file from a source filesystem.
- `AddFolder(sourceFS, checkDuplicate, area) error`: Recursively adds a folder's content.
- `AddData(data, path, ...)`: Directly adds binary data as a file.
- `AddReader(r, files, ...)`: Streams data into logical paths.

### Logical Operations
- `DeleteFile(name, digest string) error`: Removes a logical file reference.
- `RenameFile(src, dest, digest string) error`: Renames a logical path.

### Lifecycle and Areas
- `BeginArea(area string)`: Starts adding content to a specific area.
- `EndArea() error`: Finalizes the current area.
- `Close() error`: Finalizes the version update and writes the `inventory.json`. See [Update Sequence](update_sequence.md).

## Usage Example

```go
// Object must be configured with a writable filesystem
obj.WithWriteFS(objFS)

// Start a new version update
vw, err := obj.StartUpdate("Commit message", "User", "user@example.com", false)
if err != nil {
    // handle error
}
defer vw.Close()

// Add content
err = vw.AddFile(sourceFS, "data.txt", true, "", false, false)
```

---
- [Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Update Sequence Diagram](update_sequence.md)
- [Functional Modules Index](MODULES.md)
