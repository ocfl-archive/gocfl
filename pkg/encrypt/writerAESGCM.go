package encrypt

import (
	"emperror.dev/errors"
	"encoding/base64"
	"encoding/json"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/proto/tink_go_proto"
	"github.com/google/tink/go/streamingaead"
	"github.com/je4/utils/v2/pkg/concurrentWriter"
	"io"
)

type KeyStruct struct {
	EncryptedKey Base64 `json:"encrypted_key"`
	Aad          Base64 `json:"associated_data"`
}

type Base64 []byte

func (b64 *Base64) MarshalJSON() ([]byte, error) {
	resultString := base64.StdEncoding.EncodeToString(*b64)
	return json.Marshal(resultString)
}

func (b64 *Base64) UnmarshalJSON(data []byte) error {
	var x string
	if err := json.Unmarshal(data, &x); err != nil {
		return errors.Wrap(err, "cannot unmarshal base64")
	}
	y, err := base64.StdEncoding.DecodeString(x)
	if err != nil {
		return errors.Wrap(err, "cannot decode base64")
	}
	*b64 = y
	return nil
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
