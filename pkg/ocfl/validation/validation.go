package validation

type Validation interface {
	AddValidationError(errno ValidationErrorCode, format string, a ...any) error
	AddValidationWarning(errno ValidationErrorCode, format string, a ...any) error
}
