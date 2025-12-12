package validation

import (
	"context"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type Validation interface {
	AddValidationError(errno ValidationErrorCode, format string, a ...any) error
	AddValidationWarning(errno ValidationErrorCode, format string, a ...any) error
}

func AddValidationError(ctx context.Context, ver version.OCFLVersion, errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(ver, errno).AppendDescription(format, a...)
	//	_, file, line, _ := runtime.Caller(1)
	return errors.WithStack(AddValidationErrors(ctx, valError))
}

func AddValidationWarning(ctx context.Context, ver version.OCFLVersion, errno ValidationErrorCode, format string, a ...any) error {
	valError := GetValidationError(ver, errno).AppendDescription(format, a...)
	//	_, file, line, _ := runtime.Caller(1)
	return errors.WithStack(AddValidationWarnings(ctx, valError))
}
