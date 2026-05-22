# VersionWriter Module

The `VersionWriter` module handles the logic of adding a new version to an existing OCFL object. This includes managing logical file paths, deduplication via digests, and finalizing the update by writing the updated [Inventory](../../inventory/README.md).

- **Interface**: `VersionWriter` (`pkg/ocfl/object/object.go`)
- **Specification**: [OCFL 1.1: 3.3.1 Version Directories](../../version/ocfl_spec_1.1.md#331-version-directories)

## Main Methods

The `VersionWriter` provides comprehensive methods for object updates:

### Adding Content
- `AddFile(sourceFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error`: Adds a file from a source filesystem.
- `AddFolder(sourceFS fs.FS, checkDuplicate bool, area string) error`: Recursively adds all files from a source folder.
- `AddData(data []byte, path string, ...)`: Directly adds binary data as a file.
- `AddReader(r io.ReadCloser, files []string, ...)`: Streams data into one or more target logical paths.

### Logical Operations
- `DeleteFile(virtualFilename string, digest string) error`: Removes a logical file reference from the new version.
- `RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error`: Renames a logical path within the version.

### Lifecycle & Areas
- `BeginArea(area string)`: Starts adding content to a specific OCFL content area.
- `EndArea() error`: Finishes the current content area.
- `Close() error`: Finalizes the version update and writes the updated `inventory.json`. See the [Creation Sequence Diagram](create_sequence.md) or [Update Sequence Diagram](update_sequence.md) for a visual walkthrough of the process.

## Usage Example

```go
// The object must be configured with a writable filesystem
obj.WithWriteFS(objFS)

// Start an update to create a new version
vw, err := obj.StartUpdate("New version message", "User Name", "user@example.com", false)
if err != nil {
    // handle error
}
defer vw.Close()
```

---
- [Back to Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [Creation Sequence Diagram](create_sequence.md)
- [Update Sequence Diagram](update_sequence.md)
- [Functional Modules Index](MODULES.md)
