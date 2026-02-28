# State

The `State` of a version represents the logical view of the file structure at that specific version. It maps logical file paths to their corresponding digests (from the manifest).

## Instantiation via Factory

As a structural module of the inventory, a `State` should be created using the `pkg/ocfl/factory` package.

```go
state := f.NewState(ctx)
```

The state is typically managed within a [Version](VERSION.md) object and can be accessed via `version.GetState()`.

## OCFL Specification

The `state` is a mandatory part of each version object in the OCFL specification.

- **Specification**: [3.5.3.2 Version State](../../../../data/specs/ocfl_1.1.md#3532-version-state)

According to the specification, the state is a JSON object where each key is a digest (from the manifest) and each value is an array of logical paths that have that digest. This allows for:
- **Deduplication**: Multiple logical paths can point to the same digest.
- **Logical Mapping**: Decoupling the user-facing file structure from the internal physical storage.

## Implementation: `State` Interface

In the code, the state is represented by the `State` interface (`pkg/ocfl/inventory/state.go`).

### Important Methods

- `GetDigest(path string) string`: Returns the digest for a given logical path.
- `Paths(digest string) []string`: Returns all logical paths associated with a specific digest.
- `Iterate() func(yield func(digest string, logical []string) bool)`: Allows iterating through all entries in the state.
- `AddFile(path string, digest string) error`: Adds a logical path mapping to the state.
- `DeleteFile(path string) error`: Removes a logical path mapping.

## Concrete Implementation: `StateBase`

The implementation in `pkg/ocfl/inventory/inventoryimpl/statebase.go` manages the internal mapping (`map[string][]string`) and ensures that paths follow the OCFL conventions.

### Example (JSON Structure within a Version)

```json
"state": {
  "7dcc35...c31": [ "foo/bar.xml", "docs/manual.xml" ],
  "cf83e1...a3e": [ "empty.txt" ]
}
```

## Navigation

- [Back to README](../README.md)
- [Go to Inventory Documentation](INVENTORY.md)
- [Go to Version Documentation](VERSION.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
