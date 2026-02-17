package factoryimpl

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory(ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) factorytypes.Factory {
	switch ver {
	case version.Version1_0:
		return NewFactory10(extensionFactory, logger)
		//	case version.Version1_1:
		//		return NewFactory11(extensionFactory, logger)
	case version.Version2_0:
		return NewFactory20(extensionFactory, logger)
	default:
		return NewFactory11(extensionFactory, logger)
	}
}
