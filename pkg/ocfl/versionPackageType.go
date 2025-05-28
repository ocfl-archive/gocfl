package ocfl

import "fmt"

type VersionPackageType uint

const (
	VersionPlain VersionPackageType = iota
	VersionZIP
	VersionTGZ
	VersionTBZ
)

var VersionPackageTypeString = map[VersionPackageType]string{
	VersionPlain: "plain",
	VersionZIP:   "zip",
	VersionTGZ:   "tar.gz",
	VersionTBZ:   "tar.bz2",
}

var VersionPackageStringType = map[string]VersionPackageType{
	VersionPackageTypeString[VersionPlain]: VersionPlain,
	VersionPackageTypeString[VersionZIP]:   VersionZIP,
	VersionPackageTypeString[VersionTGZ]:   VersionTGZ,
	VersionPackageTypeString[VersionTBZ]:   VersionTBZ,
}

func (v VersionPackageType) String() string {
	if str, ok := VersionPackageTypeString[v]; ok {
		return str
	}
	return fmt.Sprintf("unknown package type %d", v)
}
