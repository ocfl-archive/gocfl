package mimereader

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"testing"
)

//go:embed test.png
var testImage []byte

func TestMimeReader(t *testing.T) {
	buf := bytes.NewBuffer(testImage)
	mimeReader, err := NewMimeReader(buf)
	if err != nil {
		t.Fatalf("cannot create mimereader: %v", err)
	}
	if mimeReader.GetMimetype() != "image/png" {
		t.Errorf("invalid mimetype: %s", mimeReader.GetMimetype())
	}
	data, err := io.ReadAll(mimeReader)
	if err != nil {
		t.Fatalf("cannot read from mimereader: %v", err)
	}
	if bytes.Compare(testImage, data) != 0 {
		t.Errorf("read data is wrong")
	}
	fmt.Println(testImage)
	fmt.Println("done")
}
