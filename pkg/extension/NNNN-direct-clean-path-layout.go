package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

// fallback for object with unregigered naming

const LegacyDirectCleanName = "NNNN-direct-clean-path-layout"
const LegacyDirectCleanDescription = "Maps OCFL object identifiers to storage paths or as an object extension that maps logical paths to content paths. This is done by replacing or removing \"dangerous characters\" from names"

func NewLegacyDirectCleanFS(fsys fs.FS) (ocfl.Extension, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &LegacyDirectCleanConfig{
		DirectCleanConfig: &DirectCleanConfig{},
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	// compatibility with old config
	if config.MaxFilenameLen > 0 && config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = config.MaxFilenameLen
		config.MaxFilenameLen = 0
	}
	if config.FallbackSubFolders > 0 && config.NumberOfFallbackTuples == 0 {
		config.NumberOfFallbackTuples = config.FallbackSubFolders
		config.FallbackSubFolders = 0
	}
	return NewLegacyDirectClean(config)
}

func NewLegacyDirectClean(config *LegacyDirectCleanConfig) (ocfl.Extension, error) {
	if config.MaxPathnameLen == 0 {
		config.MaxPathnameLen = 32000
	}
	if config.MaxPathSegmentLen == 0 {
		config.MaxPathSegmentLen = 127
	}
	if config.FallbackDigestAlgorithm == "" {
		config.FallbackDigestAlgorithm = checksum.DigestSHA512
	}
	if config.FallbackFolder == "" {
		config.FallbackFolder = "fallback"
	}

	sl := &LegacyDirectClean{DirectClean: &DirectClean{DirectCleanConfig: config.DirectCleanConfig}}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.Errorf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName())
	}

	var err error
	if sl.hash, err = checksum.GetHash(config.FallbackDigestAlgorithm); err != nil {
		return nil, errors.Wrapf(err, "hash %s not supported", config.FallbackDigestAlgorithm)
	}

	return sl, nil
}

type LegacyDirectCleanConfig struct {
	*DirectCleanConfig
}

type LegacyDirectClean struct {
	*DirectClean
}

func (sl *LegacyDirectClean) IsRegistered() bool {
	return false
}
func (sl *LegacyDirectClean) GetName() string { return LegacyDirectCleanName }
