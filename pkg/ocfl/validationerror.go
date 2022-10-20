package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"golang.org/x/exp/slices"
	"strings"
)

type ValidationErrorCode string

const (
	E000 = ValidationErrorCode("E000")
	E001 = ValidationErrorCode("E001")
	E002 = ValidationErrorCode("E002")
	E003 = ValidationErrorCode("E003")
	E004 = ValidationErrorCode("E004")
	E005 = ValidationErrorCode("E005")
	E006 = ValidationErrorCode("E006")
	E007 = ValidationErrorCode("E007")
	E008 = ValidationErrorCode("E008")
	E009 = ValidationErrorCode("E009")
	E010 = ValidationErrorCode("E010")
	E011 = ValidationErrorCode("E011")
	E012 = ValidationErrorCode("E012")
	E013 = ValidationErrorCode("E013")
	E014 = ValidationErrorCode("E014")
	E015 = ValidationErrorCode("E015")
	E016 = ValidationErrorCode("E016")
	E017 = ValidationErrorCode("E017")
	E018 = ValidationErrorCode("E018")
	E019 = ValidationErrorCode("E019")
	E020 = ValidationErrorCode("E020")
	E021 = ValidationErrorCode("E021")
	E022 = ValidationErrorCode("E022")
	E023 = ValidationErrorCode("E023")
	E024 = ValidationErrorCode("E024")
	E025 = ValidationErrorCode("E025")
	E026 = ValidationErrorCode("E026")
	E027 = ValidationErrorCode("E027")
	E028 = ValidationErrorCode("E028")
	E029 = ValidationErrorCode("E029")
	E030 = ValidationErrorCode("E030")
	E031 = ValidationErrorCode("E031")
	E032 = ValidationErrorCode("E032")
	E033 = ValidationErrorCode("E033")
	E034 = ValidationErrorCode("E034")
	E035 = ValidationErrorCode("E035")
	E036 = ValidationErrorCode("E036")
	E037 = ValidationErrorCode("E037")
	E038 = ValidationErrorCode("E038")
	E039 = ValidationErrorCode("E039")
	E040 = ValidationErrorCode("E040")
	E041 = ValidationErrorCode("E041")
	E042 = ValidationErrorCode("E042")
	E043 = ValidationErrorCode("E043")
	E044 = ValidationErrorCode("E044")
	E045 = ValidationErrorCode("E045")
	E046 = ValidationErrorCode("E046")
	E047 = ValidationErrorCode("E047")
	E048 = ValidationErrorCode("E048")
	E049 = ValidationErrorCode("E049")
	E050 = ValidationErrorCode("E050")
	E051 = ValidationErrorCode("E051")
	E052 = ValidationErrorCode("E052")
	E053 = ValidationErrorCode("E053")
	E054 = ValidationErrorCode("E054")
	E055 = ValidationErrorCode("E055")
	E056 = ValidationErrorCode("E056")
	E057 = ValidationErrorCode("E057")
	E058 = ValidationErrorCode("E058")
	E059 = ValidationErrorCode("E059")
	E060 = ValidationErrorCode("E060")
	E061 = ValidationErrorCode("E061")
	E062 = ValidationErrorCode("E062")
	E063 = ValidationErrorCode("E063")
	E064 = ValidationErrorCode("E064")
	E066 = ValidationErrorCode("E066")
	E067 = ValidationErrorCode("E067")
	E068 = ValidationErrorCode("E068")
	E069 = ValidationErrorCode("E069")
	E070 = ValidationErrorCode("E070")
	E071 = ValidationErrorCode("E071")
	E072 = ValidationErrorCode("E072")
	E073 = ValidationErrorCode("E073")
	E074 = ValidationErrorCode("E074")
	E075 = ValidationErrorCode("E075")
	E076 = ValidationErrorCode("E076")
	E077 = ValidationErrorCode("E077")
	E078 = ValidationErrorCode("E078")
	E079 = ValidationErrorCode("E079")
	E080 = ValidationErrorCode("E080")
	E081 = ValidationErrorCode("E081")
	E082 = ValidationErrorCode("E082")
	E083 = ValidationErrorCode("E083")
	E084 = ValidationErrorCode("E084")
	E085 = ValidationErrorCode("E085")
	E086 = ValidationErrorCode("E086")
	E087 = ValidationErrorCode("E087")
	E088 = ValidationErrorCode("E088")
	E089 = ValidationErrorCode("E089")
	E090 = ValidationErrorCode("E090")
	E091 = ValidationErrorCode("E091")
	E092 = ValidationErrorCode("E092")
	E093 = ValidationErrorCode("E093")
	E094 = ValidationErrorCode("E094")
	E095 = ValidationErrorCode("E095")
	E096 = ValidationErrorCode("E096")
	E097 = ValidationErrorCode("E097")
	E098 = ValidationErrorCode("E098")
	E099 = ValidationErrorCode("E099")
	E100 = ValidationErrorCode("E100")
	E101 = ValidationErrorCode("E101")
	E102 = ValidationErrorCode("E102")
	E103 = ValidationErrorCode("E103")
	E104 = ValidationErrorCode("E104")
	E105 = ValidationErrorCode("E105")
	E106 = ValidationErrorCode("E106")
	E107 = ValidationErrorCode("E107")
	E108 = ValidationErrorCode("E108")
	E110 = ValidationErrorCode("E110")
	E111 = ValidationErrorCode("E111")
	E112 = ValidationErrorCode("E112")
	W001 = ValidationErrorCode("W001")
	W002 = ValidationErrorCode("W002")
	W003 = ValidationErrorCode("W003")
	W004 = ValidationErrorCode("W004")
	W005 = ValidationErrorCode("W005")
	W007 = ValidationErrorCode("W007")
	W008 = ValidationErrorCode("W008")
	W009 = ValidationErrorCode("W009")
	W010 = ValidationErrorCode("W010")
	W011 = ValidationErrorCode("W011")
	W012 = ValidationErrorCode("W012")
	W013 = ValidationErrorCode("W013")
	W014 = ValidationErrorCode("W014")
	W015 = ValidationErrorCode("W015")
	W016 = ValidationErrorCode("W016")
)

