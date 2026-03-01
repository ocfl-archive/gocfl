# appendfs

The `appendfs` package provides an extended file system interface that goes beyond the standard functionality of `io/fs`. The primary goal of this interface is to ensure that the provided file system is capable of creating new files and directories.

## Interface: FS

The `FS` interface combines several interfaces from the standard `io/fs` package and the external `github.com/je4/filesystem/v3/pkg/writefs` package. It serves as a central abstraction for file systems that can be both read from and written to.

```go
type FS interface {
    fs.FS
    writefs.MkDirFS
    writefs.CreateFS
}
```

It includes the following functionalities:
- **`fs.FS`**: Standard read access (`Open(name string) (File, error)`).
- **`writefs.MkDirFS`**: Creation of directories.
- **`writefs.CreateFS`**: Creation of files for write access.

## Functions

### `Sub(fsys FS, path string) (FS, error)`

The `Sub` function creates a sub-file system (`Sub-FS`) based on a given path. It ensures that the resulting file system also satisfies the `FS` interface.

This is particularly useful for restricting operations to a specific subdirectory of an OCFL storage or an object, while still maintaining write permissions.

**Example:**

```go
subFS, err := appendfs.Sub(rootFS, "v1/content")
if err != nil {
    // Error handling
}

// Now you can work in subFS as if "v1/content" were the root.
f, err := subFS.Create("data.txt")
```

### `EnsureFS(fsys fs.FS) (FS, error)`

The helper function `EnsureFS` checks if a given `fs.FS` implements the `appendfs.FS` interface (i.e., supports write access). If it does, it returns the casted interface; otherwise, an error is reported.

## Project Usage

`appendfs.FS` is used in many places within `gocfl` to abstract the storage layer. It allows other packages (such as `ocfl`, `extension`, etc.) to create files and directories without having to worry about the underlying implementation (local file system, S3, etc.).
