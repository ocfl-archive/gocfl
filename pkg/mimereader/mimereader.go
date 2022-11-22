package mimereader

import (
	"bytes"
	"emperror.dev/errors"
	"io"
	"mime"
	"net/http"
)

const detectSize = 512

type MimeReader struct {
	io.Reader
	buffer   *bytes.Buffer
	mimetype string
}

func NewMimeReader(reader io.Reader) (*MimeReader, error) {
	mr := &MimeReader{Reader: reader}
	return mr, mr.init()
}

func (mr *MimeReader) init() error {
	var data = make([]byte, detectSize)
	num, err := mr.Reader.Read(data)
	if err != nil {
		return errors.Wrap(err, "cannot read from input")
	}
	if num == 0 {
		return errors.New("no data read")
	}
	contentType := http.DetectContentType(data)
	mr.mimetype, _, _ = mime.ParseMediaType(contentType)
	mr.buffer = bytes.NewBuffer(data[:num])
	return nil
}

func (mr *MimeReader) Read(p []byte) (n int, err error) {
	if mr.buffer.Len() > 0 {
		l, err := mr.buffer.Read(p)
		if err != nil {
			return l, err
		}
		if l < cap(p) {
			data := make([]byte, cap(p)-l)
			l2, err := mr.Reader.Read(data)
			if err != nil {
				return l + l2, err
			}
			l3 := copy(p[len(p):], data)
			if l3 != l2 {
				return 0, errors.New("error copying additional data")
			}
			l += l3
		}
		return l, nil
	}
	return mr.Reader.Read(p)
}

func (mr *MimeReader) GetMimetype() string {
	return mr.mimetype
}
