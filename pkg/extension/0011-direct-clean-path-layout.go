package extension

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"golang.org/x/exp/constraints"
)

const DirectCleanName = "0011-direct-clean-path-layout"
const DirectCleanDescription = "Maps OCFL object identifiers to storage paths or as an object extension that maps logical paths to content paths. This is done by replacing or removing \"dangerous characters\" from names"

var directCleanRuleAll = regexp.MustCompile("[\u0000-\u001f\u007f\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000\n\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRuleWhitespace = regexp.MustCompile("[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]")
var directCleanRuleEqual = regexp.MustCompile("=(u[a-zA-Z0-9]{4})")
var directCleanRule_1_5 = regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRule_2_4_6 = regexp.MustCompile("^[\\-~\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]*(.*?)[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*$")
var directCleanRulePeriods = regexp.MustCompile("^\\.+$")

var directCleanErrFilenameTooLong = errors.New("filename too long")
var directCleanErrPathnameTooLong = errors.New("pathname too long")

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

func NewDirectCleanFS(fsys fs.FS) (extension.Extension, error) {
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
	// compatibility with old config
	if config.MaxFilenameLen > 0 && config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = config.MaxFilenameLen
		config.MaxFilenameLen = 0
	}
	if config.FallbackSubFolders > 0 && config.NumberOfFallbackTuples == 0 {
		config.NumberOfFallbackTuples = config.FallbackSubFolders
		config.FallbackSubFolders = 0
	}
	return NewDirectClean(config)
}

func NewDirectClean(config *DirectCleanConfig) (extension.Extension, error) {
	if config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = 32000
	}
	if config.MaxPathSegmentLen == 0 {
		config.MaxPathSegmentLen = 127
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

type DirectCleanConfig struct {
	*extension.ExtensionConfig
	MaxPathnameLen              int                      `json:"maxPathnameLen"`
	MaxPathSegmentLen           int                      `json:"maxPathSegmentLen"`
	ReplacementString           string                   `json:"replacementString"`
	WhitespaceReplacementString string                   `json:"whitespaceReplacementString"`
	UTFEncode                   bool                     `json:"utfEncode"`
	FallbackDigestAlgorithm     checksum.DigestAlgorithm `json:"fallbackDigestAlgorithm"`
	FallbackFolder              string                   `json:"fallbackFolder"`
	NumberOfFallbackTuples      int                      `json:"numberOfFallbackTuples"`
	FallbackTupleSize           int                      `json:"fallbackTupleSize"`

	// compatibility with old config
	MaxFilenameLen     int `json:"maxFilenameLen,omitempty"`
	FallbackSubFolders int `json:"fallbackSubdirs,omitempty"`
}

type DirectClean struct {
	*DirectCleanConfig
	fsys      fs.FS
	hash      hash.Hash  `json:"-"`
	hashMutex sync.Mutex `json:"-"`
}

func (sl *DirectClean) Terminate() error {
	return nil
}

func (sl *DirectClean) GetFS() fs.FS {
	return sl.fsys
}

func (sl *DirectClean) GetConfig() any {
	return sl.DirectCleanConfig
}

// interface Extension

func (sl *DirectClean) IsRegistered() bool {
	return true
}

func (sl *DirectClean) GetName() string { return DirectCleanName }

func (sl *DirectClean) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *DirectClean) SetParams(params map[string]string) error {
	return nil
}

func (sl *DirectClean) WriteConfig() error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
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

func (sl *DirectClean) WriteLayout(fsys fs.FS) error {
	configWriter, err := writefs.Create(fsys, "ocfl_layout.json")
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
func (sl *DirectClean) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
	return sl.build(id)
}

func (sl *DirectClean) BuildObjectManifestPath(object object.Object, originalPath string, area string) (string, error) {
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
	parts := len(digestString) / sl.MaxPathSegmentLen
	rest := len(digestString) % sl.MaxPathSegmentLen
	if rest > 0 {
		parts++
	}
	// cut the digest if it's too long for filename length
	result := ""
	for i := 0; i < parts; i++ {
		result = filepath.Join(result, digestString[i*sl.MaxPathSegmentLen:min((i+1)*sl.MaxPathSegmentLen, len(digestString))])
	}

	// add all necessary subfolders
	for i := 0; i < sl.NumberOfFallbackTuples; i++ {
		// paranoia, but safe
		result = filepath.Join(string(([]rune(digestString))[sl.NumberOfFallbackTuples-i-1:sl.NumberOfFallbackTuples-i-1+sl.FallbackTupleSize]), result)
	}
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
		if lenN > sl.MaxPathSegmentLen {
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
	_ extension.Extension                  = &DirectClean{}
	_ storageroot.ExtensionStorageRootPath = &DirectClean{}
	_ object.ExtensionObjectContentPath    = &DirectClean{}
)
