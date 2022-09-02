package storageroot

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const StorageLayoutDirectCleanName = "NNNN-flat-direct-clean-storage-layout"

var directCleanRuleAll = regexp.MustCompile("[\u0000-\u001f\u007f\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000\n\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRuleWhitespace = regexp.MustCompile("[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]")
var directCleanRule_1_5 = regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRule_2_4_6 = regexp.MustCompile("^[\\-~\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*(.*?)[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*$")
var directCleanRulePeriods = regexp.MustCompile("^\\.+$")

var directCleanErrFilenameTooLong = errors.New("filename too long")
var directCleanErrPathnameTooLong = errors.New("pathname too long")

type StorageLayoutDirectClean struct {
	*StorageLayoutDirectCleanConfig
}

type StorageLayoutDirectCleanConfig struct {
	*Config
	MaxPathnameLen              int    `json:"maxPathnameLen,omitempty"`
	MaxFilenameLen              int    `json:"maxFilenameLen,omitempty"`
	ReplacementString           string `json:"replacementString,omitempty"`
	WhitespaceReplacementString string `json:"whitespaceReplacementString,omitempty"`
	UTFEncode                   bool   `json:"utfEncode,omitempty"`
}

func NewStorageLayoutDirectClean(config *StorageLayoutDirectCleanConfig) (*StorageLayoutDirectClean, error) {
	if config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = 32000
	}
	if config.MaxFilenameLen == 0 {
		config.MaxFilenameLen = 127
	}
	sl := &StorageLayoutDirectClean{StorageLayoutDirectCleanConfig: config}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}
	return sl, nil
}

func encodeUTFCode(s string) string {
	return "=u" + strings.Trim(fmt.Sprintf("%U", []rune(s)), "U+[]")
}

func (sl *StorageLayoutDirectClean) ExecuteID(fname string) (string, error) {

	fname = strings.ToValidUTF8(fname, sl.ReplacementString)

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		if len(n) == 0 {
			continue
		}
		if sl.UTFEncode {
			n = directCleanRuleAll.ReplaceAllStringFunc(n, encodeUTFCode)
			if n[0] == '~' || directCleanRulePeriods.MatchString(n) {
				n = encodeUTFCode(string(n[0])) + n[1:]
			}
		} else {
			n = directCleanRuleWhitespace.ReplaceAllString(n, sl.WhitespaceReplacementString)
			n = directCleanRule_1_5.ReplaceAllString(n, sl.ReplacementString)
			n = directCleanRule_2_4_6.ReplaceAllString(n, "$1")
			if directCleanRulePeriods.MatchString(n) {
				n = sl.ReplacementString + n[1:]
			}
		}

		lenN := len(n)
		if lenN > sl.MaxFilenameLen {
			return "", errors.Wrapf(directCleanErrFilenameTooLong, "filename: %s", n)
		}

		if lenN > 0 {
			result = append(result, n)
		}
	}

	fname = strings.Join(result, "/")

	if len(fname) > sl.MaxPathnameLen {
		return "", errors.Wrapf(directCleanErrPathnameTooLong, "pathname: %s", fname)
	}

	return fname, nil
}

func (sl *StorageLayoutDirectClean) Name() string { return StorageLayoutDirectCleanName }

func (sl *StorageLayoutDirectClean) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.StorageLayoutDirectCleanConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
