package storageroot

import "embed"

// go:embed initial/*.json
//
//go:embed NNNN-direct-clean-path-layout/*.json
//go:embed NNNN-gocfl-extension-manager/*.json
//go:embed initial/*.json
var DefaultStorageRootExtensionFS embed.FS
