# Manifest

The `Manifest` is the structural module of the inventory that bridges the gap between file digests (checksums) and their actual physical storage locations within the OCFL object.

## Instantiation via Factory

As a structural module of the inventory, a `Manifest` should be created using the `pkg/ocfl/factory` package.

```go
manifest := f.NewManifest(ctx)
```

After creation, it is linked to the `Inventory` using the `WithManifest()` method:

```go
inv.WithManifest(manifest)
```

## OCFL Specification

The manifest is a mapping of digests to physical paths within the OCFL object.

- **Specification**: [3.5.2 Manifest](../../../../data/specs/ocfl_1.1.md#352-manifest)

According to the OCFL specification, the manifest is a JSON object. Each key in the manifest is a digest value (calculated using the algorithm specified in `digestAlgorithm`). The value for each key is an array of paths relative to the OCFL object root directory.

### Path Rules:
- The separator must be a forward slash (`/`).
- Paths must not begin or end with `/`.
- Path elements must not be `.`, `..`, or empty.

## Implementation: `Manifest` Interface

In the code, the manifest is represented by the `Manifest` interface (`pkg/ocfl/inventory/manifest.go`).

### Important Methods

- `AddFile(filename string, digest string) (bool, error)`: Registers a file with its digest in the manifest.
- `GetFiles(digest string) ([]string, error)`: Returns all physical paths associated with this digest.
- `Iterate() func(yield func(digest string, internal []string) bool)`: Allows iterating through all entries in the manifest.
- `GetFilesFlat() iter.Seq[string]`: Returns a flat list of all physical file paths.

## Concrete Implementation: `ManifestBase`

The implementation in `pkg/ocfl/inventory/inventoryimpl/manifestbase.go` ensures that paths comply with OCFL conventions and manages the internal mapping (`map[string][]string`).

### Example (JSON Structure)

```json
"manifest": {
  "7dcc35...c31": [ "v1/content/foo/bar.xml" ],
  "cf83e1...a3e": [ "v1/content/empty.txt" ],
  "ffccf6...62e": [ "v2/content/image.tiff" ]
}
```

## Navigation

- [Back to README](../README.md)
- [Go to Inventory Documentation](INVENTORY.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
