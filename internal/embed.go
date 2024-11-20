package internal

import (
	"embed"
)

//go:embed siegfried/default.sig errors.toml
var InternalFS embed.FS
