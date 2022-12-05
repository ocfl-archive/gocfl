package object

import "embed"

//go:embed */*.json
var DefaultObjectExtensionFS embed.FS
