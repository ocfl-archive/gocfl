package internal

import (
	"embed"
)

//go:embed errors.toml
//go:embed extensions/object/*/* extensions/storageroot/*/*
var InternalFS embed.FS
