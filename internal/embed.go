package internal

import (
	"embed"
)

//go:embed siegfried/default.sig
var InternalFS embed.FS
