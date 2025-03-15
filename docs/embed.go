package docs

import (
	"embed"
)

//go:embed NNNN-*.md
//go:embed ocfl_spec_1.1.md
var FullDocs embed.FS

//go:embed ocfl_spec_1.1.md
var OCFLDocs embed.FS

var Documentations = map[string]embed.FS{
	"full": FullDocs,
	"ocfl": OCFLDocs,
}
