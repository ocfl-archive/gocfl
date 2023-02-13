package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"strings"
)

const FlatOmitPrefixStorageLayoutName = "0006-flat-omit-prefix-storage-layout"
const FlatOmitPrefixStorageLayoutDescription = "removes prefix after last occurrence of delimiter"

type FlatOmitPrefixStorageLayoutConfig struct {
	*ocfl.ExtensionConfig
	Delimiter string `json:"delimiter"`
}
type FlatOmitPrefixStorageLayout struct {
	*FlatOmitPrefixStorageLayoutConfig
	fs ocfl.OCFLFSRead
}

func NewFlatOmitPrefixStorageLayoutFS(fsys ocfl.OCFLFSRead) (*FlatOmitPrefixStorageLayout, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &FlatOmitPrefixStorageLayoutConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewFlatOmitPrefixStorageLayout(config)
}
func NewFlatOmitPrefixStorageLayout(config *FlatOmitPrefixStorageLayoutConfig) (*FlatOmitPrefixStorageLayout, error) {
	sl := &FlatOmitPrefixStorageLayout{FlatOmitPrefixStorageLayoutConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (sl *FlatOmitPrefixStorageLayout) IsRegistered() bool {
	return true
}

func (sl *FlatOmitPrefixStorageLayout) Stat(w io.Writer, statInfo []ocfl.StatInfo) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.FlatOmitPrefixStorageLayoutConfig, "", "  ")
	return string(str)
}

func (sl *FlatOmitPrefixStorageLayout) SetFS(fs ocfl.OCFLFSRead) {
	sl.fs = fs
}

func (sl *FlatOmitPrefixStorageLayout) SetParams(params map[string]string) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetName() string { return FlatOmitPrefixStorageLayoutName }
func (sl *FlatOmitPrefixStorageLayout) WriteConfig() error {
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
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) WriteLayout(fs ocfl.OCFLFS) error {
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
		Extension:   FlatOmitPrefixStorageLayoutName,
		Description: FlatOmitPrefixStorageLayoutDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	last := strings.LastIndex(id, sl.Delimiter)
	if last < 0 {
		return id, nil
	}
	return id[last+len(sl.Delimiter):], nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &FlatOmitPrefixStorageLayout{}
	_ ocfl.ExtensionStorageRootPath = &FlatOmitPrefixStorageLayout{}
)
