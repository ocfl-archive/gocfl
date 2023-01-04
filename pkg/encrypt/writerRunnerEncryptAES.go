package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"emperror.dev/errors"
	"github.com/blend/go-sdk/crypto"
	"io"
	"sync"
)

type WriterRunnerEncryptAES struct {
	writer io.Writer
	//	block     cipher.Block
	//	stream    cipher.Stream
	iv        []byte
	encKey    []byte
	errors    []error
	errorLock sync.Mutex
}

func NewWriterRunnerEncryptAES(writer io.Writer, encKey, IV []byte) (*WriterRunnerEncryptAES, error) {
	if encKey == nil {
		var err error
		encKey, err = crypto.CreateKey(crypto.DefaultKeySize)
		if err != nil {
			return nil, errors.Wrap(err, "cannot generate key")
		}
	}
	if IV == nil {
		IV = make([]byte, aes.BlockSize)
		if _, err := rand.Read(IV); err != nil {
			return nil, errors.Wrap(err, "cannot read random IV")
		}
	}
	/*
		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create cipher block")
		}
	*/
	wrc := &WriterRunnerEncryptAES{
		writer: writer,
		//		block:     block,
		iv:     IV,
		encKey: encKey,
		//		stream:    cipher.NewCTR(block, IV),
		errorLock: sync.Mutex{},
	}
	return wrc, nil
}

func (w *WriterRunnerEncryptAES) setError(err error) {
	w.errorLock.Lock()
	defer w.errorLock.Unlock()
	w.errors = append(w.errors, err)
}

func (w *WriterRunnerEncryptAES) Do(reader io.Reader, done chan bool) {
	// we should end in all cases
	defer func() {
		done <- true
	}()

	block, err := aes.NewCipher(w.encKey)
	if err != nil {
		w.setError(errors.Wrap(err, "cannot create cipher block"))
		return
	}

	cReader := &cipher.StreamReader{
		S: cipher.NewCTR(block, w.iv),
		R: reader,
	}

	if _, err := io.Copy(w.writer, cReader); err != nil {
		w.setError(errors.Wrapf(err, "cannot write cipher stream"))
		return
	}

}

func (w *WriterRunnerEncryptAES) GetName() string {
	return "WriterRunnerEncryptAES"
}

func (w *WriterRunnerEncryptAES) GetKey() []byte {
	return w.encKey
}

func (w *WriterRunnerEncryptAES) GetIV() []byte {
	return w.iv
}

var (
	_ WriterRunnerEncrypt = (*WriterRunnerEncryptAES)(nil)
)
