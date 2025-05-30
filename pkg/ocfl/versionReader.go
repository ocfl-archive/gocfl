package ocfl

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
)

type VersionReader interface {
	GetVersion() string
	GetFS() (fs.FS, io.Closer, error)
	GetContentFilenameChecksum(digestAlgs []checksum.DigestAlgorithm) (map[string]map[checksum.DigestAlgorithm]string, error)
	GetContentFilename() ([]string, error)
}
