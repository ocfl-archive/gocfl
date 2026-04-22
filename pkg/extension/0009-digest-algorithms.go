package extension

import (
	"encoding/json"

	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const DigestAlgorithmsName = "0009-digest-algorithms"
const DigestAlgorithmsDescription = "controlled vocabulary of digest algorithm names that may be used to indicate the given algorithm in fixity blocks of OCFL Objects"

func init() {
	extension.RegisterExtension(DigestAlgorithmsName, NewDigestAlgorithms, nil)
}

var algorithms = []checksum.DigestAlgorithm{
	checksum.DigestBlake2b160,
	checksum.DigestBlake2b256,
	checksum.DigestBlake2b384,
	checksum.DigestBlake2b512,
	checksum.DigestMD5,
	checksum.DigestSHA512,
	checksum.DigestSHA256,
	checksum.DigestSHA1,
	checksum.DigestSize,
}

func NewDigestAlgorithms() (extension.Extension, error) {
	var config = &DigestAlgorithmsConfig{
		ExtensionConfig: &extension.ExtensionConfig{
			ExtensionName: DigestAlgorithmsName,
		},
	}
	sl := &DigestAlgorithms{
		DigestAlgorithmsConfig: config,
	}
	return sl, nil
}

type DigestAlgorithmsConfig struct {
	*extension.ExtensionConfig
}
type DigestAlgorithms struct {
	*DigestAlgorithmsConfig
	logger ocfllogger.OCFLLogger
}

func (sl *DigestAlgorithms) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	sl.logger = logger.With("extension", DigestAlgorithmsName)
	return sl
}

func (sl *DigestAlgorithms) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, sl.DigestAlgorithmsConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal DigestAlgorithmsConfig '%s'", string(data))
	}
	return nil
}

func (sl *DigestAlgorithms) Terminate() error {
	return nil
}

func (sl *DigestAlgorithms) GetConfig() any {
	return sl.DigestAlgorithmsConfig
}

func (sl *DigestAlgorithms) IsRegistered() bool {
	return true
}

func (sl *DigestAlgorithms) GetFixityDigests() []checksum.DigestAlgorithm {
	return algorithms
}

func (sl *DigestAlgorithms) GetName() string { return DigestAlgorithmsName }

func (sl *DigestAlgorithms) SetParams(params map[string]string) error {
	return nil
}

func (sl *DigestAlgorithms) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// check interface satisfaction
var (
	_ extension.Extension          = &DigestAlgorithms{}
	_ object.ExtensionFixityDigest = &DigestAlgorithms{}
)
