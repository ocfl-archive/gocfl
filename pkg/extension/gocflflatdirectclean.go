package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const FlatDirectCleanName = "NNNN-direct-clean-path-layout"

var flatDirectCleanRuleAll = regexp.MustCompile("[\u0000-\u001f\u007f\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000\n\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var flatDirectCleanRuleWhitespace = regexp.MustCompile("[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]")
var flatDirectCleanRule_1_5 = regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var flatDirectCleanRule_2_4_6 = regexp.MustCompile("^[\\-~\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*(.*?)[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*$")
var flatDirectCleanRulePeriods = regexp.MustCompile("^\\.+$")

var flatDirectCleanErrFilenameTooLong = errors.New("filename too long")
var flatDirectCleanErrPathnameTooLong = errors.New("pathname too long")

type DirectClean struct {
	*DirectCleanConfig
}

type DirectCleanConfig struct {
	*Config
	MaxPathnameLen              int    `json:"maxPathnameLen,omitempty"`
	MaxFilenameLen              int    `json:"maxFilenameLen,omitempty"`
	ReplacementString           string `json:"replacementString,omitempty"`
	WhitespaceReplacementString string `json:"whitespaceReplacementString,omitempty"`
	UTFEncode                   bool   `json:"utfEncode,omitempty"`
}

func NewFlatDirectClean(config *DirectCleanConfig) (*DirectClean, error) {
	if config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = 32000
	}
	if config.MaxFilenameLen == 0 {
		config.MaxFilenameLen = 127
	}
	sl := &DirectClean{DirectCleanConfig: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}

func encodeUTFCode(s string) string {
	return "=u" + strings.Trim(fmt.Sprintf("%U", []rune(s)), "U+[]")
}

func (sl *DirectClean) ExecutePath(fname string) (string, error) {

	fname = strings.ToValidUTF8(fname, sl.ReplacementString)

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {

		if sl.UTFEncode {
			n = flatDirectCleanRuleAll.ReplaceAllStringFunc(n, encodeUTFCode)
		} else {
			n = flatDirectCleanRuleWhitespace.ReplaceAllString(n, sl.WhitespaceReplacementString)
			n = flatDirectCleanRule_1_5.ReplaceAllString(n, sl.ReplacementString)
		}
		n = flatDirectCleanRule_2_4_6.ReplaceAllString(n, "$1")
		if flatDirectCleanRulePeriods.MatchString(n) {
			if sl.UTFEncode {
				n = strings.Replace(n, ".", encodeUTFCode("."), -1)
			} else {
				n = strings.Replace(n, ".", sl.ReplacementString, -1)
			}
		}

		lenN := len(n)
		if lenN > sl.MaxFilenameLen {
			return "", errors.Wrapf(flatDirectCleanErrFilenameTooLong, "filename: %s", n)
		}

		if lenN > 0 {
			result = append(result, n)
		}
	}

	fname = strings.Join(result, "/")

	if len(fname) > sl.MaxPathnameLen {
		return "", errors.Wrapf(flatDirectCleanErrPathnameTooLong, "pathname: %s", fname)
	}

	return fname, nil
}

func (sl *DirectClean) Name() string { return FlatDirectCleanName }

func (sl *DirectClean) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
