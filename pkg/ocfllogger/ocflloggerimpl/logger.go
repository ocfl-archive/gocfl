// Package ocfllogger provides a logging interface for OCFL operations that integrates
// with the OCFL validation system. It wraps a standard logger and adds functionality
// to track validation errors and maintain OCFL-specific context.
package ocflloggerimpl

import (
	"context"
	"strings"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"golang.org/x/exp/maps"
)

// NewOCFLLogger creates a new OCFLLogger instance.
// If data is nil, an empty map is initialized.
// If validationStatus is nil, a new status object is created.
func NewOCFLLogger(ctx context.Context, logger zLogger.ZLogger, data map[string]string, ver version.OCFLVersion, validationStatus *validation.Status) ocfllogger.OCFLLogger {
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

// NewNopLogger creates a new OCFLLogger instance that does nothing.
func NewNopLogger() ocfllogger.OCFLLogger {
	return NewOCFLLogger(context.Background(), new(zerolog.New(zerolog.Nop())), nil, version.Default, nil)
}

// OCFLLoggerImpl is the concrete implementation of OCFLLogger.
type OCFLLoggerImpl struct {
	zLogger.ZLogger
	data             map[string]string
	ctx              context.Context
	ver              version.OCFLVersion
	validationStatus *validation.Status
}

// ValidationErrors returns all accumulated validation errors.
func (l *OCFLLoggerImpl) ValidationErrors() []*validation.Error {
	l.validationStatus.Compact()
	return l.validationStatus.Errors
}

// ClearValidationErrors clears all accumulated validation errors.
func (l *OCFLLoggerImpl) ClearValidationErrors() {
	l.validationStatus.Errors = make([]*validation.Error, 0)
}

// WithVersion returns a new logger with the given OCFL version.
func (l *OCFLLoggerImpl) WithVersion(ver version.OCFLVersion) ocfllogger.OCFLLogger {
	return NewOCFLLogger(l.ctx, l.ZLogger, l.data, ver, l.validationStatus)
}

// Logger returns the underlying zLogger.ZLogger.
func (l *OCFLLoggerImpl) Logger() zLogger.ZLogger {
	return l.ZLogger
}

// With returns a new logger with the given field added to the context.
func (l *OCFLLoggerImpl) With(name, value string) ocfllogger.OCFLLogger {
	newData := maps.Clone(l.data)
	newData[strings.ToLower(name)] = value
	return NewOCFLLogger(l.ctx, new(l.ZLogger.With().Str(name, value).Logger()), newData, l.ver, l.validationStatus)
}

// ValidationError logs an OCFL validation error and adds it to the internal validation status.
func (l *OCFLLoggerImpl) ValidationError(code validation.ErrorCode, format string, a ...any) ocfllogger.OCFLLogger {
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

var _ ocfllogger.OCFLLogger = (*OCFLLoggerImpl)(nil)
