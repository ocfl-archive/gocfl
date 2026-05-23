# Object Documentation

This directory provides technical documentation for the OCFL Object components of the `gocfl` library.

## Core Components

### [Architecture](architecture.md)
Interface hierarchy and relationships.

### [Object](OBJECT.md)
The central `Object` interface. It provides metadata access and acts as a factory for operational modules.

### [Loader](LOADER.md)
Responsible for loading existing objects and parsing inventories. See also: [Load Sequence](load_sequence.md).

### [Initializer](INITIALIZER.md)
Handles creating new objects and base structures. See also: [Creation Sequence](create_sequence.md).

### [VersionWriter](VERSION_WRITER.md)
Manages adding versions, staging files, and updating inventories. See also: [Update Sequence](update_sequence.md).

### [Extractor](EXTRACTOR.md)
Retrieves files or versions and writes them to a destination filesystem.

### [Validator](VALIDATOR.md)
Validates integrity and OCFL compliance.

### [Modules and Extensions](MODULES.md)
Internal module details and interaction with the Extension Manager.

## See Also

- [Object Package README](../README.md) - Overview and examples.
- [OCFL Specification](https://ocfl.io/)
- [Storage Root Docs](../../storageroot/docs/README.md)
- [Inventory Docs](../../inventory/README.md)
- [Extension Docs](../../extension/docs/README.md)
