package docs

import (
	"embed"
)

//go:embed NNNN-content-subpath.md
//go:embed NNNN-direct-clean-path-layout.md
//go:embed NNNN-indexer.md
//go:embed NNNN-gocfl-extension-manager.md
var ExtensionDocs embed.FS
