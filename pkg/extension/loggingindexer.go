package extension

import (
	"emperror.dev/errors"
	"fmt"
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

func NewLoggingIndexer(config *LoggingIndexerConfig) (*LoggingIndexer, error) {
	li := &LoggingIndexer{LoggingIndexerConfig: config, metadata: map[string]any{}}
	if config.ExtensionName != li.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, li.Name()))
	}
	return li, nil
}

func (li *LoggingIndexer) Name() string {
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
func (li *LoggingIndexer) WriteConfig(config io.Writer) error {
	return nil

}
