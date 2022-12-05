package storageroot

import "embed"

//go:embed */*.json
var DefaultStoragerootExtensionFS embed.FS
