package ocfl

import "fmt"

type VersionPackagesType uint

const (
	VersionPlain VersionPackagesType = iota
	VersionZIP
	VersionTAR
	VersionTGZ
	VersionTBZ
)

var VersionPackageTypeString = map[VersionPackagesType]string{
	VersionPlain: "plain",
	VersionZIP:   "zip",
	VersionTAR:   "tar",
	VersionTGZ:   "tar.gz",
	VersionTBZ:   "tar.bz2",
}

var VersionPackageStringType = map[string]VersionPackagesType{
	VersionPackageTypeString[VersionPlain]: VersionPlain,
	VersionPackageTypeString[VersionZIP]:   VersionZIP,
	VersionPackageTypeString[VersionTAR]:   VersionTAR,
	VersionPackageTypeString[VersionTGZ]:   VersionTGZ,
	VersionPackageTypeString[VersionTBZ]:   VersionTBZ,
}

func (v VersionPackagesType) String() string {
	if str, ok := VersionPackageTypeString[v]; ok {
		return str
	}
	return fmt.Sprintf("unknown package type %d", v)
}
