package inventoryimpl

import (
	"fmt"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type ValidationError struct {
	Errno   validation.ValidationErrorCode
	Message string
}

func (e ValidationError) String() string {
	return fmt.Sprintf("%s: %s", e.Errno, e.Message)
}

func NewDummyValidation() *DummyValidation {
	return &DummyValidation{
		Error:   []ValidationError{},
		Warning: []ValidationError{},
	}
}

type DummyValidation struct {
	Error   []ValidationError
	Warning []ValidationError
}

func (d *DummyValidation) HasWarning(errno validation.ValidationErrorCode) bool {
	for _, e := range d.Warning {
		if e.Errno == errno {
			return true
		}
	}
	return false
}

func (d *DummyValidation) HasError(errno validation.ValidationErrorCode) bool {
	for _, e := range d.Error {
		if e.Errno == errno {
			return true
		}
	}
	return false
}

func (d *DummyValidation) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	d.Error = append(d.Error, ValidationError{
		Errno:   errno,
		Message: fmt.Sprintf(format, a...),
	})
	return nil
}

func (d *DummyValidation) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	d.Warning = append(d.Warning, ValidationError{
		Errno:   errno,
		Message: fmt.Sprintf(format, a...),
	})
	return nil
}

var _ validation.Validation = &DummyValidation{}
