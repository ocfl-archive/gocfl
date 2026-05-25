package objectimpl

import (
	"context"
	"io/fs"
	"regexp"

	"emperror.dev/errors"
	"github.com/ocfl-archive/filesystem/pkg/zipfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

// Validator20Config holds the configuration for the OCFL 2.0 validator.
type Validator20Config struct{}

// NewObjectBaseValidator20 creates a new instance of an OCFL 2.0 object validator.
// It supports zipped version folders by overriding the getVersionFS mechanism.
func NewObjectBaseValidator20(ctx context.Context, fact factory.FactoryObject, conf any, logger ocfllogger.OCFLLogger) object.Validator {
	validatorConfig, ok := conf.(*Validator20Config)
	if conf != nil && !ok {
		logger.Error().Msg("invalid config type for validator")
	}
	val := &validator20{
		validator: &validator{
			ctx:                ctx,
			factory:            fact,
			logger:             logger.With("task", "validator"),
			config:             (*ValidatorConfig)(validatorConfig),
			allowedFilesRegexp: allowedFilesRegexp20,
		},
		config: validatorConfig,
	}
	val.getVersionFS = val._getVersionFS
	return val
}

// validator20 is the internal implementation of the OCFL 2.0 object validator.
type validator20 struct {
	*validator
	config *Validator20Config
}

// _getVersionFS overrides the default implementation to support both directories and .zip files for versions.
func (val *validator20) _getVersionFS(version string) (fs.FS, error) {
	fsys := val.GetReadFS()
	fi, err := fs.Stat(fsys, version)
	if err == nil && fi.IsDir() {
		return val.validator._getVersionFS(version)
	}
	fi, err = fs.Stat(fsys, version+".zip")
	if err == nil && !fi.IsDir() {
		return zipfs.NewFSFile(fsys, version+".zip", val.logger.Logger())
	}
	return nil, errors.Errorf("folders %s and %s.zip not found", version, version)
}

// WithObject attaches an OCFL object to the validator.
func (val *validator20) WithObject(o object.Object) object.Validator {
	val.validator.WithObject(o)
	return val
}

// allowedFilesRegexp20 matches the filenames allowed in an OCFL object's root or version directory.
var allowedFilesRegexp20 = regexp.MustCompile(`^(inventory.json(\.sha512|\.sha384|\.sha256|\.sha1|\.md5)?|0=ocfl_object_[0-9]+\.[0-9]+)$`)

var _ object.Validator = (*validator20)(nil)
