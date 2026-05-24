package factoryimpl

import (
	"context"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object/objectimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject20(extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryObject {
	f := &factoryObject20{
		FactoryBaseObject: NewFactoryBaseObject(version.Version2_0, inventory.InventorySpec2_0, extensionFactory, logger).(*FactoryBaseObject),
		logger:            logger,
	}
	f.FactoryBaseObject.factory = f
	return f
}

type factoryObject20 struct {
	*FactoryBaseObject
	conf   map[object.ConfigName]any
	logger ocfllogger.OCFLLogger
}

func (f *factoryObject20) WithConfig(conf map[object.ConfigName]any) factory.FactoryObject {
	f.FactoryBaseObject.WithConfig(conf)
	f.conf = conf
	return f
}

func (f *factoryObject20) Copy() factory.FactoryObject {
	fNew := &factoryObject20{
		FactoryBaseObject: f.FactoryBaseObject.Copy().(*FactoryBaseObject),
		conf:              f.conf,
		logger:            f.logger,
	}
	fNew.FactoryBaseObject.factory = fNew
	return fNew
}

func (f *factoryObject20) NewLoader(ctx context.Context) object.Loader {
	return objectimpl.NewLoader20(
		ctx,
		f,
		f.GetConfig()[object.LoaderName],
		f.logger,
	)
}

func (f *factoryObject20) NewVersionWriter(ctx context.Context) object.VersionWriter {
	return objectimpl.NewVersionWriter20(
		ctx,
		f,
		f.GetConfig()[object.VersionWriterName],
		f.logger,
	)
}

var _ factory.FactoryObject = (*factoryObject20)(nil)
