# OCFL Demo Examples

This directory contains runnable Go examples for common OCFL operations using the `gocfl` library. These examples are based on the [Quickstart Guide](../docs/quickstart.md).

## Examples

1. **[Storage Root Initialization](./storageroot_init/main.go)**
   - Demonstrates how to initialize a new OCFL Storage Root.
2. **[Object Creation](object_add/main.go)**
   - Demonstrates how to create and initialize a new OCFL Object within a storage root.
3. **[Adding and Updating Content](object_upd/main.go)**
   - Demonstrates how to add files to an object and create subsequent versions with updated content.
4. **[Object Validation](object_validate/main.go)**
   - Demonstrates how to validate an existing OCFL Object.
5. **[Extracting Objects](./object_extract/main.go)**
   - Demonstrates how to extract the content of an OCFL Object to a destination filesystem.

## How to Run

Each example is a self-contained Go program that uses the local filesystem. You need to provide the path to the OCFL Storage Root via the `-path` parameter.

```bash
# 1. Initialize a Storage Root
go run github.com/ocfl-archive/gocfl/v3/demo/storageroot_init/main.go@latest -path ./my-storage-root

# 2. Create an Object
go run github.com/ocfl-archive/gocfl/v3/demo/object_add/main.go@latest -path ./my-storage-root -id my-object-1

# 3. Add and Update Content
go run github.com/ocfl-archive/gocfl/v3/demo/object_upd/main.go@latest -path ./my-storage-root -id my-object-1

# 4. Validate an Object
go run github.com/ocfl-archive/gocfl/v3/demo/object_validate/main.go@latest -path ./my-storage-root -id my-object-1

# 5. Extract an Object
go run github.com/ocfl-archive/gocfl/v3/demo/object_extract/main.go@latest -path ./my-storage-root -id my-object-1 -dest ./extraction-dir
```

## Parameters

- `-path`: Absolute or relative path to the OCFL Storage Root.
- `-id`: (Optional for most) The OCFL Object ID to create, update, or extract. Defaults to `my-object-id`.
- `-dest`: (For extraction only) The destination directory where the object content will be extracted.

## Note on Filesystems

These examples use `vfsrw` with the `os` backend to interact with your local filesystem. You can use these examples as a starting point for building your own OCFL-aware applications.
