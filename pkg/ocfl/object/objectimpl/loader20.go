package objectimpl

import (
	"context"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewLoader20(ctx context.Context, factory factory.FactoryObject, config any, logger ocfllogger.OCFLLogger) *Loader20 {
	return &Loader20{
		Loader: NewLoader(ctx, factory, config, logger),
	}
}

type Loader20 struct {
	*Loader
}

func (l *Loader20) WithObject(o object.Object) object.Loader {
	l.Loader.WithObject(o)
	return l
}

var _ object.Loader = (*Loader20)(nil)
