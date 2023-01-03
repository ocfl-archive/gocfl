package checksum

import (
	"fmt"
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
		var resultmap = map[DigestAlgorithm]string{
			DigestMD5:    "ed076287532e86365e841e92bfc50d8c",
			DigestSHA256: "7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069",
			DigestSHA512: "861844d6704e8573fec34d967e20bcfef3d424cf48be04e6dc08f2bd58c729743371015ead891cc3cf1c9d34b49264b510751b1ff9e537937bc46b5d6ff4ecc8",
		}
		fmt.Printf("%v\n", w.GetChecksums())
		for key, val := range w.GetChecksums() {
			if resultmap[key] != val {
				t.Errorf("invalid result for '%s': %s != %s", key, val, resultmap[key])
			}
		}
	}()
}
