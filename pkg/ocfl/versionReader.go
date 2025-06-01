package ocfl

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
)

type VersionReader interface {
	GetVersion() string
	GetFS() (fs.FS, io.Closer, error)
	GetFilenameChecksum(digestAlgorithm checksum.DigestAlgorithm, fixityAlgorithms []checksum.DigestAlgorithm, fullContentFiles []string) (fileChecksums map[string]map[checksum.DigestAlgorithm]string, fullContent map[string][]byte, partsChecksum map[string]string, error error)
	GetContentFilename() ([]string, error)
}
