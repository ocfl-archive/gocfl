package objectimpl

import (
	"context"
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewLoader(ctx context.Context, factory factory.Factory, logger ocfllogger.OCFLLogger) *Loader {
	return &Loader{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "loader"),
	}
}

type Loader struct {
	object.Object
	ctx              context.Context
	factory          factory.Factory
	extensionFactory extension.Factory
	sourceFS         fs.FS
	logger           ocfllogger.OCFLLogger
}

func (l *Loader) Load() error {
	//TODO implement me
	panic("implement me")
}

func (l *Loader) WithExtensionFactory(factory extension.Factory) object.Loader {
	l.extensionFactory = factory
	return l
}

func (l *Loader) WithObject(o object.Object) object.Loader {
	l.Object = o
	return l
}

func (l *Loader) WithFS(sourceFS fs.FS) object.Loader {
	l.sourceFS = sourceFS
	return l
}

var _ object.Loader = (*Loader)(nil)
