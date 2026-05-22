package ocfllogger

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

// OCFLLogger is the interface for OCFL-aware logging.
type OCFLLogger interface {
	// Logger returns the underlying zLogger.ZLogger.
	Logger() zLogger.ZLogger
	// Trace returns a zerolog.Event for trace level logging.
	Trace() *zerolog.Event
	// Debug returns a zerolog.Event for debug level logging.
	Debug() *zerolog.Event
	// Info returns a zerolog.Event for info level logging.
	Info() *zerolog.Event
	// Warn returns a zerolog.Event for warning level logging.
	Warn() *zerolog.Event
	// Error returns a zerolog.Event for error level logging.
	Error() *zerolog.Event
	// Err returns a zerolog.Event for the given error.
	Err(err error) *zerolog.Event
	// Fatal returns a zerolog.Event for fatal level logging.
	Fatal() *zerolog.Event
	// Panic returns a zerolog.Event for panic level logging.
	Panic() *zerolog.Event
	// With returns a new logger with the given field added to the context.
	With(name, value string) OCFLLogger
	// WithVersion returns a new logger with the given OCFL version.
	WithVersion(ver version.OCFLVersion) OCFLLogger
	// ValidationError logs an OCFL validation error and adds it to the internal validation status.
	// It uses the logger's current OCFL version to look up the error description.
	ValidationError(code validation.ErrorCode, format string, a ...any) OCFLLogger
	// ValidationErrors returns all accumulated validation errors.
	ValidationErrors() []*validation.Error
	// ClearValidationErrors clears all accumulated validation errors.
	ClearValidationErrors()
}
