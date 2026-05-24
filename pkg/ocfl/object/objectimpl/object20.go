package objectimpl

import (
	"context"
	"io/fs"

	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewObject20(ctx context.Context, fact factory.FactoryObject, extensionFactory extension.Factory[object.ExtensionManager], config any, logger ocfllogger.OCFLLogger) object.Object {
	return &Object20{
		ObjectBase: NewObjectBase(ctx, fact, version.Version2_0, extensionFactory, config, logger).(*ObjectBase),
	}
}

type Object20 struct {
	*ObjectBase
}

func (f *Object20) WithInventory(inv inventory.Inventory) object.Object {
	f.ObjectBase.WithInventory(inv)
	return f
}

func (f *Object20) WithExtensionManager(manager object.ExtensionManager) object.Object {
	f.ObjectBase.WithExtensionManager(manager)
	return f
}

func (f *Object20) WithReadFS(fsys fs.FS) object.Object {
	f.ObjectBase.WithReadFS(fsys)
	return f
}

func (f *Object20) WithWriteFS(fsys appendfs.FS) object.Object {
	f.ObjectBase.WithWriteFS(fsys)
	return f
}

var _ object.Object = (*Object20)(nil)