type ValidationError struct {
	Code         ValidationErrorCode
	Description  string
	Ref          string
	Description2 string
}

type ValidationStatus struct {
	Errors, Warnings []*ValidationError
}

func (status ValidationStatus) Compact() {
	status.Errors = slices.CompactFunc(status.Errors, func(E1, E2 *ValidationError) bool {
		return E1.Code == E2.Code && E1.Description2 == E2.Description2
	})
	status.Warnings = slices.CompactFunc(status.Warnings, func(E1, E2 *ValidationError) bool {
		return E1.Code == E2.Code && E1.Description2 == E2.Description2
	})

}

func NewContextValidation(parent context.Context) context.Context {
	return context.WithValue(parent, "validationStatus", &ValidationStatus{
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
	})
}

func GetValidationStatus(ctx context.Context) (*ValidationStatus, error) {
	statusAny := ctx.Value("validationStatus")
	if statusAny == nil {
		return nil, errors.New("no Value validationStatus in context")
	}
	status, ok := statusAny.(*ValidationStatus)
	if !ok {
		return nil, errors.New("validationStatus not of type *ValidationStatus")
	}
	return status, nil
}

func addValidationErrors(ctx context.Context, vErrs ...*ValidationError) error {
	status, err := GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot add validation error")
	}
	status.Errors = append(status.Errors, vErrs...)
	return nil
}

func addValidationWarnings(ctx context.Context, vWarns ...*ValidationError) error {
	status, err := GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot add validation error")
	}
	status.Warnings = append(status.Warnings, vWarns...)
	return nil
}

func (ve *ValidationError) AppendDescription(format string, a ...any) *ValidationError {
	return &ValidationError{
		Code:         ve.Code,
		Description:  ve.Description,
		Ref:          ve.Ref,
		Description2: strings.TrimSpace(ve.Description2 + " " + fmt.Sprintf(format, a...)),
	}
}

func (verr *ValidationError) Error() string {
	return fmt.Sprintf("Validation Error #%s - %s (%s) [%s]", verr.Code, verr.Description, verr.Ref, verr.Description2)
}

func GetValidationError(version OCFLVersion, errno ValidationErrorCode) *ValidationError {
	var errlist map[ValidationErrorCode]*ValidationError
	var mapping map[ValidationErrorCode]ValidationErrorCode
	switch version {
	case "1.1":
		errlist = OCFLValidationError1_1
		mapping = OCFLValidationErrorMapping1_1
	default:
		//case "1.0":
		errlist = OCFLValidationError1_0
		mapping = OCFLValidationErrorMapping1_0
		//		errlist = map[ValidationErrorCode]*ValidationError{}
	}
	err, ok := errlist[errno]
	if !ok {
		errno, ok = mapping[errno]
		if ok {
			err, ok = errlist[errno]
			if ok {
				return err
			}
		}
		return &ValidationError{
			Code:        errno,
			Description: fmt.Sprintf("unknown error %s", errno),
			Ref:         "",
		}
	}
	return err
}
