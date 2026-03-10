package internal

import (
	"embed"
)

//go:embed siegfried/default.sig errors.toml
//go:embed extensions/object/*/* extensions/storageroot/*/*
var InternalFS embed.FS
