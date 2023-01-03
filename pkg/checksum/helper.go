package checksum

import (
	"emperror.dev/errors"
	"fmt"
	"io"
)

func Copy(dst io.Writer, src io.Reader, checksums []DigestAlgorithm) (map[DigestAlgorithm]string, error) {
	cw := NewChecksumWriter(checksums, dst)
	if _, err := io.Copy(cw, src); err != nil {
		cw.Close()
		return nil, errors.Wrap(err, "cannot copy")
	}
	if err := cw.Close(); err != nil {
		return nil, errors.Wrap(err, "error closing checksumwriter")
	}
	return cw.GetChecksums(), nil
}

func Checksum(src io.Reader, checksum DigestAlgorithm) (string, error) {
	sink, err := GetHash(checksum)
	if err != nil {
		return "", errors.New(fmt.Sprintf("invalid checksum type %s", checksum))
	}
	if _, err := io.Copy(sink, src); err != nil {
		return "", errors.Wrapf(err, "cannot create checkum %s", checksum)
	}
	csString := fmt.Sprintf("%x", sink.Sum(nil))
	return csString, nil
}
