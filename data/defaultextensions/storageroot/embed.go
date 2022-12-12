package storageroot

import "embed"

//go:embed */*.json
var DefaultStorageRootExtensionFS embed.FS
