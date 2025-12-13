package inventory

type FactoryBase struct{}

func (f *FactoryBase) NewVersions() Versions {
	return &VersionsBase{
		Versions: map[string]*VersionBase{},
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
