package inventory

type FactoryBase struct{}

func (f *FactoryBase) NewUser() User {
	return NewUserBase("", "")
}

func (f *FactoryBase) NewManifest() Manifest {
	return NewManifestBase()
}

func (f *FactoryBase) NewVersions() Versions {
	return &VersionsBase{
		Versions: map[string]*VersionBase{},
	}
}
func (f *FactoryBase) NewVersion() Version {
	return &VersionBase{}
}

func (f *FactoryBase) NewState() State {
	return &StateManifestBase{
		State: map[string][]string{},
	}
}

var _ Factory = (*FactoryBase)(nil)
