# Object Documentation

This directory contains detailed technical documentation for the OCFL Object components of the `gocfl` library.

## Overview

An OCFL Object is the basic unit of storage in an OCFL repository. It contains all versions of a set of files, along with an inventory that describes the object's state and history.

## Components

### [Object](OBJECT.md)
The central interface (`Object`) that represents an OCFL Object. It provides access to object metadata and serves as a factory for operational modules.

### [Loader](LOADER.md)
The `Loader` module is responsible for loading an existing OCFL object from a filesystem and parsing its inventory.

### [Initializer](INITIALIZER.md)
The `Initializer` module handles the creation of a new OCFL object, including the initial inventory and object structure.

### [VersionWriter](VERSION_WRITER.md)
The `VersionWriter` is used to add new versions to an existing object. It manages the process of staging files, updating the inventory, and committing the changes.

### [Extractor](EXTRACTOR.md)
The `Extractor` module allows retrieving specific files or entire versions from the object and writing them to a destination filesystem.

### [Checker](CHECKER.md)
The `Checker` module provides functionality to validate the integrity and OCFL compliance of an object.

### [Modules and Extensions](MODULES.md)
Detailed information about the internal modules and how the object interacts with the Extension Manager.

## Related Documentation

- [Object Package README](../README.md) - High-level overview and usage examples.
- [OCFL Specification 1.1](../../version/ocfl_spec_1.1.md)
- [Storage Root Documentation](../../storageroot/docs/README.md)
- [Inventory Documentation](../../inventory/README.md)
- [Extension Documentation](../../extension/docs/README.md)
