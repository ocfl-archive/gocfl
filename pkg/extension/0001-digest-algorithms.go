package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
)

const DigestAlgorithmsName = "0001-digest-algorithms"
const DigestAlgorithmsDescription = "controlled vocabulary of digest algorithm names that may be used to indicate the given algorithm in fixity blocks of OCFL Objects"

var algorithms = []checksum.DigestAlgorithm{
	checksum.DigestBlake2b160,
	checksum.DigestBlake2b256,
	checksum.DigestBlake2b384,
	checksum.DigestBlake2b512,
	checksum.DigestMD5,
	checksum.DigestSHA512,
	checksum.DigestSHA256,
	checksum.DigestSHA1,
}

type DigestAlgorithmsConfig struct {
	*ocfl.ExtensionConfig
}
type DigestAlgorithms struct {
	*DigestAlgorithmsConfig
	fs ocfl.OCFLFSRead
}

func NewDigestAlgorithmsFS(fsys ocfl.OCFLFSRead) (*DigestAlgorithms, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &DigestAlgorithmsConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewDigestAlgorithms(config)
}

func NewDigestAlgorithms(config *DigestAlgorithmsConfig) (*DigestAlgorithms, error) {
	sl := &DigestAlgorithms{DigestAlgorithmsConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (sl *DigestAlgorithms) IsRegistered() bool {
	return true
}

func (sl *DigestAlgorithms) GetFixityDigests() []checksum.DigestAlgorithm {
	return algorithms
}

func (sl *DigestAlgorithms) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.DigestAlgorithmsConfig, "", "  ")
	return string(str)
}

func (sl *DigestAlgorithms) SetFS(fs ocfl.OCFLFSRead) {
	sl.fs = fs
}

func (sl *DigestAlgorithms) SetParams(params map[string]string) error {
	return nil
}

func (sl *DigestAlgorithms) GetName() string { return DigestAlgorithmsName }
func (sl *DigestAlgorithms) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	fsRW, ok := sl.fs.(ocfl.OCFLFS)
	if !ok {
		return errors.Errorf("filesystem is read only - '%s'", sl.fs.String())
	}
	configWriter, err := fsRW.Create("config.json")
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
	_ ocfl.Extension             = &DigestAlgorithms{}
	_ ocfl.ExtensionFixityDigest = &DigestAlgorithms{}
)
