package encrypt

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/concurrentWriter"
	"io"
)

type WriterRunnerEncrypt interface {
	concurrentWriter.WriterRunner
	GetKey() []byte
	GetIV() []byte
}

type EncryptWriter struct {
	writer *concurrentWriter.ConcurrentWriter
	runner WriterRunnerEncrypt
}

func NewEncryptWriterAES(dest io.Writer, encKey, IV []byte, writer ...io.Writer) (*EncryptWriter, error) {
	runner, err := NewWriterRunnerEncryptAES(dest, encKey, IV)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create  WriterRunnerEncryptAES")
	}
	w := concurrentWriter.NewConcurrentWriter([]concurrentWriter.WriterRunner{runner}, writer...)
	c := &EncryptWriter{
		writer: w,
		runner: runner,
	}
	return c, nil
}

func (c *EncryptWriter) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c *EncryptWriter) Close() error {
	if err := c.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close concurrent writer")
	}
	return nil
}

func (c *EncryptWriter) GetKey() []byte {
	return c.runner.GetKey()
}

func (c *EncryptWriter) GetIV() []byte {
	return c.runner.GetIV()
}

var (
	_ io.WriteCloser = (*EncryptWriter)(nil)
)
