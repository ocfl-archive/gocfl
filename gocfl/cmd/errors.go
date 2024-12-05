package cmd

import (
	"fmt"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

type errorID = archiveerror.ID

const (
	ERRORIDUnknownError = "IDUnknownError"
	ErrorReplaceMe      = "ErrorReplaceMe"
)

// factoryError provides a helper to return a factory error across the
// cmd package. The function ensures that the correct process slice is
// retrieved and returned to the caller to aid in debugging.
func factoryError(lookup errorID, detail string, err error, module string) (string, error) {
	factoryErr := ErrorFactory.NewError(lookup, detail, nil)
	errWithAdditional := factoryErr.WithAdditional(module, 2, nil)
	return fmt.Sprintf("%s", factoryErr.Type), errWithAdditional
}
