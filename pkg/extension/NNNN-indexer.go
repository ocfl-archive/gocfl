package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
)

const IndexerName = "NNNN-indexer"
const IndexerDescription = "technical metadata for all files"

func GetIndexerParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: IndexerName,
			Param:         "indexer-url",
			File:          "IndexerUrl",
			Description:   "url for indexer format recognition service",
		},
	}
}

type IndexerConfig struct {
	*ocfl.ExtensionConfig
}
type Indexer struct {
	*IndexerConfig
	fs ocfl.OCFLFS
}

func (sl *Indexer) GetConfigString() string {
	//TODO implement me
	panic("implement me")
}

func NewIndexerFS(fs ocfl.OCFLFS) (*Indexer, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &IndexerConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewIndexer(config)
}
func NewIndexer(config *IndexerConfig) (*Indexer, error) {
	sl := &Indexer{IndexerConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (sl *Indexer) GetName() string { return IndexerName }

func (sl *Indexer) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *Indexer) SetParams(params map[string]string) error {
	return nil
}

func (sl *Indexer) WriteConfig() error {
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

func (sl *Indexer) AddFileBefore(object ocfl.Object, source, dest string) error {
	return nil
}
func (sl *Indexer) UpdateFileBefore(object ocfl.Object, source, dest string) error {
	return nil
}
func (sl *Indexer) DeleteFileBefore(object ocfl.Object, dest string) error {
	// nothing to do
	return nil
}
func (sl *Indexer) AddFileAfter(object ocfl.Object, source, dest string) error {
	return nil
}
func (sl *Indexer) UpdateFileAfter(object ocfl.Object, source, dest string) error {
	return nil
}
func (sl *Indexer) DeleteFileAfter(object ocfl.Object, dest string) error {
	// nothing to do
	return nil
}

var (
	_ ocfl.Extension = &Indexer{}
)
