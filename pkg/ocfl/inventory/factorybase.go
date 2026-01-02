package inventory

import (
	"context"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/interfaces"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec InventorySpec, logger zLogger.ZLogger) interfaces.Factory {
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

func (f *FactoryBase) NewInventory(ctx context.Context, objectFolder string, contentDir string) interfaces.Inventory {
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
		Head:             interfaces.NewVersionNumber(),
		ContentDirectory: contentDir,
		Manifest:         f.NewManifest(),
		Versions:         f.NewVersions(),
		Fixity:           f.NewFixity(),
		logger:           f.logger,
	}
	return i
}

func (f *FactoryBase) NewFixity() interfaces.Fixity {
	return NewFixityBase()
}

func (f *FactoryBase) NewUser() interfaces.User {
	return NewUserBase()
}

func (f *FactoryBase) NewManifest() interfaces.Manifest {
	return NewManifestBase()
}

func (f *FactoryBase) NewVersions() interfaces.Versions {
	return NewVersionsBase(f)
}
func (f *FactoryBase) NewVersion() interfaces.Version {
	return NewVersionBase(f)
}

func (f *FactoryBase) NewState() interfaces.State {
	return NewStateBase()
}

var _ interfaces.Factory = (*FactoryBase)(nil)
