# Version and Versions

The `Versions` structural module of the inventory manages the history of an OCFL object. Every change is recorded in a new `Version`, which contains metadata about the change and the `state` of the files at that point in time.

## Instantiation via Factory

As structural modules of the inventory, `Versions` and individual `Version` objects should be created using the `pkg/ocfl/factory` package.

```go
versions := f.NewVersions(ctx)
version := f.NewVersion(ctx, vNumber)
```

After creation, the `Versions` container is linked to the `Inventory` using the `WithVersions()` method:

```go
inv.WithVersions(versions)
```

## OCFL Specification

Each version directory in an OCFL object contains a record of the state of the object at that version.

- **Specification**: [3.5.3 Versions](../../../../data/specs/ocfl_1.1.md#353-versions)

According to the OCFL specification, a version object must include the following fields:

- `created`: Timestamp of creation (RFC3339).
- `message`: Optional description of the changes.
- `user`: Information about the creator (name and address).
- `state`: A mapping of digests (from the manifest) to logical file paths within this version.

## Implementation: `Version` and `Versions` Interfaces

In the code, a version is represented by the `Version` interface (`pkg/ocfl/inventory/version.go`) and the collection by the `Versions` interface (`pkg/ocfl/inventory/versions.go`).

### Important Methods

- `GetCreated() time.Time`: Returns the creation time.
- `GetMessage() string`: Returns the version message.
- `GetUser() User`: Returns the [User](TYPES.md#user) object.
- `GetState() State`: Returns the [State](STATE.md) of the version.
- `GetVersionNumber() *VersionNumber`: Returns the [VersionNumber](TYPES.md#versionnumber) (e.g., v1).

## Concrete Implementation: `VersionBase`

The default implementation is found in `pkg/ocfl/inventory/inventoryimpl/versionbase.go`. It manages:
- Storage of metadata.
- Access to the `State` object, which manages the logical file structure.

### Example (JSON Structure within `versions`)

```json
"v1": {
  "created": "2024-01-28T17:00:00Z",
  "message": "Initial ingest",
  "user": {
    "name": "Junie",
    "address": "mailto:junie@example.com"
  },
  "state": {
    "7dcc35...c31": [ "foo/bar.xml" ],
    "cf83e1...a3e": [ "empty.txt" ]
  }
}
```

## Navigation

- [Back to README](../README.md)
- [Go to Inventory Documentation](INVENTORY.md)
- [Go to State Documentation](STATE.md)
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
