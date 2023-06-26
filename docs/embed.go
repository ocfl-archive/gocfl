package docs

import (
	"embed"
)

//go:embed NNNN-*.md
//go:embed ocfl_spec_1.1.md
var ExtensionDocs embed.FS
