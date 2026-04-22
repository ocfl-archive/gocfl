package extension

import (
	"io"
	"io/fs"
	"net/url"

	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const LoggingIndexerName = "NNNN-indexer-logging-object"

type LoggingIndexerConfig struct {
	*Config
}

type LoggingIndexer struct {
	*LoggingIndexerConfig
	metadata map[string]any
	logger   ocfllogger.OCFLLogger
}

func (li *LoggingIndexer) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	li.logger = logger.With("extension", LoggingIndexerName)
	return li
}

func (sl *LoggingIndexer) Load(fsys fs.FS) error {
	// no config file currently defined; placeholder to satisfy interface
	return nil
}

func (sl *LoggingIndexer) Terminate() error {
	return nil
}

func (sl *LoggingIndexer) GetConfig() any {
	//TODO implement me
	panic("implement me")
}

func (sl *LoggingIndexer) IsRegistered() bool {
	return false
}

func (li *LoggingIndexer) SetParams(params map[string]string) error {
	//TODO implement me
	panic("implement me")
}

func (li *LoggingIndexer) WriteConfig(appendfs.FS) error {
	//TODO implement me
	panic("implement me")
}

func NewLoggingIndexer() (*LoggingIndexer, error) {
	config := &LoggingIndexerConfig{Config: &Config{ExtensionName: LoggingIndexerName}}
	li := &LoggingIndexer{LoggingIndexerConfig: config, metadata: map[string]any{}}
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
	_ extensiontypes.Extension = &LoggingIndexer{}
)
