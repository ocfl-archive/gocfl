package defaultextensions

import "embed"

//go:embed object/*/*.json storageroot/*/*.json
var DefaultExtensionFS embed.FS
