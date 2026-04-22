package specs

import (
	_ "embed"
)

//go:embed mets.xsd
var METSXSD []byte

//go:embed xlink.xsd
var XLinkXSD []byte

//go:embed premis.xsd
var PremisXSD []byte
