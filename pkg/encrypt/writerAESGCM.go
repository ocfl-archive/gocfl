package encrypt

import (
	"emperror.dev/errors"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/proto/tink_go_proto"
	"github.com/google/tink/go/streamingaead"
	"github.com/je4/utils/v2/pkg/concurrentWriter"
	"io"
)

type KeyStruct struct {
	EncryptedKey string `json:"encrypted_key"`
	Aad          string `json:"associated_data"`
}

type WriterAESGCM struct {
	*concurrentWriter.ConcurrentWriter
	encWriter io.WriteCloser
	handle    *keyset.Handle
}

func NewEncryptWriterAESGCM(dest io.Writer, aad []byte, keyTemplate *tink_go_proto.KeyTemplate, writer ...io.Writer) (*WriterAESGCM, error) {
	var err error
	c := &WriterAESGCM{}

	if keyTemplate == nil {
		keyTemplate = streamingaead.AES256GCMHKDF1MBKeyTemplate()
	}
	c.handle, err = keyset.NewHandle(keyTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create keyset handle")
	}

	a, err := streamingaead.New(c.handle)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create streamingaead")
	}

	c.encWriter, err = a.NewEncryptingWriter(dest, aad)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create encrypting writer")
	}

	runner := concurrentWriter.NewGenericCopyRunner(c.encWriter, "aesgcm")
	c.ConcurrentWriter = concurrentWriter.NewConcurrentWriter([]concurrentWriter.WriterRunner{runner}, writer...)

	return c, nil
}

func (c *WriterAESGCM) Close() error {
	var errs = []error{}
	if err := c.encWriter.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "cannot close encrypting writer"))
	}
	if err := c.ConcurrentWriter.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "cannot close concurrent writer"))
	}
	return errors.Combine(errs...)
}

func (c *WriterAESGCM) GetKeysetHandle() *keyset.Handle {
	return c.handle
}

var (
	_ io.WriteCloser = (*WriterAESGCM)(nil)
)
