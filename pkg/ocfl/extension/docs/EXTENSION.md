# Extension Interface

An `Extension` (`extension.go`) is a single functional unit that implements a specific OCFL extension (e.g., storage layout, fixity, or metadata).

- **Specification**: [OCFL 1.1: 5. Extensions](../../../../data/specs/ocfl_1.1.md#5-extensions)
- **External Docs**: [OCFL Extensions](https://ocfl.io/extensions/)

## Interface Definition

Key methods of the `Extension` interface:
- `GetName() string`: Returns the unique name of the extension (e.g., `NNNN-direct-clean-path-layout`).
- `Load(fsys fs.FS) error`: Loads the extension's configuration from a filesystem.
- `SetParams(params map[string]string) error`: Configures the extension via key-value parameters.
- `WriteConfig(fsys appendfs.FS) error`: Persists the extension's configuration to the OCFL object or storage root.
- `GetConfig() any`: Returns the extension's configuration object.
- `Terminate() error`: Performs cleanup when the extension is no longer needed.

## Usage

Each extension should implement this interface. Common implementations in this library (located in [`pkg/extension`](../../../../../gocfl-extensions/extension)) include:
- **Storage layout extensions**: (e.g., [`0002-flat-direct-storage-layout`](../../../../../gocfl-extensions/extension/0002-flat-direct-storage-layout.go) - [Doc](../../../../docs/0011-direct-clean-path-layout.md)). These are used within [Storage Roots](../../storageroot/README.md) to determine object placement.
- **Fixity extensions**: For additional checksums (e.g., [`0001-digest-algorithms`](../../../../../gocfl-extensions/extension/0001-digest-algorithms.go)).
- **Metadata extensions**: For custom descriptive or technical metadata (e.g., [`NNNN-mets`](../../../../../gocfl-extensions/extension/NNNN-mets.go) - [Doc](../../../../docs/NNNN-mets.md)).
- **Custom functional extensions**: (e.g., [`NNNN-thumbnail`](../../../../../gocfl-extensions/extension/NNNN-thumbnail.go) - [Doc](../../../../docs/NNNN-thumbnail.md)).

Extensions are typically managed by a [Manager](MANAGER.md) (either for [Objects](../../object/docs/OBJECT.md) or [Storage Roots](../../storageroot/docs/STORAGEROOT.md)) and created by a [Factory](FACTORY.md).

---
- [Back to Extension Overview](../README.md)
- [OCFL Object Structure](../../object/docs/OBJECT.md)
- [OCFL Storage Root](../../storageroot/docs/STORAGEROOT.md)
