# OCFL Actions

The `ocflactions` package provides high-level orchestration functions for common OCFL tasks. It serves as a primary API for many use cases, simplifying operations that involve multiple internal modules.

## Core Functions

### [CheckObject](actions.go)
Validates an OCFL object by loading it and running its checker.
- Loads the object using `initocfl.LoadObject`.
- Executes the object's checker to verify its integrity.

### [Extract](actions.go)
Extracts files from a specific version of an OCFL object to a destination filesystem.
- Supports extracting with or without a manifest.
- Allows specifying a specific OCFL version or defaults to the latest.
- Can target specific content areas within the object.

### [ExtractMeta](actions.go)
Retrieves the inventory metadata of an OCFL object.
- Provides a high-level way to access object metadata without manual inventory parsing.

## Usage Example

```go
import (
    "context"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflactions"
    "github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
    "io/fs"
)

// ... setup objectFS (fs.FS) and logger ...

ctx := context.Background()
err := ocflactions.CheckObject(ctx, objectFS, logger)
if err != nil {
    // handle error
}
```

---
- [Back to OCFL Core Packages](../README.md)
