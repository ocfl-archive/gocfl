package ocfl

type Validation interface {
	addValidationError(errno ValidationErrorCode, format string, a ...any)
	addValidationWarning(errno ValidationErrorCode, format string, a ...any)
}
