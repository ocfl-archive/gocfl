package docs

import (
	"embed"
)

//go:embed NNNN-*.md 0011-direct-clean-path-layout.md
//go:embed ocfl_spec_1.1.md
var ExtensionDocs embed.FS
