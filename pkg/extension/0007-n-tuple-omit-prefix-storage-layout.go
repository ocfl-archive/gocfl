package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
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
	fs ocfl.OCFLFS
}

func NewNTupleOmitPrefixStorageLayoutFS(fsys ocfl.OCFLFSRead) (*NTupleOmitPrefixStorageLayout, error) {
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

func (sl *NTupleOmitPrefixStorageLayout) IsRegistered() bool {
	return true
}

func (sl *NTupleOmitPrefixStorageLayout) Stat(w io.Writer, statInfo []ocfl.StatInfo) error {
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.NTupleOmitPrefixStorageLayoutConfig, "", "  ")
	return string(str)
}

func (sl *NTupleOmitPrefixStorageLayout) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *NTupleOmitPrefixStorageLayout) SetParams(params map[string]string) error {
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) GetName() string { return NTupleOmitPrefixStorageLayoutName }
func (sl *NTupleOmitPrefixStorageLayout) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := sl.fs.Create("config.json")
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

func (sl *NTupleOmitPrefixStorageLayout) WriteLayout(fs ocfl.OCFLFS) error {
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
		Extension:   NTupleOmitPrefixStorageLayoutName,
		Description: NTupleOmitPrefixStorageLayoutDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *NTupleOmitPrefixStorageLayout) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	last := strings.LastIndex(id, sl.Delimiter)
	if last >= 0 {
		id = id[last+len(sl.Delimiter):]
	}
	var base string
	if sl.ReverseObjectRoot {
		base = reverse(id)
	} else {
		base = id
	}
	// todo: finalize
	var pathComponents = []string{}
	for i := 0; i*sl.TupleSize < len(base); i++ {

	}
	var _ = pathComponents
	return id[last+len(sl.Delimiter):], nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &NTupleOmitPrefixStorageLayout{}
	_ ocfl.ExtensionStorageRootPath = &NTupleOmitPrefixStorageLayout{}
)
