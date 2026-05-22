# Storage Root Documentation

This directory contains detailed technical documentation for the OCFL Storage Root components of the `gocfl` library.

## Overview

The Storage Root is the top-level directory of an OCFL compliant repository. It contains OCFL objects and repository-wide configuration and extensions.

## Components

### [Architecture](architecture.md)
Visual representation of the interface hierarchy and their relationships.

### [StorageRoot](STORAGEROOT.md)
The central interface (`StorageRoot`) that represents an OCFL Storage Root. It coordinates filesystem access, extension management (especially storage layouts), and object discovery.

### [Loader](LOADER.md)
The `Loader` module is responsible for identifying and loading an existing OCFL Storage Root from a filesystem. It parses the conformance declaration and initializes the extension manager.

### [Initializer](INITIALIZER.md)
The `Initializer` module handles the creation of a new OCFL Storage Root. This includes writing the conformance declaration (Namaste file), setting up the `extensions` directory, and persisting the storage layout configuration.

### [Factory](FACTORY.md)
The `Factory` interface provides methods to instantiate the various storage root components. It allows the library to support different OCFL versions by providing version-specific implementations of the functional modules.

## Related Documentation

- [Storage Root Package README](../README.md) - High-level overview and usage examples.
- [OCFL Specification 1.1](../../version/ocfl_spec_1.1.md)
- [Architecture Overview](architecture.md)
- [Load Sequence Diagram](load_sequence.md)
- [Initialization Sequence Diagram](init_sequence.md)
- [Object Documentation](../../object/docs/README.md)
- [Extension Documentation](../../extension/docs/README.md)
