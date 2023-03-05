package object

import "embed"

//go:embed NNNN-content-subpath/*.json
//go:embed initial/*.json
//go:embed NNNN-direct-clean-path-layout/*.json
//go:embed 0001-digest-algorithms/*.json
//go:embed NNNN-indexer/*.json
var DefaultObjectExtensionFS embed.FS
