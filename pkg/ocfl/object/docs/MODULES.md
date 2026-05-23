# Functional Modules

In the `pkg/ocfl/object` package, OCFL operations are encapsulated into specialized "Functional Modules." These modules manage different stages of an OCFL object's lifecycle and are accessed via the `Object` interface.

## Available Modules

- [**Loader**](LOADER.md): Reads and parses existing objects.
- [**Initializer**](INITIALIZER.md): Creates new OCFL objects.
- [**VersionWriter**](VERSION_WRITER.md): Adds new versions to an object.
- [**Validator**](VALIDATOR.md): Performs validation and integrity checks.
- [**Extractor**](EXTRACTOR.md): Retrieves content from specific versions.

## Extension Manager Integration

The `ExtensionManager` coordinates how [OCFL Extensions](../../extension/README.md) interact with these modules. Extensions can hook into the object lifecycle (e.g., during `Load`, `Validate`, or `VersionWriter` operations) through specialized interfaces. See [Object Extension Hooks](HOOKS.md) for details.

---
- [Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
