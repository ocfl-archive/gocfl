package validation

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"golang.org/x/exp/slices"
)

type ErrorCode string

const (
	E000 = ErrorCode("E000")
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

type Error struct {
	Code         ErrorCode
	Description  string
	Ref          string
	Description2 string
	Context      string
	Version      version.OCFLVersion
}

type Status struct {
	Errors []*Error
}

func validationSort(E1, E2 *Error) int {
	sr1 := strings.HasPrefix(E1.Context, "storage root")
	sr2 := strings.HasPrefix(E2.Context, "storage root")
	if sr1 != sr2 {
		return -1 // sr1 && !sr2
	}
	return cmp.Compare(E1.Context+string(E1.Code)+E1.Description2, E2.Context+string(E2.Code)+E2.Description2)
}

// removes duplicate errors
func (status *Status) Compact() {
	slices.SortFunc(status.Errors, validationSort)
	status.Errors = slices.CompactFunc(status.Errors, func(E1, E2 *Error) bool {
		return E1.Context == E2.Context && E1.Code == E2.Code && E1.Description2 == E2.Description2
	})
	/*
		slices.SortFunc(status.Warnings, func(E1, E2 *ValidationError) bool {
			return E1.Context+string(E1.Code)+E1.Description2 < E2.Context+string(E2.Code)+E2.Description2
		})
		status.Warnings = slices.CompactFunc(status.Warnings, func(E1, E2 *ValidationError) bool {
			return E1.Context == E2.Context && E1.Code == E2.Code && E1.Description2 == E2.Description2
		})
	*/

}

func NewContextValidation(parent context.Context) context.Context {
	return context.WithValue(parent, "validationStatus", &Status{
		Errors: []*Error{},
		//		Warnings: []*ValidationError{},
	})
}

func GetValidationStatus(ctx context.Context) (*Status, error) {
	statusAny := ctx.Value("validationStatus")
	if statusAny == nil {
		return nil, errors.New("no Value validationStatus in context")
	}
	status, ok := statusAny.(*Status)
	if !ok {
		return nil, errors.New("validationStatus not of type *ValidationStatus")
	}
	return status, nil
}

func AddValidationErrors(ctx context.Context, vErrs ...*Error) error {
	status, err := GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot add validation error")
	}
	status.Errors = append(status.Errors, vErrs...)
	return nil
}

func AddValidationWarnings(ctx context.Context, vWarns ...*Error) error {
	status, err := GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot add validation error")
	}
	//	status.Warnings = append(status.Warnings, vWarns...)
	status.Errors = append(status.Errors, vWarns...)
	return nil
}

func (ve *Error) AppendDescription(format string, a ...any) *Error {
	if format == "" {
		return ve
	}
	return &Error{
		Code:         ve.Code,
		Description:  ve.Description,
		Ref:          ve.Ref,
		Description2: strings.TrimSpace(ve.Description2 + " " + fmt.Sprintf(format, a...)),
	}
}

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
	}
}

func (verr *Error) Error() string {
	if len(verr.Code) > 0 && verr.Code[0] == 'W' {
		return fmt.Sprintf("[%s] Validation Warning #%s - %s (%s) [%s]", verr.Context, verr.Code, verr.Description, verr.Ref, verr.Description2)
	} else {
		return fmt.Sprintf("[%s] Validation Error #%s - %s (%s) [%s]", verr.Context, verr.Code, verr.Description, verr.Ref, verr.Description2)
	}
}

func (ve *Error) DetailString() string {
	switch ve.Version {
	case "1.1":
		return fmt.Sprintf("ocfl11.%s", ve.Code)
	default:
		return fmt.Sprintf("ocfl10.%s", ve.Code)
	}
}

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
