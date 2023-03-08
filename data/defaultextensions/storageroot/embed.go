package storageroot

import "embed"

// go:embed initial/*.json
//
//go:embed NNNN-direct-clean-path-layout
var DefaultStorageRootExtensionFS embed.FS
