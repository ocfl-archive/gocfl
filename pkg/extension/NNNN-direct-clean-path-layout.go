package extension

import (
	"emperror.dev/errors"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/utils/v2/pkg/checksum"
	"golang.org/x/exp/constraints"
	"hash"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const DirectCleanName = "NNNN-direct-clean-path-layout"
const DirectCleanDescription = "Maps OCFL object identifiers to storage paths or as an object extension that maps logical paths to content paths. This is done by replacing or removing \"dangerous characters\" from names"

var directCleanRuleAll = regexp.MustCompile("[\u0000-\u001f\u007f\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000\n\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRuleWhitespace = regexp.MustCompile("[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]")
var directCleanRuleEqual = regexp.MustCompile("=(u[a-zA-Z0-9]{4})")
var directCleanRule_1_5 = regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRule_2_4_6 = regexp.MustCompile("^[\\-~\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]*(.*?)[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*$")
var directCleanRulePeriods = regexp.MustCompile("^\\.+$")

var directCleanErrFilenameTooLong = errors.New("filename too long")
var directCleanErrPathnameTooLong = errors.New("pathname too long")

type DirectClean struct {
	*DirectCleanConfig
	fs        ocfl.OCFLFSRead
	hash      hash.Hash  `json:"-"`
	hashMutex sync.Mutex `json:"-"`
}

type DirectCleanConfig struct {
	*ocfl.ExtensionConfig
	MaxPathnameLen              int                      `json:"maxPathnameLen"`
	MaxFilenameLen              int                      `json:"maxFilenameLen"`
	ReplacementString           string                   `json:"replacementString"`
	WhitespaceReplacementString string                   `json:"whitespaceReplacementString"`
	UTFEncode                   bool                     `json:"utfEncode"`
	FallbackDigestAlgorithm     checksum.DigestAlgorithm `json:"fallbackDigestAlgorithm"`
	FallbackFolder              string                   `json:"fallbackFolder"`
	FallbackSubFolders          int                      `json:"fallbackSubdirs"`
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func NewDirectCleanFS(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &DirectCleanConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewDirectClean(config)
}

func NewDirectClean(config *DirectCleanConfig) (ocfl.Extension, error) {
	if config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = 32000
	}
	if config.MaxFilenameLen == 0 {
		config.MaxFilenameLen = 127
	}
	if config.FallbackDigestAlgorithm == "" {
		config.FallbackDigestAlgorithm = checksum.DigestSHA512
	}
	if config.FallbackFolder == "" {
		config.FallbackFolder = "fallback"
	}

	sl := &DirectClean{DirectCleanConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.Errorf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName())
	}

	var err error
	if sl.hash, err = checksum.GetHash(config.FallbackDigestAlgorithm); err != nil {
		return nil, errors.Wrapf(err, "hash %s not supported", config.FallbackDigestAlgorithm)
	}

	return sl, nil
}

func encodeUTFCode(s string) string {
	return "=u" + strings.Trim(fmt.Sprintf("%U", []rune(s)), "U+[]")
}

// interface Extension

func (sl *DirectClean) IsRegistered() bool {
	return false
}

func (sl *DirectClean) GetName() string { return DirectCleanName }

func (sl *DirectClean) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.DirectCleanConfig, "", "  ")
	return string(str)
}

func (sl *DirectClean) SetFS(fs ocfl.OCFLFSRead) {
	sl.fs = fs
}

func (sl *DirectClean) SetParams(params map[string]string) error {
	return nil
}

func (sl *DirectClean) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	fsRW, ok := sl.fs.(ocfl.OCFLFS)
	if !ok {
		return errors.Errorf("filesystem is read only - '%s'", sl.fs.String())
	}

	configWriter, err := fsRW.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.DirectCleanConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *DirectClean) WriteLayout(fs ocfl.OCFLFS) error {
	configWriter, err := fs.Create("ocfl_layout.json")
	if err != nil {
		return errors.Wrap(err, "cannot open ocfl_layout.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(struct {
		Extension   string `json:"extension"`
		Description string `json:"description"`
	}{
		Extension:   PathDirectName,
		Description: DirectCleanDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// interface
func (sl *DirectClean) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	return sl.build(id)
}

func (sl *DirectClean) BuildObjectInternalPath(object ocfl.Object, originalPath string, area string) (string, error) {
	return sl.build(originalPath)
}

func (sl *DirectClean) fallback(fname string) (string, error) {
	// internal mutex for reusing the hash object
	sl.hashMutex.Lock()

	// reset hash function
	sl.hash.Reset()
	// add path
	if _, err := sl.hash.Write([]byte(fname)); err != nil {
		sl.hashMutex.Unlock()
		return "", errors.Wrapf(err, "cannot hash path '%s'", fname)
	}
	sl.hashMutex.Unlock()

	// get digest and encode it
	digestString := hex.EncodeToString(sl.hash.Sum(nil))

	// check whether digest fits in filename length
	parts := len(digestString) / sl.MaxFilenameLen
	rest := len(digestString) % sl.MaxFilenameLen
	if rest > 0 {
		parts++
	}
	// cut the digest if it's too long for filename length
	result := ""
	for i := 0; i < parts; i++ {
		result = filepath.Join(result, digestString[i*sl.MaxFilenameLen:min((i+1)*sl.MaxFilenameLen, len(digestString))])
	}

	// add all necessary subfolders
	for i := 0; i < sl.FallbackSubFolders; i++ {
		// paranoia, but safe
		result = filepath.Join(string(([]rune(digestString))[sl.FallbackSubFolders-i-1]), result)
	}
	/*
		result = filepath.Join(sl.FallbackFolder, result)
		result = filepath.Clean(result)
		result = filepath.ToSlash(result)
		result = strings.TrimLeft(result, "/")
	*/
	result = strings.TrimLeft(filepath.ToSlash(filepath.Clean(filepath.Join(sl.FallbackFolder, result))), "/")
	if len(result) > sl.MaxPathnameLen {
		return result, errors.Errorf("result has length of %d which is more than max allowed length of %d", len(result), sl.MaxPathnameLen)
	}
	return result, nil
}

func (sl *DirectClean) build(fname string) (string, error) {

	fname = strings.ToValidUTF8(fname, sl.ReplacementString)

	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		if len(n) == 0 {
			continue
		}
		if sl.UTFEncode {
			n = directCleanRuleEqual.ReplaceAllString(n, "=u003D$1")
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
			return sl.fallback(fname)
			//return "", errors.Wrapf(directCleanErrFilenameTooLong, "filename: %s", n)
		}

		if lenN > 0 {
			result = append(result, n)
		}
	}

	fname = strings.Join(result, "/")

	if len(fname) > sl.MaxPathnameLen {
		return sl.fallback(fname)
		//return "", errors.Wrapf(directCleanErrPathnameTooLong, "pathname: %s", fname)
	}

	return fname, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                  = &DirectClean{}
	_ ocfl.ExtensionStorageRootPath   = &DirectClean{}
	_ ocfl.ExtensionObjectContentPath = &DirectClean{}
)
