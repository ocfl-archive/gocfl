# Initial Extension

This package implements the OCFL `initial` extension.

## Overview

The `initial` extension allows indication that the semantics of a particular extension takes precedence over all other extensions. It ensures that the special extension name `initial` is a registered extension name.

For more details on the specification, see [initial.md](initial.md).

## Usage

This extension is typically used to define which extension manager or functional extension should be applied first.

```go
import "github.com/ocfl-archive/gocfl/v3/pkg/extensions/ext_initial"
```
