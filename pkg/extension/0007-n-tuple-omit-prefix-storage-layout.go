package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"strings"
)

const NTupleOmitPrefixStorageLayoutName = "0007-n-tuple-omit-prefix-storage-layout"
const NTupleOmitPrefixStorageLayoutDescription = "pairtree-like root directory structure derived from prefix-omitted object identifiers"

// function, which takes a string as
// argument and return the reverse of string.
func reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {

		// swap the letters of the string,
		// like first with last and so on.
		rns[i], rns[j] = rns[j], rns[i]
	}

	// return the reversed string.
	return string(rns)
}

func NewNTupleOmitPrefixStorageLayoutFS(fsys fs.FS) (*NTupleOmitPrefixStorageLayout, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &NTupleOmitPrefixStorageLayoutConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewNTupleOmitPrefixStorageLayout(config)
}
func NewNTupleOmitPrefixStorageLayout(config *NTupleOmitPrefixStorageLayoutConfig) (*NTupleOmitPrefixStorageLayout, error) {
	sl := &NTupleOmitPrefixStorageLayout{NTupleOmitPrefixStorageLayoutConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type NTupleOmitPrefixStorageLayoutConfig struct {
	*ocfl.ExtensionConfig
	Delimiter         string `json:"delimiter"`
	TupleSize         int    `json:"tupleSize"`
	NumberOfTuples    int    `json:"numberOfTuples"`
	ZeroPadding       string `json:"zeroPadding"`
	ReverseObjectRoot bool   `json:"reverseObjectRoot"`
}

type NTupleOmitPrefixStorageLayout struct {
	*NTupleOmitPrefixStorageLayoutConfig
	fsys fs.FS
}

func (sl *NTupleOmitPrefixStorageLayout) Terminate() error {
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) GetFS() fs.FS {
	return sl.fsys
}

func (sl *NTupleOmitPrefixStorageLayout) GetConfig() any {
	return sl.NTupleOmitPrefixStorageLayoutConfig
}

func (sl *NTupleOmitPrefixStorageLayout) IsRegistered() bool {
	return true
}

func (sl *NTupleOmitPrefixStorageLayout) Stat(w io.Writer, statInfo []ocfl.StatInfo) error {
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *NTupleOmitPrefixStorageLayout) SetParams(params map[string]string) error {
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) GetName() string { return NTupleOmitPrefixStorageLayoutName }
func (sl *NTupleOmitPrefixStorageLayout) WriteConfig() error {
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
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) WriteLayout(fsys fs.FS) error {
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
		Extension:   NTupleOmitPrefixStorageLayoutName,
		Description: NTupleOmitPrefixStorageLayoutDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	/*
	  1) Remove the prefix, which is everything to the left of the right-most instance of the delimiter, as well as the delimiter. If there is no delimiter, the whole id is used; if the delimiter is found at the end, an error is thrown.
	*/
	if sl.Delimiter != "" {
		last := strings.LastIndex(id, sl.Delimiter)
		if last >= 0 {
			id = id[last+len(sl.Delimiter):]
		}
	}
	/*
	 2) Optionally, add zero-padding to the left or right of the remaining id, depending on zeroPadding configuration.
	*/
	var targetLength = sl.TupleSize * sl.NumberOfTuples
	/*
		if targetLength < len(id) {
			return "", errors.Errorf("'%s' longer than %v", id, targetLength)
		}
	*/
	var str = strings.Builder{}
	str.Grow(max(targetLength, len(id)))
	l := len(id)
	if sl.ZeroPadding == "right" {
		str.WriteString(id)
	}
	for i := 0; i < targetLength-l; i++ {
		str.WriteString("0")
	}
	if sl.ZeroPadding == "left" {
		str.WriteString(id)
	}
	base := str.String()
	/*
		3) Optionally reverse the remaining id, depending on reverseObjectRoot
	*/
	if sl.ReverseObjectRoot {
		base = reverse(base)
	}

	/*
		4) Starting at the leftmost character of the resulting id and working right, divide the id into numberOfTuples each containing tupleSize characters.
	*/
	var pathComponents = []string{}
	for i := 0; i < targetLength/sl.TupleSize; i++ {
		pathComponents = append(pathComponents, base[i*sl.TupleSize:(i+1)*sl.TupleSize])
	}
	/*
		5) Create the start of the object root path by joining the tuples, in order, using the filesystem path separator.
		6) Complete the object root path by joining the prefix-omitted id (from step 1) onto the end.
	*/
	pathComponents = append(pathComponents, id)
	result := strings.Join(pathComponents, "/")
	return result, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &NTupleOmitPrefixStorageLayout{}
	_ ocfl.ExtensionStorageRootPath = &NTupleOmitPrefixStorageLayout{}
)
