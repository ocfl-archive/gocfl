package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase() *FactoryBase {
	return &FactoryBase{}
}

type FactoryBase struct{}

func (f *FactoryBase) NewFixity(algorithms []checksum.DigestAlgorithm) Fixity {
	return NewFixityBase(algorithms)
}

func (f *FactoryBase) NewUser() User {
	return NewUserBase("", "")
}

func (f *FactoryBase) NewManifest() Manifest {
	return NewManifestBase()
}

func (f *FactoryBase) NewVersions() Versions {
	return &VersionsBase{
		Versions: map[version.OCFLVersion]*VersionBase{},
	}
}
func (f *FactoryBase) NewVersion() Version {
	return &VersionBase{}
}

func (f *FactoryBase) NewState() State {
	return &StateBase{
		State: map[string][]string{},
	}
}

var _ Factory = (*FactoryBase)(nil)
