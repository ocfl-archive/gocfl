package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
)

const StorageLayoutFlatDirectName = "0002-flat-direct-storage-layout"
const StorageLayoutFlatDirectDescription = "one to one mapping without changes"

type StorageLayoutFlatDirectConfig struct {
	*ocfl.ExtensionConfig
}
type StorageLayoutFlatDirect struct {
	*StorageLayoutFlatDirectConfig
	fs ocfl.OCFLFS
}

func NewStorageLayoutFlatDirectFS(fsys ocfl.OCFLFSRead) (*StorageLayoutFlatDirect, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &StorageLayoutFlatDirectConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewStorageLayoutFlatDirect(config)
}
func NewStorageLayoutFlatDirect(config *StorageLayoutFlatDirectConfig) (*StorageLayoutFlatDirect, error) {
	sl := &StorageLayoutFlatDirect{StorageLayoutFlatDirectConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (sl *StorageLayoutFlatDirect) IsRegistered() bool {
	return true
}

func (sl *StorageLayoutFlatDirect) Stat(w io.Writer, statInfo []ocfl.StatInfo) error {
	return nil
}

func (sl *StorageLayoutFlatDirect) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.StorageLayoutFlatDirectConfig, "", "  ")
	return string(str)
}

func (sl *StorageLayoutFlatDirect) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *StorageLayoutFlatDirect) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutFlatDirect) GetName() string { return StorageLayoutFlatDirectName }
func (sl *StorageLayoutFlatDirect) WriteConfig() error {
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

func (sl *StorageLayoutFlatDirect) WriteLayout(fs ocfl.OCFLFS) error {
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
		Extension:   StorageLayoutFlatDirectName,
		Description: StorageLayoutFlatDirectDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *StorageLayoutFlatDirect) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	return id, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &StorageLayoutFlatDirect{}
	_ ocfl.ExtensionStorageRootPath = &StorageLayoutFlatDirect{}
)
