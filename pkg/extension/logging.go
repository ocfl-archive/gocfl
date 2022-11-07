package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"io"
	"net/url"
)

type Logging interface {
	Name() string
	Start() error
	AddFile(fullpath url.URL) error
	DeleteFile(fullpath url.URL) error
	MoveFile(src, target url.URL) error
	WriteLog(logfile io.Writer) error
	WriteConfig(config io.Writer) error
}

func NewLogging(config []byte) (Path, error) {
	var cfg = &Config{}
	if err := json.Unmarshal(config, cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
	}
	var path Path
	var err error
	switch cfg.ExtensionName {
	case PathDirectName:
		var conf = &PathDirectConfig{
			Config: cfg,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if path, err = NewPathDirect(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case PathDirectCleanName:
		var conf = &PathDirectCleanConfig{
			Config: cfg,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if path, err = NewPathDirectClean(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	default:
		return nil, ErrNotSupported
	}
	return path, nil
}
