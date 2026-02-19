package ocfllogger

import (
	"context"
	"strings"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/rs/zerolog"
	"golang.org/x/exp/maps"
)

func NewOCFLLogger(ctx context.Context, logger zLogger.ZLogger, data map[string]string) *OCFLLoggerImpl {
	if data == nil {
		data = make(map[string]string)
	}
	return &OCFLLoggerImpl{
		ctx:     ctx,
		ZLogger: logger,
		data:    data,
	}
}

type OCFLLogger interface {
	Trace() *zerolog.Event
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Err(err error) *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
	With(name, value string) *OCFLLoggerImpl
	ValidationError(ver version.OCFLVersion, code validation.ValidationErrorCode, format string, a ...interface{}) *OCFLLoggerImpl
}

type OCFLLoggerImpl struct {
	zLogger.ZLogger
	data map[string]string
	ctx  context.Context
}

func (l *OCFLLoggerImpl) With(name, value string) *OCFLLoggerImpl {
	newData := maps.Clone(l.data)
	newData[strings.ToLower(name)] = value
	return NewOCFLLogger(l.ctx, new(l.ZLogger.With().Str(name, value).Logger()), newData)
}

func (l *OCFLLoggerImpl) ValidationError(ver version.OCFLVersion, code validation.ValidationErrorCode, format string, a ...interface{}) *OCFLLoggerImpl {
	validationError := validation.GetValidationError(ver, code).AppendContext(format, a...)
	var event *zerolog.Event
	validation.AddValidationErrors(l.ctx, validation.GetValidationError(ver, code).AppendDescription(format, a...))
	if validationError.Code[0] == 'W' {
		event = l.ZLogger.Warn()
	} else {
		event = l.ZLogger.Error()
	}
	event.Str("OCFLVersion", ver.String()).
		Str("validationErrorCode", string(validationError.Code)).
		Str("validationErrorMessage", validationError.Description).
		Msgf(format, a...)
	return l
}

var _ OCFLLogger = (*OCFLLoggerImpl)(nil)
