package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

const ZipReaderName = "NNNN-ZipReader"
const ZipReaderDescription = "use of zip packages for versions"

func GetZipReaderParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: ZipReaderName,
			Functions:     []string{"add", "update", "extract", "extractmeta"},
			Description:   "reads zip packages of versions",
		},
	}
}

func NewZipReaderFS(fsys fs.FS) (*ZipReader, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &ZipReaderConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal ZipReaderConfig '%s'", string(data))
	}
	return NewZipReader(config)
}
func NewZipReader(config *ZipReaderConfig) (*ZipReader, error) {
	sl := &ZipReader{
		ZipReaderConfig: config,
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type ZipReaderConfig struct {
	*ocfl.ExtensionConfig
}
type ZipReader struct {
	*ZipReaderConfig
	fsys fs.FS
}

func (sl *ZipReader) Terminate() error {
	return nil
}

func (sl *ZipReader) GetFS() fs.FS {
	return sl.fsys
}

func (sl *ZipReader) GetConfig() any {
	return sl.ZipReaderConfig
}

func (sl *ZipReader) IsRegistered() bool {
	return false
}

func (sl *ZipReader) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *ZipReader) SetParams(params map[string]string) error {
	return nil
}

func (sl *ZipReader) GetName() string { return ZipReaderName }

func (sl *ZipReader) WriteConfig() error {
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
	if err := jenc.Encode(sl.ZipReaderConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *ZipReader) GetPackageReader(version string, objectFS fs.FS, packages ocfl.VersionPackages) (ocfl.PackageReader, error) {
	if packages == nil {
		return nil, nil
	}
	versionPackage, ok := packages.GetVersion(version)
	if !ok {
		return nil, nil
	}
	if versionPackage.Metadata.Format != ocfl.VersionPackageTypeString[ocfl.VersionZIP] {
		return nil, nil
	}
	// Create a new ZipReader for the version package

	//TODO implement me
	panic("implement me")
}

// check interface satisfaction
var (
	_ ocfl.Extension              = &ZipReader{}
	_ ocfl.ExtensionPackageReader = &ZipReader{}
)
