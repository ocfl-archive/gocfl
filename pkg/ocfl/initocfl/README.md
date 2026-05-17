# OCFL Initialization Module (`pkg/ocfl/initocfl`)

The `initocfl` module provides high-level functions for loading and initializing OCFL structures (Storage Roots and Objects). It abstracts the complexity of version-specific instantiation and the configuration of Extension Managers.

## Goals
The primary goals of this module are:
- Providing a simple API for creating new OCFL structures.
- Automatic detection of OCFL versions when loading existing structures.
- Correct setup of Extension Managers for Storage Roots and Objects.
- Orchestration of various factory modules (`factory`, `extension`, `storageroot`, `object`).

## Core Functions

### Loading Existing Structures
The module provides functions to detect the version of an existing OCFL structure on the filesystem and to instantiate the corresponding object (Storage Root or Object) with the loaded metadata and extensions.

- `LoadStorageRoot(ctx, fsys, extensionParams, logger)`: Loads a Storage Root from an `fs.FS`.
- `LoadObject(ctx, fsys, extensionParams, logger)`: Loads an OCFL Object from an `fs.FS`.

### Initializing New Structures
For creating new OCFL structures, the module provides functions that prepare the corresponding object for a specific OCFL version and create the initial directory structure as well as files (e.g., Namaste files).

- `InitStorageRoot(ctx, fsys, extensionConfigFS, ver, digest, params, logger)`: Initializes a new Storage Root.
- `InitObject(ctx, fsys, extensionConfigFS, ver, id, digest, extensionParams, logger)`: Initializes a new OCFL Object.

### Helper Functions
- `SetupExtensionManager[T](params, fsys, logger)`: A generic function for setting up an Extension Manager (`T` is either `storageroot.ExtensionManager` or `object.ExtensionManager`).
- `NewFactoryObject(ver, extensionFactory, logger)`: Creates an Object Factory suitable for the OCFL version.
- `NewFactoryStorageRoot(ver, extensionFactory, logger)`: Creates a StorageRoot Factory suitable for the OCFL version.

## Usage

### Loading a Storage Root

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
)

// fsys is an fs.FS containing the OCFL Storage Root
sr, srCloser, err := initocfl.LoadStorageRoot(ctx, fsys, nil, logger)
if err != nil {
    // error handling
}
defer srCloser.Close()
```

### Initializing a New Object

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
    "github.com/je4/utils/v2/pkg/checksum"
)

// fsys must be an appendfs.FS to allow write access
obj, err := initocfl.InitObject(ctx, fsys, nil, version.Version1_1, "my-object-id", checksum.DigestSHA512, nil, logger)
if err != nil {
    // error handling
}
```

## Dependencies
This module integrates functionalities from:
- `pkg/ocfl/factory`: For uniform creation of components.
- `pkg/ocfl/extension`: For managing OCFL extensions.
- `pkg/ocfl/storageroot`: For Storage Root logic.
- `pkg/ocfl/object`: For Object logic.
- `pkg/ocfl/version`: For defining OCFL specification versions.
