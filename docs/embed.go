package docs

import (
	"embed"
)

//go:embed NNNN-*.md 00*.md initial.md
//go:embed ocfl_spec_1.1.md
var ExtensionDocs embed.FS
