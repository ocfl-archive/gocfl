package specs

import (
	_ "embed"
)

//go:embed ocfl_1.1.md
var OCFL1_1 []byte

//go:embed mets.xsd
var METSXSD []byte

//go:embed xlink.xsd
var XLinkXSD []byte

//go:embed premis.xsd
var PremisXSD []byte
