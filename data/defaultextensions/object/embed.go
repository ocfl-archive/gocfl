package object

import "embed"

//go:embed initial/*.json
//go:embed NNNN-content-subpath/*.json
//go:embed NNNN-direct-clean-path-layout/*.json
var DefaultObjectExtensionFS embed.FS
