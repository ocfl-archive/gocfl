package factoryimpl

import (
	"fmt"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewDynamicFactory(ver version.OCFLVersion, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &dynamicFactory{
		Factory:          NewFactory(ver, extensionFactory, logger),
		extensionFactory: extensionFactory,
		logger:           logger,
	}
}

type dynamicFactory struct {
	factory.Factory
	extensionFactory *extensionimpl.Factory
	logger           ocfllogger.OCFLLogger
}

func (f *dynamicFactory) SetVersion(ver version.OCFLVersion) error {
	if !version.ValidVersion(ver) {
		return fmt.Errorf("invalid version: %s", ver)
	}
	if f.Factory.GetVersion() == ver {
		return nil
	}
	f.Factory = NewFactory(ver, f.extensionFactory, f.logger)
	return nil
}

func (f *dynamicFactory) Copy() factory.Factory {
	return &dynamicFactory{
		Factory:          f.Factory.Copy(),
		extensionFactory: f.extensionFactory,
		logger:           f.logger,
	}
}

var _ factory.Factory = (*dynamicFactory)(nil)
