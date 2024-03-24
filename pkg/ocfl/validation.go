package ocfl

type Validation interface {
	addValidationError(errno ValidationErrorCode, format string, a ...any) error
	addValidationWarning(errno ValidationErrorCode, format string, a ...any) error
}
