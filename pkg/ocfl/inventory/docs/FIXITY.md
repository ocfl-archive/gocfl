# Fixity

The `Fixity` structural module of the inventory allows for the storage of additional checksums for files listed in the manifest. This serves long-term integrity assurance by allowing alternative algorithms (e.g., MD5, SHA-1, SHA-256) to be used in addition to the primary algorithm.

## Instantiation via Factory

As a structural module of the inventory, a `Fixity` should be created using the `pkg/ocfl/factory` package.

```go
fixity := f.NewFixity(ctx)
```

After creation, it is linked to the `Inventory` using the `WithFixity()` method:

```go
inv.WithFixity(fixity)
```

## OCFL Specification

The fixity block is used to store additional checksums using algorithms other than the main `digestAlgorithm`.

- **Specification**: [3.5.4 Fixity](../../../../data/specs/ocfl_1.1.md#354-fixity)

According to the OCFL specification, the `fixity` block is optional. It is a JSON object whose keys are the names of the used digest algorithms. Each of these keys points to an object that, in turn, maps digests to arrays of physical paths (similar to the manifest).

## Implementation: `Fixity` Interface

In the code, this is represented by the `Fixity` interface (`pkg/ocfl/inventory/fixity.go`).

### Important Methods

- `AddFile(manifestFilename string, digests map[checksum.DigestAlgorithm]string) (bool, error)`: Adds fixity data for a physical file.
- `GetDigestAlgorithms() iter.Seq[checksum.DigestAlgorithm]`: Returns all algorithms used in the fixity block.
- `GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error)`: Searches for files based on a specific algorithm and digest.
- `Checksums(s string) map[checksum.DigestAlgorithm]string`: Returns all available checksums for a specific file path.

## Concrete Implementation: `FixityBase`

The implementation in `pkg/ocfl/inventory/inventoryimpl/fixitybase.go` manages the structure `map[checksum.DigestAlgorithm]map[string][]string`.

### Example (JSON Structure)

```json
"fixity": {
  "md5": {
    "1dcc35...c31": [ "v1/content/foo/bar.xml" ]
  },
  "sha1": {
    "2dcc35...c31": [ "v1/content/foo/bar.xml" ]
  }
}
```

## Navigation

- [Back to README](../README.md)
- [Go to Inventory Documentation](INVENTORY.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Types Documentation](TYPES.md)
