package config

import _ "embed"

//go:embed default.toml
var DefaultConfig []byte
