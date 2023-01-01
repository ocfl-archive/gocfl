package checksum

import (
	"os"
	"testing"
)

func TestChecksumWriter(t *testing.T) {
	w := NewChecksumWriter(
		[]DigestAlgorithm{DigestSHA512, DigestSHA256, DigestMD5},
		os.Stdout,
	)
	if _, err := w.Write([]byte("Hello World!")); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	}()
}
