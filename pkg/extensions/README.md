# Core OCFL Extensions

This directory contains the core extensions for `gocfl`. These extensions are essential for the operation of the OCFL implementation and are initialized automatically even if no specific configuration is provided for them.

They are implemented as extensions to allow for future flexibility, enabling them to be replaced by alternative implementations if necessary.

## Extensions

### Initial (`initial`)

[README](ext_initial/README.md)

The `initial` extension is responsible for defining which extension manager should be used. It acts as a bootstrap mechanism to set up the extension management system.

- **Name**: `initial`
- **Purpose**: Defines the extension manager name.
- **Behavior**: Provides a default extension manager name if none is configured.

### GOCFL Extension Manager (`NNNN-gocfl-extension-manager`)

[README](ext_NNNN_gocfl_extension_manager/README.md)

The `NNNN-gocfl-extension-manager` is the primary orchestrator for all other extensions in `gocfl`. It implements various OCFL extension hooks and manages the execution order and lifecycle of other extensions.

- **Name**: `NNNN-gocfl-extension-manager`
- **Purpose**: Manages and coordinates all extensions.
- **Features**:
  - Handles extension registration and initialization.
  - Implements hooks for storage root layout, object manifest/state path transformations, and file operation lifecycle (before/after hooks).
  - Manages metadata extraction and fixity digest generation across multiple extensions.
  - Controls the execution order of extensions through sorting and exclusion logic.
