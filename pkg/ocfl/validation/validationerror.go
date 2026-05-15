// Package validation provides structures and functions for handling OCFL validation errors and warnings.
//
// It maps OCFL specification error codes to descriptive error objects and provides
// a mechanism to collect these errors during the validation process.
// Base error definitions are loaded from validationerror1_0.go and validationerror1_1.go.
package validation

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// ErrorCode represents an OCFL validation error or warning code (e.g., "E001", "W001").
type ErrorCode string

const (
	E000 = ErrorCode("E000") // Unknown error
	E001 = ErrorCode("E001")
	E002 = ErrorCode("E002")
	E003 = ErrorCode("E003")
	E004 = ErrorCode("E004")
	E005 = ErrorCode("E005")
	E006 = ErrorCode("E006")
	E007 = ErrorCode("E007")
	E008 = ErrorCode("E008")
	E009 = ErrorCode("E009")
	E010 = ErrorCode("E010")
	E011 = ErrorCode("E011")
	E012 = ErrorCode("E012")
	E013 = ErrorCode("E013")
	E014 = ErrorCode("E014")
	E015 = ErrorCode("E015")
	E016 = ErrorCode("E016")
	E017 = ErrorCode("E017")
	E018 = ErrorCode("E018")
	E019 = ErrorCode("E019")
	E020 = ErrorCode("E020")
	E021 = ErrorCode("E021")
	E022 = ErrorCode("E022")
	E023 = ErrorCode("E023")
	E024 = ErrorCode("E024")
	E025 = ErrorCode("E025")
	E026 = ErrorCode("E026")
	E027 = ErrorCode("E027")
	E028 = ErrorCode("E028")
	E029 = ErrorCode("E029")
	E030 = ErrorCode("E030")
	E031 = ErrorCode("E031")
	E032 = ErrorCode("E032")
	E033 = ErrorCode("E033")
	E034 = ErrorCode("E034")
	E035 = ErrorCode("E035")
	E036 = ErrorCode("E036")
	E037 = ErrorCode("E037")
	E038 = ErrorCode("E038")
	E039 = ErrorCode("E039")
	E040 = ErrorCode("E040")
	E041 = ErrorCode("E041")
	E042 = ErrorCode("E042")
	E043 = ErrorCode("E043")
	E044 = ErrorCode("E044")
	E045 = ErrorCode("E045")
	E046 = ErrorCode("E046")
	E047 = ErrorCode("E047")
	E048 = ErrorCode("E048")
	E049 = ErrorCode("E049")
	E050 = ErrorCode("E050")
	E051 = ErrorCode("E051")
	E052 = ErrorCode("E052")
	E053 = ErrorCode("E053")
	E054 = ErrorCode("E054")
	E055 = ErrorCode("E055")
	E056 = ErrorCode("E056")
	E057 = ErrorCode("E057")
	E058 = ErrorCode("E058")
	E059 = ErrorCode("E059")
	E060 = ErrorCode("E060")
	E061 = ErrorCode("E061")
	E062 = ErrorCode("E062")
	E063 = ErrorCode("E063")
	E064 = ErrorCode("E064")
	E066 = ErrorCode("E066")
	E067 = ErrorCode("E067")
	E068 = ErrorCode("E068")
	E069 = ErrorCode("E069")
	E070 = ErrorCode("E070")
	E071 = ErrorCode("E071")
	E072 = ErrorCode("E072")
	E073 = ErrorCode("E073")
	E074 = ErrorCode("E074")
	E075 = ErrorCode("E075")
	E076 = ErrorCode("E076")
	E077 = ErrorCode("E077")
	E078 = ErrorCode("E078")
	E079 = ErrorCode("E079")
	E080 = ErrorCode("E080")
	E081 = ErrorCode("E081")
	E082 = ErrorCode("E082")
	E083 = ErrorCode("E083")
	E084 = ErrorCode("E084")
	E085 = ErrorCode("E085")
	E086 = ErrorCode("E086")
	E087 = ErrorCode("E087")
	E088 = ErrorCode("E088")
	E089 = ErrorCode("E089")
	E090 = ErrorCode("E090")
	E091 = ErrorCode("E091")
	E092 = ErrorCode("E092")
	E093 = ErrorCode("E093")
	E094 = ErrorCode("E094")
	E095 = ErrorCode("E095")
	E096 = ErrorCode("E096")
	E097 = ErrorCode("E097")
	E098 = ErrorCode("E098")
	E099 = ErrorCode("E099")
	E100 = ErrorCode("E100")
	E101 = ErrorCode("E101")
	E102 = ErrorCode("E102")
	E103 = ErrorCode("E103")
	E104 = ErrorCode("E104")
	E105 = ErrorCode("E105")
	E106 = ErrorCode("E106")
	E107 = ErrorCode("E107")
	E108 = ErrorCode("E108")
	E110 = ErrorCode("E110")
	E111 = ErrorCode("E111")
	E112 = ErrorCode("E112")
	W000 = ErrorCode("W000")
	W001 = ErrorCode("W001")
	W002 = ErrorCode("W002")
	W003 = ErrorCode("W003")
	W004 = ErrorCode("W004")
	W005 = ErrorCode("W005")
	W007 = ErrorCode("W007")
	W008 = ErrorCode("W008")
	W009 = ErrorCode("W009")
	W010 = ErrorCode("W010")
	W011 = ErrorCode("W011")
	W012 = ErrorCode("W012")
	W013 = ErrorCode("W013")
	W014 = ErrorCode("W014")
	W015 = ErrorCode("W015")
	W016 = ErrorCode("W016")
)

