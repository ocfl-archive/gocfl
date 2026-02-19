package factoryimpl

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewDynamicFactory(ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &dynamicFactory{
		Factory:          NewFactory(ver, extensionFactory, logger),
		extensionFactory: extensionFactory,
		logger:           logger,
	}
}

type dynamicFactory struct {
	factory.Factory
	extensionFactory *extensionimpl.ExtensionFactory
	logger           ocfllogger.OCFLLogger
}

func (f *dynamicFactory) SetVersion(ver version.OCFLVersion) {
	if f.Factory.GetVersion() == ver {
		return
	}
	f.Factory = NewFactory(ver, f.extensionFactory, f.logger)
}

func (f *dynamicFactory) Copy() factory.Factory {
	return &dynamicFactory{
		Factory:          f.Factory.Copy(),
		extensionFactory: f.extensionFactory,
		logger:           f.logger,
	}
}

var _ factory.Factory = (*dynamicFactory)(nil)
