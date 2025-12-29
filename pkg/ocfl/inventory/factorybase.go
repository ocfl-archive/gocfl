package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
)

func NewFactoryBase() Factory {
	return &FactoryBase{}
}

type FactoryBase struct{}

func (f *FactoryBase) NewFixity(algorithms []checksum.DigestAlgorithm) Fixity {
	return NewFixityBase(f)
}

func (f *FactoryBase) NewUser() User {
	return NewUserBase()
}

func (f *FactoryBase) NewManifest() Manifest {
	return NewManifestBase(f)
}

func (f *FactoryBase) NewVersions() Versions {
	return NewVersionsBase(f)
}
func (f *FactoryBase) NewVersion() Version {
	return NewVersionBase(f)
}

func (f *FactoryBase) NewState() State {
	return NewStateBase()
}

var _ Factory = (*FactoryBase)(nil)
