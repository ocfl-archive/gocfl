package ocfl

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
)

type gfcCallback func(name string, reader io.Reader) error

type VersionReader interface {
	GetVersion() string
	GetFS() (fs.FS, io.Closer, error)
	GetFilenameChecksum(digestAlgorithm checksum.DigestAlgorithm, fixityAlgorithms []checksum.DigestAlgorithm, fn gfcCallback) (fileChecksums map[string]map[checksum.DigestAlgorithm]string, partsChecksum map[string]string, error error)
	GetContentFilename() ([]string, error)
}
