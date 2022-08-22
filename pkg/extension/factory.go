package extension

import (
	"emperror.dev/errors"
	"encoding/json"
)

var ErrNotSupported = errors.New("storage layout not supported")

func NewDefaultStorageLayout() (StorageLayout, error) {
	var layout StorageLayout
	var err error
	var cfg = &DirectCleanConfig{
		Config:                      &Config{ExtensionName: FlatDirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
	}
	if layout, err = NewFlatDirectClean(cfg); err != nil {
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
	case FlatDirectCleanName:
		var conf = &DirectCleanConfig{
			Config:                      cfg,
			MaxPathnameLen:              32000,
			MaxFilenameLen:              127,
			WhitespaceReplacementString: " ",
			ReplacementString:           "_",
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewFlatDirectClean(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case FlatDirectName:
		if layout, err = NewFlatDirect(cfg); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case HashAndIdNTupleName:
		var conf = &HashAndIdNTupleConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			TupleSize:       0,
			NumberOfTuples:  0,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewHashAndIdNTuple(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case HashedNTupleName:
		var conf = &HashedNTupleConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			TupleSize:       0,
			NumberOfTuples:  0,
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewHashedNTuple(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	case PairTreeName:
		var conf = &PairTreeConfig{
			Config:          cfg,
			DigestAlgorithm: "",
			ShortyLength:    0,
			UriBase:         "",
		}
		if err := json.Unmarshal(config, conf); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal json - %s", string(config))
		}
		if layout, err = NewPairTree(conf); err != nil {
			return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
		}
	default:
		return nil, ErrNotSupported
	}
	return layout, nil
}
