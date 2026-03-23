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

func NewOCFLLogger(ctx context.Context, logger zLogger.ZLogger, data map[string]string, ver version.OCFLVersion, validationStatus *validation.Status) *OCFLLoggerImpl {
	if data == nil {
		data = make(map[string]string)
	}
	if validationStatus == nil {
		validationStatus = validation.NewStatus()
	}
	return &OCFLLoggerImpl{
		ctx:              ctx,
		ZLogger:          logger,
		data:             data,
		ver:              ver,
		validationStatus: validationStatus,
	}
}

type OCFLLogger interface {
	Logger() zLogger.ZLogger
	Trace() *zerolog.Event
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Err(err error) *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
	With(name, value string) OCFLLogger
	WithVersion(ver version.OCFLVersion) OCFLLogger
	ValidationError(code validation.ErrorCode, format string, a ...interface{}) *OCFLLoggerImpl
	ValidationErrors() []*validation.Error
	ClearValidationErrors()
}

type OCFLLoggerImpl struct {
	zLogger.ZLogger
	data             map[string]string
	ctx              context.Context
	ver              version.OCFLVersion
	validationStatus *validation.Status
}

func (l *OCFLLoggerImpl) ValidationErrors() []*validation.Error {
	return l.validationStatus.Errors
}

func (l *OCFLLoggerImpl) ClearValidationErrors() {
	l.validationStatus.Errors = make([]*validation.Error, 0)
}

func (l *OCFLLoggerImpl) WithVersion(ver version.OCFLVersion) OCFLLogger {
	return NewOCFLLogger(l.ctx, l.ZLogger, l.data, ver, l.validationStatus)
}

func (l *OCFLLoggerImpl) Logger() zLogger.ZLogger {
	return l.ZLogger
}

func (l *OCFLLoggerImpl) With(name, value string) OCFLLogger {
	newData := maps.Clone(l.data)
	newData[strings.ToLower(name)] = value
	return NewOCFLLogger(l.ctx, new(l.ZLogger.With().Str(name, value).Logger()), newData, l.ver, l.validationStatus)
}

func (l *OCFLLoggerImpl) ValidationError(code validation.ErrorCode, format string, a ...interface{}) *OCFLLoggerImpl {
	validationError := validation.GetValidationError(l.ver, code).AppendContext(format, a...)
	var event *zerolog.Event
	l.validationStatus.Add(validationError)

	if validationError.Code[0] == 'W' {
		event = l.Logger().Warn()
	} else {
		event = l.Logger().Error()
	}
	event.Str("OCFLVersion", l.ver.String()).
		Str("validationErrorCode", string(validationError.Code)).
		Str("validationErrorMessage", validationError.Description).
		Msgf(format, a...)
	return l
}

var _ OCFLLogger = (*OCFLLoggerImpl)(nil)
