package storageroot

import (
	"emperror.dev/errors"
	"encoding/json"
	"io"
)

type StorageLayout interface {
	ExecuteID(id string) (string, error)
	Name() string
	WriteConfig(config io.Writer) error
}

type Config struct {
	ExtensionName string `json:"extensionName"`
}

var ErrNotSupported = errors.New("storage layout not supported")

func NewDefaultStorageLayout() (StorageLayout, error) {
	var layout StorageLayout
	var err error
	var cfg = &StorageLayoutDirectCleanConfig{
		Config:                      &Config{ExtensionName: StorageLayoutDirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
		UTFEncode:                   true,
	}
	if layout, err = NewStorageLayoutDirectClean(cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
	}
	return layout, nil
}

func NewStorageLayout(config []byte) (StorageLayout, error) {
	var cfg = &Config{}
	if err := json.Unmarshal(config, cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
	}
	var layout StorageLayout
	var err error
	switch cfg.ExtensionName {
	case StorageLayoutDirectCleanName:
		var conf = &StorageLayoutDirectCleanConfig{
			Config:                      cfg,
			MaxPathnameLen:              32000,
			MaxFilenameLen:              127,
			WhitespaceReplacementString: " ",
			ReplacementString:           "_",
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewStorageLayoutDirectClean(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case StorageLayoutFlatDirectName:
		if layout, err = NewStorageLayoutFlatDirect(cfg); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case StorageLayoutHashAndIdNTupleName:
		var conf = &StorageLayoutHashAndIdNTupleConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			TupleSize:       0,
			NumberOfTuples:  0,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewStorageLayoutHashAndIdNTuple(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case StorageLayoutHashedNTupleName:
		var conf = &StorageLayoutHashedNTupleConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			TupleSize:       0,
			NumberOfTuples:  0,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewStorageLayoutHashedNTuple(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case StorageLayoutPairTreeName:
		var conf = &StorageLayoutPairTreeConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			ShortyLength:    0,
			UriBase:         "",
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewStorageLayoutPairTree(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	default:
		return nil, ErrNotSupported
	}
	return layout, nil
}
