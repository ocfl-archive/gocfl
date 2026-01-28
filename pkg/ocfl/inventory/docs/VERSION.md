# Version

Every change to an OCFL object is recorded in a new `Version`. A version contains metadata about the change and the `state` of the files at that point in time.

## OCFL Specification (v1.1)

According to the OCFL 1.1 specification ([3.5.3.1 Version](../ocfl11.md#version)), a version object must include the following fields:

- `created`: Timestamp of creation (RFC3339).
- `message`: Optional description of the changes.
- `user`: Information about the creator (name and address).
- `state`: A mapping of digests (from the manifest) to logical file paths within this version.

## Implementation: `Version` Interface

In the code, a version is represented by the `Version` interface (`pkg/ocfl/inventory/version.go`).

### Important Methods

- `GetCreated() time.Time`: Returns the creation time.
- `GetMessage() string`: Returns the version message.
- `GetUser() User`: Returns the [User](TYPES.md#user) object.
- `GetState() State`: Returns the [State](TYPES.md#state) of the version.
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
- [Go to Manifest Documentation](MANIFEST.md)
- [Go to Fixity Documentation](FIXITY.md)
- [Go to Types Documentation](TYPES.md)
