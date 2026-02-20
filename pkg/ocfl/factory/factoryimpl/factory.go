package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewFactory(ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, logger ocfllogger.OCFLLogger) factorytypes.Factory {
	switch ver {
	case version.Version1_0:
		return NewFactory10(extensionFactory, logger)
	case version.Version1_1:
		return NewFactory11(extensionFactory, logger)
	case version.Version2_0:
		return NewFactory20(extensionFactory, logger)
		// todo: should we do a default??? or add errors
	default:
		return NewFactory11(extensionFactory, logger)
	}
}
