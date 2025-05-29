package object

import "embed"

// go:embed NNNN-thumbnail/config.json
//
//go:embed NNNN-filesystem/config.json
//go:embed NNNN-indexer/config.json
//go:embed 0011-direct-clean-path-layout/config.json
//go:embed 0001-digest-algorithms/config.json
var DefaultObjectExtensionFS embed.FS
