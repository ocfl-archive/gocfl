package extension

import (
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/stat"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"

	"io"
	"io/fs"
	"strings"
)

const FlatOmitPrefixStorageLayoutName = "0006-flat-omit-prefix-storage-layout"
const FlatOmitPrefixStorageLayoutDescription = "removes prefix after last occurrence of delimiter"

func NewFlatOmitPrefixStorageLayoutFS(fsys fs.FS) (*FlatOmitPrefixStorageLayout, error) {
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

type FlatOmitPrefixStorageLayoutConfig struct {
	*extension.ExtensionConfig
	Delimiter string `json:"delimiter"`
}
type FlatOmitPrefixStorageLayout struct {
	*FlatOmitPrefixStorageLayoutConfig
	fsys fs.FS
}

func (sl *FlatOmitPrefixStorageLayout) Terminate() error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetFS() fs.FS {
	return sl.fsys
}

func (sl *FlatOmitPrefixStorageLayout) GetConfig() any {
	return sl.FlatOmitPrefixStorageLayoutConfig
}

func (sl *FlatOmitPrefixStorageLayout) IsRegistered() bool {
	return true
}

func (sl *FlatOmitPrefixStorageLayout) Stat(w io.Writer, statInfo []stat.StatInfo) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *FlatOmitPrefixStorageLayout) SetParams(params map[string]string) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetName() string { return FlatOmitPrefixStorageLayoutName }
func (sl *FlatOmitPrefixStorageLayout) WriteConfig() error {
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

func (sl *FlatOmitPrefixStorageLayout) WriteLayout(fsys fs.FS) error {
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
		Extension:   FlatOmitPrefixStorageLayoutName,
		Description: FlatOmitPrefixStorageLayoutDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
	last := strings.LastIndex(id, sl.Delimiter)
	if last < 0 {
		return id, nil
	}
	return id[last+len(sl.Delimiter):], nil
}

// check interface satisfaction
var (
	_ extension.Extension                  = &FlatOmitPrefixStorageLayout{}
	_ storageroot.ExtensionStorageRootPath = &FlatOmitPrefixStorageLayout{}
)
