# Functional Modules

In the `pkg/ocfl/object` package, core OCFL operations are encapsulated into specialized "Functional Modules." These modules represent different stages of an OCFL object's lifecycle and are accessed via the main `Object` interface.

Unlike the high-level orchestration functions in `pkg/ocfl/functions`, these modules represent the internal logic and orchestration of the `Object` implementation itself.

## Available Modules

The following functional modules are available:

- [**Loader**](LOADER.md): Responsible for reading and parsing existing objects from the filesystem.
- [**Initializer**](INITIALIZER.md): Handles the creation of brand-new OCFL objects.
- [**VersionWriter**](VERSION_WRITER.md): Manages the addition of new versions to an object.
- [**Checker**](CHECKER.md): Performs validation and integrity checks on an object.
- [**Extractor**](EXTRACTOR.md): Used to retrieve content from specific versions of an object.

## Integration with Extension Manager

The `ExtensionManager` (`pkg/ocfl/object/extensionManager.go`) coordinates how [OCFL Extensions](../../extension/README.md) interact with these functional modules.

Extensions can hook into various points of the object's lifecycle (e.g., during `Load`, `Check`, or `VersionWriter` operations) through specialized interfaces defined in `pkg/ocfl/object/extension.go`.

---
- [Back to Object Documentation Overview](README.md)
- [The Object Interface](OBJECT.md)
- [OCFL Actions](../actions.go)
