package inventory

import (
	"context"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec InventorySpec, logger zLogger.ZLogger) Factory {
	return &FactoryBase{
		logger:  logger,
		version: version,
		spec:    spec,
	}
}

type FactoryBase struct {
	logger  zLogger.ZLogger
	version version.OCFLVersion
	spec    InventorySpec
}

func (f *FactoryBase) NewInventory(ctx context.Context, objectFolder string, contentDir string) Inventory {
	if contentDir == "" {
		contentDir = "content"
	}
	i := &InventoryBase{
		ctx:     ctx,
		factory: f,
		//object:                 object,
		version:       f.version,
		folder:        objectFolder,
		paddingLength: 0,
		//fixityDigestAlgorithms: []checksum.DigestAlgorithm{},
		Type:             f.spec,
		Head:             NewVersionNumber(),
		ContentDirectory: contentDir,
		Manifest:         f.NewManifest(),
		Versions:         f.NewVersions(),
		Fixity:           f.NewFixity(),
		logger:           f.logger,
	}
	return i
}

func (f *FactoryBase) NewFixity() Fixity {
	return NewFixityBase()
}

func (f *FactoryBase) NewUser() User {
	return NewUserBase()
}

func (f *FactoryBase) NewManifest() Manifest {
	return NewManifestBase()
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
