package extension

import (
	"encoding/json"

	"io"
	"io/fs"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const FlatOmitPrefixStorageLayoutName = "0006-flat-omit-prefix-storage-layout"
const FlatOmitPrefixStorageLayoutDescription = "removes prefix after last occurrence of delimiter"

func init() {
	extension.RegisterExtension(FlatOmitPrefixStorageLayoutName, NewFlatOmitPrefixStorageLayout, nil)
}

func NewFlatOmitPrefixStorageLayout() (extensiontypes.Extension, error) {
	config := &FlatOmitPrefixStorageLayoutConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: FlatOmitPrefixStorageLayoutName},
		Delimiter:       ":",
	}
	sl := &FlatOmitPrefixStorageLayout{FlatOmitPrefixStorageLayoutConfig: config}
	return sl, nil
}

func (sl *FlatOmitPrefixStorageLayout) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", FlatOmitPrefixStorageLayoutName)
	return sl
}

func (sl *FlatOmitPrefixStorageLayout) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.FlatOmitPrefixStorageLayoutConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal FlatOmitPrefixStorageLayoutConfig '%s'", string(data))
	}
	if sl.Delimiter == "" {
		sl.Delimiter = ":"
	}
	return nil
}

type FlatOmitPrefixStorageLayoutConfig struct {
	*extensiontypes.ExtensionConfig
	Delimiter string `json:"delimiter"`
}
type FlatOmitPrefixStorageLayout struct {
	*FlatOmitPrefixStorageLayoutConfig
	logger ocfllogger.OCFLLogger
}

func (sl *FlatOmitPrefixStorageLayout) Terminate() error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetConfig() any {
	return sl.FlatOmitPrefixStorageLayoutConfig
}

func (sl *FlatOmitPrefixStorageLayout) IsRegistered() bool {
	return true
}

func (sl *FlatOmitPrefixStorageLayout) Stat(w io.Writer, statInfo []object.StatInfo) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) SetParams(params map[string]string) error {
	return nil
}

func (sl *FlatOmitPrefixStorageLayout) GetName() string { return FlatOmitPrefixStorageLayoutName }
func (sl *FlatOmitPrefixStorageLayout) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
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

func (sl *FlatOmitPrefixStorageLayout) WriteLayout(fsys appendfs.FS) error {
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
	_ extensiontypes.Extension             = &FlatOmitPrefixStorageLayout{}
	_ storageroot.ExtensionStorageRootPath = &FlatOmitPrefixStorageLayout{}
)
