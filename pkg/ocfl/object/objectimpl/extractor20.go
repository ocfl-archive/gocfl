package objectimpl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/zipfs"
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
	ext.getVersionFS = ext._getVersionFS
	return ext
}

type extractor20 struct {
	*extractor
	config *Extractor20Config
	//	versionReadFS fs.FS
}

func (ext *extractor20) _getVersionFS(version string) (fs.FS, error) {
	fi, err := fs.Stat(ext.objectReadFS, version)
	if err == nil && !fi.IsDir() {
		return ext.extractor._getVersionFS(version)
	}
	fi, err = fs.Stat(ext.objectReadFS, version+".zip")
	if err == nil && !fi.IsDir() {
		return zipfs.NewFSFile(ext.objectReadFS, version+".zip", ext.logger.Logger())
	}
	return nil, errors.Errorf("folders %s and %s.zip not found", version, version)
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
