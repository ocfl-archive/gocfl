# GOCFL Extension Manager

The `NNNN-gocfl-extension-manager` extension is a special extension that manages the execution and exclusion of other OCFL extensions. It allows for defining a specific order of execution and excluding redundant or conflicting extensions.

## Documentation

The technical specification and detailed documentation for this extension can be found in [NNNN-gocfl-extension-manager.md](NNNN-gocfl-extension-manager.md).

## Features

- **Sorted Execution**: Define the order in which extensions are called.
- **Extension Exclusion**: Prevent specific extensions from running based on configuration.
- **Generic Support**: Works with both OCFL Storage Root and Object extensions.
- **Initial Extension Support**: Integrates with the `initial` extension to determine the active manager.
