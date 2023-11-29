package object

import "embed"

//go:embed NNNN-filesystem/config.json
//go:embed NNNN-indexer/config.json
//go:embed NNNN-direct-clean-path-layout/config.json
//go:embed 0001-digest-algorithms/config.json
var DefaultObjectExtensionFS embed.FS
