# Creating custom OCFL Extensions

This document provides a detailed guide on creating custom extensions for the `gocfl` library. It supplements the technical documentation in [pkg/ocfl/extension/docs/EXTENSION.md](../pkg/ocfl/extension/docs/EXTENSION.md) with practical code examples.

## Basic Structure of an Extension

Each extension must implement the `extension.Extension` interface. Usually, you embed `extension.ExtensionConfig` to automatically handle standard fields like the name.

```go
package myextension

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const MyExtensionName = "NNNN-my-custom-extension"

// MyExtensionConfig holds the configuration stored in config.json.
type MyExtensionConfig struct {
	*extension.ExtensionConfig
	CustomField string `json:"customField"`
}

type MyExtension struct {
	*MyExtensionConfig
	logger ocfllogger.OCFLLogger
}

// NewMyExtension is the builder function for registration.
func NewMyExtension() (extension.Extension, error) {
	return &MyExtension{
		MyExtensionConfig: &MyExtensionConfig{
			ExtensionConfig: &extension.ExtensionConfig{
				ExtensionName: MyExtensionName,
			},
		},
	}, nil
}

func (m *MyExtension) GetName() string { return MyExtensionName }

func (m *MyExtension) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	m.logger = logger.With("extension", MyExtensionName)
	return m
}

func (m *MyExtension) Load(data json.RawMessage, extFS fs.FS) error {
	if err := json.Unmarshal(data, m.MyExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal config for %s", MyExtensionName)
	}
	return nil
}

func (m *MyExtension) SetParams(params map[string]string) error {
	// Optional: set parameters from the command line or environment
	if val, ok := params["my-param"]; ok {
		m.CustomField = val
	}
	return nil
}

func (m *MyExtension) WriteConfig(fsys appendfs.FS) error {
	// Stores the configuration in the extension's config.json
	writer, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot create config.json")
	}
	defer writer.Close()
	return json.NewEncoder(writer).Encode(m.MyExtensionConfig)
}

func (m *MyExtension) GetConfig() any { return m.MyExtensionConfig }
func (m *MyExtension) IsRegistered() bool { return true }
func (m *MyExtension) Terminate() error { return nil }
```

## Registration

The extension must be registered so it can be loaded by its name. This typically happens in an `init()` function.

```go
func init() {
	// Registration for Storage Roots (layouts etc.)
	extension.RegisterExtensionStorageRoot(MyExtensionName, NewMyExtension, GetMyParams, nil)
	// Registration for Objects (hooks, metadata etc.)
	extension.RegisterExtensionObject(MyExtensionName, NewMyExtension, GetMyParams, nil)
}

// GetMyParams defines external parameters (e.g., for the CLI)
func GetMyParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{
		{
			ExtensionName: MyExtensionName,
			Functions:     []string{"add", "update"},
			Param:         "my-param",
			Description:   "A custom parameter for my extension",
		},
	}, nil
}
```

## Extension Hooks

The actual power of extensions comes from implementing specific hook interfaces. The `ExtensionManager` detects these via type assertion.

### 1. Storage Root Path Hook
Used to determine where an object is stored in the Storage Root based on its ID.

```go
// Defined in pkg/ocfl/storageroot/storagerootExtension.go
// type ExtensionStorageRootPath interface { ... }

func (m *MyExtension) BuildStorageRootPath(sr storageroot.StorageRoot, id string) (string, error) {
	// Example: simple hashing of the ID for path creation
	hash := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x/%x/%s", hash[0:1], hash[1:2], id), nil
}

func (m *MyExtension) WriteLayout(fsys appendfs.FS) error {
	// Optional: write an ocfl_layout.json
	return nil
}
```

### 2. Object Content Path Hook
Transforms paths within the object (e.g., for encryption or deduplication).

```go
// Defined in pkg/ocfl/object/extension.go
func (m *MyExtension) BuildObjectManifestPath(originalPath string, area string) (string, error) {
	// Example: move all files into a 'data' subfolder
	return "data/" + originalPath, nil
}
```

### 3. Content Change Hooks
Enable actions before or after file operations (Add, Update, Delete).

```go
// Implements object.ExtensionContentChange
func (m *MyExtension) AddFileBefore(object object.VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	m.logger.Debug().Msgf("Before adding %s", source)
	return nil
}

func (m *MyExtension) AddFileAfter(versionWriter object.VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	m.logger.Debug().Msgf("After adding %s (Digest: %s)", internalPath, digest)
	return nil
}
```

### 4. Metadata Hook
Adds custom fields to the object's metadata output.

```go
// Implements object.ExtensionMetadata
func (m *MyExtension) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	return map[string]any{
		"custom-info": m.CustomField,
		"file-count":  len(obj.GetInventory().GetState()),
	}, nil
}
```

### 5. New Version Hook
Allows automatically triggering or initializing new versions.

```go
// Implements object.ExtensionNewVersion
func (m *MyExtension) NeedNewVersion(object object.VersionWriter) (bool, error) {
	// Example: force new version if a certain file exists
	return slices.Contains(object.GetInventory().GetStateFiles(), "trigger.txt"), nil
}

func (m *MyExtension) DoNewVersion(object object.VersionWriter) error {
	// Execute actions in the new (automatically created) version
	return object.AddFile(someFS, "auto-generated.txt", false, "", false, false)
}
```

### 6. Stream Hook
Allows access to a file's data stream during the write process. This can be used to extract metadata or, as in this example, to measure the length of the data stream. Since the data stream is split using `io.MultiWriter`, all registered stream hooks are served in parallel.

```go
// Implements object.ExtensionStream
func (m *MyExtension) StreamObject(object object.VersionWriter, reader io.Reader, stateFiles []string, dest string) error {
	// We use a counter writer or io.Copy to io.Discard to count the bytes
	// In this simple example, we just read everything and measure the length
	count, err := io.Copy(io.Discard, reader)
	if err != nil {
		return errors.Wrapf(err, "error reading stream for %s", dest)
	}
	m.logger.Info().Msgf("File %s has a stream length of %d bytes", dest, count)
	return nil
}
```

## Best Practices

1.  **Error Handling**: Use `emperror.dev/errors` to add context to errors.
2.  **Logging**: Use the logger set via `WithLogger`. It already contains the extension name as context.
3.  **Idempotency**: Hooks can be called multiple times. Ensure this has no negative side effects.
4.  **Performance**: Hooks like `BuildObjectManifestPath` are called for every file. Avoid expensive operations like I/O there if possible.
5.  **Documentation**: Use the `doc` field when registering to provide a Markdown string (via `go:embed`). This can be displayed by tools like `gocfl help extensions`.

---
- [Back to Extension Package](../pkg/ocfl/extension/README.md)
- [List of Object Hooks](../pkg/ocfl/object/docs/HOOKS.md)
