package objectimpl

import (
	"context"

	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type Extractor20Config struct{}

func NewExtractor20(ctx context.Context, fact factory.FactoryObject, conf any, logger ocfllogger.OCFLLogger) object.Extractor {
	extractorConfig, ok := conf.(*Extractor20Config)
	if conf != nil && !ok {
		logger.Error().Msg("invalid config type for extractor")
	}
	ext := &extractor20{
		extractor: &extractor{
			ctx:     ctx,
			factory: fact,
			logger:  logger.With("task", "extractor"),
			config:  (*ExtractorConfig)(extractorConfig),
		},
		config: extractorConfig,
	}
	ext.versionFSMap = fact.NewVersionFSMap(ctx)
	return ext
}

type extractor20 struct {
	*extractor
	config *Extractor20Config
}

func (ext *extractor20) WithObject(o object.Object) object.Extractor {
	ext.extractor.WithObject(o)
	return ext
}

func (ext *extractor20) WithDestFS(destFS appendfs.FS) object.Extractor {
	ext.extractor.WithDestFS(destFS)
	return ext
}

var _ object.Extractor = (*extractor20)(nil)
