package checksum

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/concurrentWriter"
	"io"
	"sync"
)

type WriterRunnerChecksum struct {
	alg       DigestAlgorithm
	digest    string
	errors    []error
	errorLock sync.Mutex
}

func NewWriterRunnerChecksum(alg DigestAlgorithm) (*WriterRunnerChecksum, error) {
	if !HashExists(alg) {
		return nil, errors.Errorf("unknown hash algorithm '%s'", alg)
	}
	wrc := &WriterRunnerChecksum{
		alg:       alg,
		errorLock: sync.Mutex{},
	}
	return wrc, nil
}

func (w *WriterRunnerChecksum) setError(err error) {
	w.errorLock.Lock()
	defer w.errorLock.Unlock()
	w.errors = append(w.errors, err)
}

func (w *WriterRunnerChecksum) Do(reader io.Reader, done chan bool) {
	// we should end in all cases
	defer func() {
		done <- true
	}()

	sink, err := GetHash(w.alg)
	if err != nil {
		w.setError(errors.New(fmt.Sprintf("invalid hash function %s", w.alg)))
		null := &NullWriter{}
		io.Copy(null, reader)
		return
	}
	if _, err := io.Copy(sink, reader); err != nil {
		w.setError(errors.Wrapf(err, "cannot create checkum %s", w.alg))
		return
	}
	w.digest = fmt.Sprintf("%x", sink.Sum(nil))
}

func (w *WriterRunnerChecksum) GetName() string {
	return fmt.Sprintf("WriterRunnerChecksum_%s", w.alg)
}

func (w *WriterRunnerChecksum) GetAlgorithm() DigestAlgorithm {
	return w.alg
}

func (w *WriterRunnerChecksum) GetDigest() (string, error) {
	w.errorLock.Lock()
	defer w.errorLock.Unlock()
	return w.digest, errors.Combine(w.errors...)
}

var (
	_ concurrentWriter.WriterRunner = (*WriterRunnerChecksum)(nil)
)
