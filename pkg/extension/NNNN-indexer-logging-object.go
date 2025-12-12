package extension

import (
	"fmt"

	"io"
	"io/fs"
	"net/url"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

const LoggingIndexerName = "NNNN-indexer-logging-object"

type LoggingIndexerConfig struct {
	*Config
}

type LoggingIndexer struct {
	*LoggingIndexerConfig
	metadata map[string]any
}

func (sl *LoggingIndexer) Terminate() error {
	return nil
}

func (sl *LoggingIndexer) GetFS() fs.FS {
	//TODO implement me
	panic("implement me")
}

func (sl *LoggingIndexer) GetConfig() any {
	//TODO implement me
	panic("implement me")
}

func (sl *LoggingIndexer) IsRegistered() bool {
	return false
}

func (li *LoggingIndexer) SetFS(fsys fs.FS, create bool) {
	//TODO implement me
	panic("implement me")
}

func (li *LoggingIndexer) SetParams(params map[string]string) error {
	//TODO implement me
	panic("implement me")
}

func (li *LoggingIndexer) WriteConfig() error {
	//TODO implement me
	panic("implement me")
}

func NewLoggingIndexer(config *LoggingIndexerConfig) (*LoggingIndexer, error) {
	li := &LoggingIndexer{LoggingIndexerConfig: config, metadata: map[string]any{}}
	if config.ExtensionName != li.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, li.GetName()))
	}
	return li, nil
}

func (li *LoggingIndexer) GetName() string {
	return LoggingIndexerName
}
func (li *LoggingIndexer) Start() error {
	li.metadata = map[string]any{}
	return nil
}
func (li *LoggingIndexer) AddFile(fullpath url.URL) error {
	return nil
}

func (li *LoggingIndexer) MoveFile(fullpath url.URL) error {
	return nil

}

func (li *LoggingIndexer) DeleteFile(fullpath url.URL) error {
	return nil

}

func (li *LoggingIndexer) WriteLog(logfile io.Writer) error {
	return nil

}

var (
	_ extension.Extension = &LoggingIndexer{}
)
