package internal

import (
	"embed"
)

//go:embed errors.toml
//go:embed extensions/object/*/* extensions/storageroot/*/*
//go:embed thumbnail/scripts/* thumbnail/thumbnail.toml
var InternalFS embed.FS
