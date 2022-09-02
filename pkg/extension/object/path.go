package object

import (
	"emperror.dev/errors"
	"encoding/json"
	"io"
)

type Path interface {
	ExecutePath(id string) (string, error)
	Name() string
	WriteConfig(config io.Writer) error
}

var ErrNotSupported = errors.New("path extension not supported")

func NewDefaultPath() (Path, error) {
	var cfg = &PathDirectConfig{
		Config: &Config{ExtensionName: DirectPathName},
	}
	path, err := NewPathDirect(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
	}
	return path, nil
}

func NewPath(config []byte) (Path, error) {
	var cfg = &Config{}
	if err := json.Unmarshal(config, cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
	}
	var path Path
	var err error
	switch cfg.ExtensionName {
	case DirectPathName:
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
