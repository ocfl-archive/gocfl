package extensionimpldata

import "embed"

//go:embed initial/config.json NNNN-gocfl-extension-manager/config.json
var DefaultInitial embed.FS
