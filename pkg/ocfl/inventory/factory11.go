package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory11(extensionFactory *extension.ExtensionFactory, extensionManager extension.ExtensionManager, logger zLogger.ZLogger) types.Factory {
	return &factory11{
		logger:  logger,
		Factory: NewFactoryBase(version.Version1_1, types.InventorySpec1_1, extensionFactory, extensionManager, logger),
	}
}

type factory11 struct {
	types.Factory
	logger zLogger.ZLogger
}

var _ types.Factory = (*factory11)(nil)
