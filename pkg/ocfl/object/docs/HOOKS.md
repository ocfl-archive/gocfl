# Object Extension Hooks

This document describes the available extension hooks for OCFL objects in the `pkg/ocfl/object` package. These hooks allow extensions to intervene in various phases of the object lifecycle or perform path transformations.

The `ExtensionManager` checks at runtime via type assertion whether a loaded extension implements one of these interfaces and calls the corresponding methods.

## Path Transformations

These hooks influence how paths are formed within the OCFL object (e.g., in the manifest or state).

### `ExtensionObjectContentPath`
Transforms paths for the object's content.
- **Method**: `BuildObjectManifestPath(originalPath string, area string) (string, error)`
- **Usage**: Called when a physical path in the OCFL object (inside `content/`) needs to be generated for a file.

### `ExtensionObjectStatePath`
Transforms paths for the state (logical view) of the object.
- **Method**: `BuildObjectStatePath(originalPath string, area string) (string, error)`
- **Usage**: Influences the path under which a file appears in the `state` section of the inventory.

### `ExtensionObjectExtractPath`
Determines the path when extracting files.
- **Method**: `BuildObjectExtractPath(originalPath string, area string) (string, error)`
- **Usage**: Used when files are extracted from the object into a local file system.

## Content and Object Changes

These hooks are called when operations are performed on the object.

### `ExtensionContentChange`
Enables actions before and after file operations.
- **Methods**:
    - `AddFileBefore(object VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error`
    - `AddFileAfter(versionWriter VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error`
    - `UpdateFileBefore(object VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error`
    - `UpdateFileAfter(object VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error`
    - `DeleteFileBefore(versionWriter VersionWriter, dest string, area string) error`
    - `DeleteFileAfter(object VersionWriter, dest string, area string) error`
- **Usage**: Ideal for logging, validation, or automatic generation of sidecar files.

### `ExtensionObjectChange`
Called for general changes to the object.
- **Methods**:
    - `UpdateObjectBefore(object VersionWriter) error`
    - `UpdateObjectAfter(object VersionWriter) error`

## Metadata and Fixity

### `ExtensionMetadata`
Enables adding custom metadata.
- **Method**: `GetMetadata(sourceFS fs.FS, obj Object) (map[string]any, error)`
- **Usage**: The returned metadata is integrated into the object's metadata output.

### `ExtensionFixityDigest`
Registers additional checksum algorithms.
- **Method**: `GetFixityDigests() []checksum.DigestAlgorithm`
- **Usage**: Allows protecting files with additional algorithms beyond the OCFL standard (SHA-512/SHA-256).

## Lifecycle and Workflow

### `ExtensionArea`
Manages paths for specific work areas (Areas).
- **Method**: `GetAreaPath(area string) (string, error)`

### `ExtensionStream`
Enables streaming of content into or out of the object.
- **Method**: `StreamObject(object VersionWriter, reader io.Reader, stateFiles []string, dest string) error`
- **Note**: Since the data stream is split using `io.MultiWriter`, all registered stream hooks are served in parallel.

### `ExtensionNewVersion`
Controls the automatic creation of subsequent versions.
- **Methods**:
    - `NeedNewVersion(object VersionWriter) (bool, error)`: Called at the end of `Close()` of a version. Returns `true` if a new subsequent version should be created automatically (e.g., for automatically generated metadata or sidecar files).
    - `DoNewVersion(object VersionWriter) error`: Called after an automatic subsequent version has been created. Allows the extension to make changes to this new version before it is automatically closed.

### `ExtensionVersionDone`
Called when a version has been fully completed.
- **Method**: `VersionDone(object Object) error`
- **Note**: This hook is currently being prepared and is intended for final actions (e.g., notifications or archiving) after the object is back in a stable state (read-only access).

---
- [Back to Extension Overview](../../extension/docs/EXTENSION.md)
- [OCFL Object Documentation](OBJECT.md)
