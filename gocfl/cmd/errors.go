package cmd

import (
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

type errorID = archiveerror.ID

const (
	ERRORIDUnknownError      = "IDUnknownError"
	ErrorExtensionInit       = "ErrorExtensionInit"
	ErrorExtensionInitErr    = "ErrorExtensionInitErr"
	ErrorExtensionRunner     = "ErrorExtensionRunner"
	ErrorFS                  = "ErrorFS"
	ErrorGOCFL               = "ErrorGOCFL"
	ErrorOCFLCreation        = "ErrorOCFLCreation"
	ErrorOCFLEnd             = "ErrorOCFLEnd"
	ErrorValidationStatus    = "ErrorValidationStatus"
	ErrorValidationStatusErr = "ErrorValidationStatus"
	LogGOCFL                 = "LogGOCFL"
)
