# ocfllogger

`ocfllogger` provides a logger implementation for OCFL operations that integrates with the OCFL validation system.

## Features

- Wrapper around `zLogger.ZLogger` (zerolog).
- Maintains context-specific data (e.g. object ID, version).
- Automatically tracks OCFL validation errors and warnings.
- Supports version-aware validation error reporting.

## Usage

```go
package main

import (
	"context"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
)

func main() {
	// Initialize a base logger (e.g. zerolog)
    var baseLogger zLogger.ZLogger = ... 

	logger := ocfl.NewOCFLLogger(context.Background(), baseLogger, nil, version.V1_1, nil)

	// Log a validation error
	// This will log an error via the base logger and add the error to the internal validation status
	logger.ValidationError(validation.E001, "Invalid inventory file")

    // Get all accumulated validation errors
    errors := logger.ValidationErrors()
}
```
