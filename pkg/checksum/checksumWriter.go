package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"golang.org/x/crypto/sha3"
	"hash"
	"io"
	"sync"
)

type DigestAlgorithm string

const (
	DigestMD5        DigestAlgorithm = "md5"
	DigestSHA1       DigestAlgorithm = "sha1"
	DigestSHA256     DigestAlgorithm = "sha256"
	DigestSHA512     DigestAlgorithm = "sha512"
	DigestBlake2b512 DigestAlgorithm = "blake2b-512"
)

type rwStruct struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

// ChecksumWriter creates concurrent threads for writing and creating checksums
type ChecksumWriter struct {
	sync.Mutex
	checksums []DigestAlgorithm
	cs        map[DigestAlgorithm]string
	errors    []error
	rws       map[DigestAlgorithm]rwStruct
	dataLock  sync.Mutex
}

func GetHash(csType DigestAlgorithm) (hash.Hash, error) {
	var sink hash.Hash
	switch csType {
	case "md5":
		sink = md5.New()
	case "sha1":
		sink = sha1.New()
	case "sha256":
		sink = sha256.New()
	case "sha512":
		sink = sha512.New()
	case "sha3-256":
		sink = sha3.New256()
	case "sha3-384":
		sink = sha3.New384()
	case "sha3-512":
		sink = sha3.New512()
	default:
		return nil, fmt.Errorf("unknown checksum %s", csType)
	}
	return sink, nil
}

func NewChecksumWriter(checksums []DigestAlgorithm) *ChecksumWriter {
	c := &ChecksumWriter{
		Mutex:     sync.Mutex{},
		checksums: checksums,
		cs:        map[DigestAlgorithm]string{},
		errors:    []error{},
		rws:       map[DigestAlgorithm]rwStruct{},
		dataLock:  sync.Mutex{},
	}
	return c
}

func ChecksumCopy(dst io.Writer, src io.Reader, checksums []DigestAlgorithm) (map[DigestAlgorithm]string, error) {
	cw := NewChecksumWriter(checksums)
	return cw.Copy(dst, src)
}

func Checksum(src io.Reader, checksum DigestAlgorithm) (string, error) {
	sink, err := GetHash(checksum)
	if err != nil {
		return "", errors.New(fmt.Sprintf("invalid checksum type %s", checksum))
	}
	if _, err := io.Copy(sink, src); err != nil {
		return "", emperror.Wrapf(err, "cannot create checkum %s", checksum)
	}
	csString := fmt.Sprintf("%x", sink.Sum(nil))
	return csString, nil
}

// start ChecksumWriter process
// supported csType's: md5, sha256, sha512
func (c *ChecksumWriter) doChecksum(reader io.Reader, csType DigestAlgorithm, done chan bool) {
	// we should end in all cases
	defer func() { done <- true }()

	sink, err := GetHash(csType)
	if err != nil {
		c.setError(errors.New(fmt.Sprintf("invalid hash function %s", csType)))
		null := &NullWriter{}
		io.Copy(null, reader)
		return
	}
	if _, err := io.Copy(sink, reader); err != nil {
		c.setError(emperror.Wrapf(err, "cannot create checkum %s", csType))
		return
	}
	csString := fmt.Sprintf("%x", sink.Sum(nil))
	c.setResult(csType, csString)
}

func (c *ChecksumWriter) setResult(csType DigestAlgorithm, checksum string) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.cs[csType] = checksum
}

func (c *ChecksumWriter) setError(err error) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.errors = append(c.errors, err)
}

func (c *ChecksumWriter) clear() {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.errors = []error{}
	c.cs = map[DigestAlgorithm]string{}
}

func (c *ChecksumWriter) Copy(dst io.Writer, src io.Reader) (map[DigestAlgorithm]string, error) {
	c.Lock()
	defer c.Unlock()

	c.clear()

	done := make(chan bool)
	// create the map of all ChecksumWriter-pipes and start async process
	for _, csType := range c.checksums {
		rw := rwStruct{}
		rw.reader, rw.writer = io.Pipe()
		c.rws[csType] = rw
		go c.doChecksum(rw.reader, csType, done)
	}

	rw := rwStruct{}
	rw.reader, rw.writer = io.Pipe()
	c.rws["_"] = rw
	go func() {
		defer func() { done <- true }()
		_, err := io.Copy(dst, rw.reader)
		if err != nil {
			c.setError(emperror.Wrap(err, "cannot copy to target destination"))
			return
		}
	}()

	go func() {
		// close all writers at the end
		defer func() {
			for _, rw := range c.rws {
				rw.writer.Close()
			}
		}()
		// create list of writer
		writers := []io.Writer{}
		for _, rw := range c.rws {
			writers = append(writers, rw.writer)
		}

		mw := io.MultiWriter(writers...)

		if _, err := io.Copy(mw, src); err != nil {
			c.setError(emperror.Wrap(err, "cannot write to destination"))
		}
	}()

	// wait until all checksums and destination done
	for cnt := 0; cnt < len(c.rws); cnt++ {
		<-done
	}

	// do error handling
	if len(c.errors) > 0 {
		var e error = nil
		for _, err := range c.errors {
			if err == nil {
				e = err
			} else {
				e = emperror.Wrapf(e, "error: %v", err)
			}
		}
		return nil, e
	}

	return c.cs, nil
}
