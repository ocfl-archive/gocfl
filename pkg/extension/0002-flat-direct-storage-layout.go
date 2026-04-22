package extension

import (
	"encoding/json"
	"io"

	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const StorageLayoutFlatDirectName = "0002-flat-direct-storage-layout"
const StorageLayoutFlatDirectDescription = "one to one mapping without changes"

func init() {
	extension.RegisterExtension(StorageLayoutFlatDirectName, NewStorageLayoutFlatDirect, nil)
}

func NewStorageLayoutFlatDirect() (extension.Extension, error) {
	var config = &StorageLayoutFlatDirectConfig{
		ExtensionConfig: &extension.ExtensionConfig{
			ExtensionName: StorageLayoutFlatDirectName,
		},
	}
	sl := &StorageLayoutFlatDirect{
		StorageLayoutFlatDirectConfig: config,
	}
	return sl, nil
}

type StorageLayoutFlatDirectConfig struct {
	*extension.ExtensionConfig
}
type StorageLayoutFlatDirect struct {
	*StorageLayoutFlatDirectConfig
	logger ocfllogger.OCFLLogger
}

func (sl *StorageLayoutFlatDirect) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	sl.logger = logger.With("extension", StorageLayoutFlatDirectName)
	return sl
}

func (sl *StorageLayoutFlatDirect) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, sl.StorageLayoutFlatDirectConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal StorageLayoutFlatDirectConfig '%s'", string(data))
	}
	return nil
}

func (sl *StorageLayoutFlatDirect) Terminate() error {
	return nil
}

func (sl *StorageLayoutFlatDirect) GetConfig() any {
	return sl.StorageLayoutFlatDirectConfig
}

func (sl *StorageLayoutFlatDirect) IsRegistered() bool {
	return true
}

func (sl *StorageLayoutFlatDirect) Stat(w io.Writer, statInfo []object.StatInfo) error {
	return nil
}

func (sl *StorageLayoutFlatDirect) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutFlatDirect) GetName() string { return StorageLayoutFlatDirectName }
func (sl *StorageLayoutFlatDirect) WriteConfig(fsys appendfs.FS) error {
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

func (sl *StorageLayoutFlatDirect) WriteLayout(fsys appendfs.FS) error {
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
		Extension:   StorageLayoutFlatDirectName,
		Description: StorageLayoutFlatDirectDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *StorageLayoutFlatDirect) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
	return id, nil
}

// check interface satisfaction
var (
	_ extension.Extension                  = &StorageLayoutFlatDirect{}
	_ storageroot.ExtensionStorageRootPath = &StorageLayoutFlatDirect{}
)
