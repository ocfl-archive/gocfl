# Object Extension Hooks

This document describes the extension hooks available for OCFL objects. These hooks allow extensions to intervene in the object lifecycle, perform path transformations, or add custom metadata.

The `ExtensionManager` uses type assertions at runtime to identify which hooks an extension implements and calls the corresponding methods.

## Path Transformations

Hooks that influence how paths are generated within the OCFL object (e.g., manifest, state, or extraction).

### `ExtensionObjectContentPath`
Transforms paths for object content.
- **Method**: `BuildObjectManifestPath(originalPath string, area string) (string, error)`
- **Usage**: Called when generating a physical path (under `content/`) for a file.

### `ExtensionObjectStatePath`
Transforms paths for the logical state.
- **Method**: `BuildObjectStatePath(originalPath string, area string) (string, error)`
- **Usage**: Influences the logical path in the `state` section of the inventory.

### `ExtensionObjectExtractPath`
Determines paths during extraction.
- **Method**: `BuildObjectExtractPath(originalPath string, area string) (string, error)`
- **Usage**: Used when extracting files to a destination filesystem.

## Content and Object Operations

Hooks called during write operations or general object updates.

### `ExtensionContentChange`
Enables actions before and after file operations.
- **Methods**:
    - `AddFileBefore(...)`, `AddFileAfter(...)`
    - `UpdateFileBefore(...)`, `UpdateFileAfter(...)`
    - `DeleteFileBefore(...)`, `DeleteFileAfter(...)`
- **Usage**: Ideal for logging, validation, or automatic generation of sidecar files.

### `ExtensionObjectChange`
Called during general object updates.
- **Methods**: `UpdateObjectBefore(...)`, `UpdateObjectAfter(...)`

## Metadata and Fixity

### `ExtensionMetadata`
Adds custom metadata.
- **Method**: `GetMetadata(sourceFS fs.FS, obj Object) (map[string]any, error)`
- **Usage**: Integrated into the object's metadata output.

### `ExtensionFixityDigest`
Registers additional checksum algorithms.
- **Method**: `GetFixityDigests() []checksum.DigestAlgorithm`
- **Usage**: Protects files with additional algorithms beyond the OCFL primary digest.

## Lifecycle and Workflow

### `ExtensionArea`
Manages paths for specific work areas.
- **Method**: `GetAreaPath(area string) (string, error)`

### `ExtensionStream`
Enables streaming content into or out of the object.
- **Method**: `StreamObject(writer VersionWriter, reader io.Reader, stateFiles []string, dest string) error`
- **Note**: Streams are processed in parallel using `io.MultiWriter`.

### `ExtensionNewVersion`
Controls automatic creation of subsequent versions.
- **Methods**:
    - `NeedNewVersion(writer VersionWriter) (bool, error)`: Checks if an automatic subsequent version is required (e.g., for auto-generated metadata).
    - `DoNewVersion(writer VersionWriter) error`: Performs changes in the automatically created version before it is closed.

### `ExtensionVersionDone`
Called when a version is finalized.
- **Method**: `VersionDone(obj Object) error`
- **Note**: Intended for final actions (e.g., notifications) once the object is stable.

---
- [Extension Overview](../../extension/docs/README.md)
- [OCFL Object Documentation](OBJECT.md)
