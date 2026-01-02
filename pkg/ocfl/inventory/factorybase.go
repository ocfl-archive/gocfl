package inventory

import (
	"context"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec types.InventorySpec, logger zLogger.ZLogger) types.Factory {
	return &FactoryBase{
		logger:  logger,
		version: version,
		spec:    spec,
	}
}

type FactoryBase struct {
	logger  zLogger.ZLogger
	version version.OCFLVersion
	spec    types.InventorySpec
}

func (f *FactoryBase) NewInventory(ctx context.Context, objectFolder string, contentDir string) types.Inventory {
	if contentDir == "" {
		contentDir = "content"
	}
	i := &InventoryBase{
		ctx:     ctx,
		factory: f,
		//object:                 object,
		version: f.version,
		folder:  objectFolder,
		//paddingLength: 0,
		//fixityDigestAlgorithms: []checksum.DigestAlgorithm{},
		Type:             f.spec,
		Head:             types.NewVersionNumber(),
		ContentDirectory: contentDir,
		Manifest:         f.NewManifest(),
		Versions:         f.NewVersions(),
		Fixity:           f.NewFixity(),
		logger:           f.logger,
	}
	return i
}

func (f *FactoryBase) NewFixity() types.Fixity {
	return NewFixityBase()
}

func (f *FactoryBase) NewUser() types.User {
	return NewUserBase()
}

func (f *FactoryBase) NewManifest() types.Manifest {
	return NewManifestBase()
}

func (f *FactoryBase) NewVersions() types.Versions {
	return NewVersionsBase(f)
}
func (f *FactoryBase) NewVersion() types.Version {
	return NewVersionBase(f)
}

func (f *FactoryBase) NewState() types.State {
	return NewStateBase()
}

var _ types.Factory = (*FactoryBase)(nil)
