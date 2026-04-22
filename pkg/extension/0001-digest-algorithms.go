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

const DigestAlgorithmsName0001 = "0001-digest-algorithms"
const DigestAlgorithmsDescription0001 = "controlled vocabulary of digest algorithm names that may be used to indicate the given algorithm in fixity blocks of OCFL Objects"

func init() {
	extension.RegisterExtension(DigestAlgorithmsName0001, NewDigestAlgorithms0001, nil)
}

var algorithms0001 = []checksum.DigestAlgorithm{
	checksum.DigestBlake2b160,
	checksum.DigestBlake2b256,
	checksum.DigestBlake2b384,
	checksum.DigestBlake2b512,
	checksum.DigestMD5,
	checksum.DigestSHA512,
	checksum.DigestSHA256,
	checksum.DigestSHA1,
}

func NewDigestAlgorithms0001() (extension.Extension, error) {
	var config = &DigestAlgorithmsConfig0001{
		ExtensionConfig: &extension.ExtensionConfig{
			ExtensionName: DigestAlgorithmsName0001,
		},
	}
	sl := &DigestAlgorithms0001{
		DigestAlgorithmsConfig0001: config,
	}
	return sl, nil
}

type DigestAlgorithmsConfig0001 struct {
	*extension.ExtensionConfig
}
type DigestAlgorithms0001 struct {
	*DigestAlgorithmsConfig0001
	logger ocfllogger.OCFLLogger
}

func (sl *DigestAlgorithms0001) WithLogger(logger ocfllogger.OCFLLogger) extension.Extension {
	sl.logger = logger.With("extension", DigestAlgorithmsName0001)
	return sl
}

func (sl *DigestAlgorithms0001) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, sl.DigestAlgorithmsConfig0001); err != nil {
		return errors.Wrapf(err, "cannot unmarshal DigestAlgorithmsConfig0001 '%s'", string(data))
	}
	return nil
}

func (sl *DigestAlgorithms0001) Terminate() error {
	return nil
}

func (sl *DigestAlgorithms0001) GetConfig() any {
	return sl.DigestAlgorithmsConfig0001
}

func (sl *DigestAlgorithms0001) IsRegistered() bool {
	return true
}

func (sl *DigestAlgorithms0001) GetFixityDigests() []checksum.DigestAlgorithm {
	return algorithms0001
}

func (sl *DigestAlgorithms0001) GetName() string { return DigestAlgorithmsName0001 }

func (sl *DigestAlgorithms0001) SetParams(params map[string]string) error {
	return nil
}

func (sl *DigestAlgorithms0001) WriteConfig(fsys appendfs.FS) error {
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
	_ extension.Extension          = &DigestAlgorithms0001{}
	_ object.ExtensionFixityDigest = &DigestAlgorithms0001{}
)
