package extension

import (
	"encoding/json"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

// fallback for object with unregigered naming

const LegacyDirectCleanName = "NNNN-direct-clean-path-layout"
const LegacyDirectCleanDescription = "Maps OCFL object identifiers to storage paths or as an object extension that maps logical paths to content paths. This is done by replacing or removing \"dangerous characters\" from names"

func NewLegacyDirectClean(logger ocfllogger.OCFLLogger) (extension.Extension, error) {
	config := &LegacyDirectCleanConfig{
		DirectCleanConfig: &DirectCleanConfig{
			ExtensionConfig: &extension.ExtensionConfig{ExtensionName: LegacyDirectCleanName},
		},
	}
	sl := &LegacyDirectClean{DirectClean: &DirectClean{DirectCleanConfig: config.DirectCleanConfig, logger: logger.With("extension", LegacyDirectCleanName)}}
	return sl, nil
}

type LegacyDirectCleanConfig struct {
	*DirectCleanConfig
}

type LegacyDirectClean struct {
	*DirectClean
}

func (sl *LegacyDirectClean) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.DirectCleanConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal LegacyDirectCleanConfig '%s'", string(data))
	}
	// compatibility with old config
	if sl.MaxFilenameLen > 0 && sl.MaxPathnameLen == 0 {
		sl.MaxPathnameLen = sl.MaxFilenameLen
		sl.MaxFilenameLen = 0
	}
	if sl.FallbackSubFolders > 0 && sl.NumberOfFallbackTuples == 0 {
		sl.NumberOfFallbackTuples = sl.FallbackSubFolders
		sl.FallbackSubFolders = 0
	}
	// defaults
	if sl.MaxPathnameLen == 0 {
		sl.MaxPathnameLen = 32000
	}
	if sl.MaxPathSegmentLen == 0 {
		sl.MaxPathSegmentLen = 127
	}
	if sl.FallbackDigestAlgorithm == "" {
		sl.FallbackDigestAlgorithm = checksum.DigestSHA512
	}
	if sl.FallbackFolder == "" {
		sl.FallbackFolder = "fallback"
	}
	// prepare hash
	if sl.hash, err = checksum.GetHash(sl.FallbackDigestAlgorithm); err != nil {
		return errors.Wrapf(err, "hash %s not supported", sl.FallbackDigestAlgorithm)
	}
	return nil
}

func (sl *LegacyDirectClean) IsRegistered() bool {
	return false
}
func (sl *LegacyDirectClean) GetName() string { return LegacyDirectCleanName }

func (sl *LegacyDirectClean) SetParams(params map[string]string) error {
	return nil
}

func (sl *LegacyDirectClean) Terminate() error {
	return nil
}

func (sl *LegacyDirectClean) GetConfig() any {
	return sl.DirectCleanConfig
}

func (sl *LegacyDirectClean) WriteConfig(fsys streamfs.FS) error {
	return nil
}

func (sl *LegacyDirectClean) GetFS() fs.FS {
	return nil
}

func (sl *LegacyDirectClean) SetFS(fsys fs.FS, create bool) {}
