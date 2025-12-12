package validation

import (
	"context"
	"runtime"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewValidator(ctx context.Context, ver version.OCFLVersion, contextString string, logger zLogger.ZLogger) (*Validator, error) {
	return &Validator{
		ctx:     ctx,
		ver:     ver,
		logger:  logger,
		context: contextString,
	}, nil
}

type Validator struct {
	ctx     context.Context
	ver     version.OCFLVersion
	logger  zLogger.ZLogger
	context string
}

func (v *Validator) AddValidationError(errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(v.ver, errno).AppendDescription(format, a...).AppendContext(v.context)
	_, file, line, _ := runtime.Caller(1)
	v.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(AddValidationErrors(v.ctx, valError))
}

func (v *Validator) AddValidationWarning(errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(v.ver, errno).AppendDescription(format, a...).AppendContext(v.context)
	_, file, line, _ := runtime.Caller(1)
	v.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(AddValidationWarnings(v.ctx, valError))
}

var _ Validation = (*Validator)(nil)