// Error represents a detailed OCFL validation issue.
type Error struct {
	Code         ErrorCode           // The OCFL validation code (e.g., E001).
	Description  string              // Official description from the OCFL specification.
	Ref          string              // URL to the relevant section in the OCFL specification.
	Description2 string              // Additional dynamic information about the specific occurrence.
	Context      string              // Architectural context (e.g., "storage root", "object root").
	Version      version.OCFLVersion // OCFL version the error refers to.
}

// NewStatus creates a new empty Status object.
func NewStatus() *Status {
	return &Status{
		Errors: []*Error{},
	}
}

// Status collects validation errors and warnings encountered during validation.
type Status struct {
	Errors []*Error // List of collected errors and warnings.
}

// Compact removes duplicate errors from the Status and sorts them.
func (status *Status) Compact() {
	slices.SortFunc(status.Errors, validationSort)
	status.Errors = slices.CompactFunc(status.Errors, func(E1, E2 *Error) bool {
		return E1.Context == E2.Context && E1.Code == E2.Code && E1.Description2 == E2.Description2
	})
}

// Add appends a validation error to the Status.
func (status *Status) Add(validationError *Error) {
	status.Errors = append(status.Errors, validationError)
}

// validationSort is a helper to sort Errors by context and code.
func validationSort(E1, E2 *Error) int {
	sr1 := strings.HasPrefix(E1.Context, "storage root")
	sr2 := strings.HasPrefix(E2.Context, "storage root")
	if sr1 != sr2 {
		if sr1 {
			return -1
		}
		return 1
	}
	return cmp.Compare(E1.Context+string(E1.Code)+E1.Description2, E2.Context+string(E2.Code)+E2.Description2)
}

// AppendDescription creates a new Error with additional information appended to Description2.
func (ve *Error) AppendDescription(format string, a ...any) *Error {
	if format == "" {
		return ve
	}
	return &Error{
		Code:         ve.Code,
		Description:  ve.Description,
		Ref:          ve.Ref,
		Description2: strings.TrimSpace(ve.Description2 + " " + fmt.Sprintf(format, a...)),
		Context:      ve.Context,
		Version:      ve.Version,
	}
}

// AppendContext creates a new Error with additional information appended to the Context.
func (ve *Error) AppendContext(format string, a ...any) *Error {
	if format == "" {
		return ve
	}
	return &Error{
		Code:         ve.Code,
		Description:  ve.Description,
		Ref:          ve.Ref,
		Description2: ve.Description2,
		Context:      strings.TrimSpace(ve.Context + " " + fmt.Sprintf(format, a...)),
		Version:      ve.Version,
	}
}

// Error implements the error interface for the Error struct.
func (verr *Error) Error() string {
	if len(verr.Code) > 0 && verr.Code[0] == 'W' {
		return fmt.Sprintf("Warning #%s [%s]  %s (%s) [%s]", verr.Code, verr.Context, verr.Description, verr.Ref, verr.Description2)
	} else {
		return fmt.Sprintf("Error #%s [%s] %s (%s) [%s]", verr.Code, verr.Context, verr.Description, verr.Ref, verr.Description2)
	}
}

// DetailString returns a string representation of the error including OCFL version prefix.
func (ve *Error) DetailString() string {
	switch ve.Version {
	case "1.1":
		return fmt.Sprintf("ocfl11.%s", ve.Code)
	default:
		return fmt.Sprintf("ocfl10.%s", ve.Code)
	}
}

// GetValidationError returns a base Error object for a given OCFL version and error code.
// If the code is not found for the version, it attempts to map it or returns an "unknown" error/warning.
func GetValidationError(version version.OCFLVersion, errno ErrorCode) *Error {
	var errlist map[ErrorCode]*Error
	var mapping map[ErrorCode]ErrorCode
	switch version {
	default:
		errlist = OCFLValidationError1_1
		mapping = OCFLValidationErrorMapping1_1
	case "1.0":
		//case "1.0":
		errlist = OCFLValidationError1_0
		mapping = OCFLValidationErrorMapping1_0
		//		errlist = map[ValidationErrorCode]*ValidationError{}
	}
	err, ok := errlist[errno]
	if !ok {
		errnomap, ok := mapping[errno]
		if !ok {
			if len(errno) > 0 && errno[0] == 'W' {
				return &Error{
					Code:        W000,
					Description: fmt.Sprintf("unknown warning %s", errno),
					Ref:         "",
				}
			}

			return &Error{
				Code:        E000,
				Description: fmt.Sprintf("unknown error %s", errno),
				Ref:         "",
			}
		}
		errno = errnomap
		err, ok = errlist[errno]
		if ok {
			return err
		}
	}
	return err
}
