# OCFL High-Level API

The `ocfl` package provides a high-level entry point for working with OCFL (Oxford Common File Layout) structures. It abstracts the complexity of version-specific instantiation, extension management, and orchestrates common actions.

## Initialization & Loading

These functions are the primary way to start working with OCFL Storage Roots and Objects. They handle version detection and the setup of extension managers automatically.

### Storage Root

#### `LoadStorageRoot`
Loads an existing OCFL Storage Root.
- **Parameters**: `ctx`, `fsys` (read-only), `extensionParams`, `conf`, `logger`.
- **Returns**: `storageroot.StorageRoot`, `error`.
- **Details**: Detects the OCFL version from the filesystem and initializes the corresponding implementation.

```go
sr, err := ocfl.LoadStorageRoot(ctx, fsys, nil, nil, logger)
if err != nil {
    log.Fatal(err)
}
defer sr.Close()
```

#### `InitStorageRoot`
Initializes a new OCFL Storage Root.
- **Parameters**: `ctx`, `fsys` (must be writable), `extensionConfigFS`, `ver`, `digest`, `params`, `logger`.
- **Returns**: `storageroot.StorageRoot`, `error`.

```go
sr, err := ocfl.InitStorageRoot(ctx, writeFS, nil, version.V1_1, checksum.DigestSHA512, nil, logger)
if err != nil {
    log.Fatal(err)
}
defer sr.Close()
```

### Object

#### `LoadObject`
Loads an existing OCFL Object.
- **Parameters**: `ctx`, `fsys` (read-only), `extensionParams`, `logger`.
- **Returns**: `object.Object`, `error`.
- **Details**: Detects the object's OCFL version and loads the inventory.

```go
obj, err := ocfl.LoadObject(ctx, objFS, nil, logger)
if err != nil {
    log.Fatal(err)
}
defer obj.Close()
```

#### `InitObject`
Initializes a new OCFL Object.
- **Parameters**: `ctx`, `fsys` (must be writable), `extensionConfigFS`, `ver`, `id`, `digest`, `extensionParams`, `logger`.
- **Returns**: `object.Object`, `error`.

```go
obj, err := ocfl.InitObject(ctx, writeFS, nil, version.V1_1, "urn:oid:1", checksum.DigestSHA512, nil, logger)
if err != nil {
    log.Fatal(err)
}
defer obj.Close()
```

---

## Orchestrated Actions

High-level actions combine multiple steps (like loading, validating, or extracting) into a single function call.

### `ValidateObject`
Performs a full validation of an OCFL object.
- **Usage**: `err := ocfl.ValidateObject(ctx, objectFS, logger)`
- **Details**: Loads the object and triggers the validation process (structural and fixity checks).

```go
err := ocfl.ValidateObject(ctx, objectFS, logger)
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

### `Extract`
Extracts files from an OCFL object to a destination filesystem.
- **Usage**: `err := ocfl.Extract(ctx, objectFS, destFS, path, version, withManifest, area, logger)`
- **Details**: Allows extracting specific versions and can include the manifest.

```go
// Extract the latest version of an object located at "path/to/object"
err := ocfl.Extract(ctx, fsys, destFS, "path/to/object", inventory.NewVersionNumber(), true, "", logger)
```

### `ExtractMeta`
Retrieves metadata from an OCFL object without extracting the content.
- **Usage**: `metadata, err := ocfl.ExtractMeta(ctx, fsys, path, logger)`
- **Returns**: `*inventory.Metadata` containing information about versions, state, and fixity.

```go
metadata, err := ocfl.ExtractMeta(ctx, fsys, "path/to/object", logger)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Object ID: %s, Latest Version: %s\n", metadata.ID, metadata.Version)
```

---

## Factories

If you need more control over the creation process, you can use the factories directly.

- `NewFactoryObject(ver, extensionFactory, logger)`: Returns a factory for creating objects of a specific OCFL version.
- `NewFactoryStorageRoot(ver, extensionFactory, logger)`: Returns a factory for creating storage roots of a specific OCFL version.

---
- [Back to OCFL Core Overview](../README.md)
- [Back to Project Root](../../README.md)
