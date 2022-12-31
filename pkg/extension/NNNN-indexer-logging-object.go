package extension

import (
	"emperror.dev/errors"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"net/url"
)

const LoggingIndexerName = "NNNN-indexer-logging-object"

type LoggingIndexerConfig struct {
	*Config
}

type LoggingIndexer struct {
	*LoggingIndexerConfig
	metadata map[string]any
}

func (li *LoggingIndexer) SetFS(fs ocfl.OCFLFS) {
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

func (li *LoggingIndexer) GetConfigString() string {
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
	_ ocfl.Extension = &LoggingIndexer{}
)
