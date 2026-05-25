package objectimpl

import (
	"context"
	"regexp"

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
			versionFSMap:       fact.NewVersionFSMap(ctx),
		},
		config: validatorConfig,
	}
	return val
}

type validator20 struct {
	*validator
	config *Validator20Config
}

// WithObject attaches an OCFL object to the validator.
func (val *validator20) WithObject(o object.Object) object.Validator {
	val.validator.WithObject(o)
	return val
}

// allowedFilesRegexp20 matches the filenames allowed in an OCFL object's root or version directory.
var allowedFilesRegexp20 = regexp.MustCompile(`^(inventory.json(\.sha512|\.sha384|\.sha256|\.sha1|\.md5)?|0=ocfl_object_[0-9]+\.[0-9]+)$`)

var _ object.Validator = (*validator20)(nil)
