package checksum

import (
	"emperror.dev/errors"
	"fmt"
	"io"
	"sync"
)

type rwStruct struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

type ChecksumWriter struct {
	//	sync.Mutex
	checksums []DigestAlgorithm
	cs        map[DigestAlgorithm]string
	errors    []error
	rws       map[DigestAlgorithm]rwStruct
	rwsWriter io.Writer
	dataLock  sync.Mutex
	done      chan bool
	open      bool
}

func NewChecksumWriter(checksums []DigestAlgorithm, writer ...io.Writer) *ChecksumWriter {
	c := &ChecksumWriter{
		//		Mutex:       sync.Mutex{},
		checksums: checksums,
		cs:        map[DigestAlgorithm]string{},
		errors:    []error{},
		rws:       map[DigestAlgorithm]rwStruct{},
		dataLock:  sync.Mutex{},
		done:      make(chan bool),
		open:      true,
	}
	c.start(writer...)
	return c
}

func (c *ChecksumWriter) start(writers ...io.Writer) {
	// all destinations
	dst := io.MultiWriter(writers...)

	// create the map of all ChecksumCopy-pipes and start async process
	for _, csType := range c.checksums {
		rw := rwStruct{}
		rw.reader, rw.writer = io.Pipe()
		c.rws[csType] = rw
		go func() {
			c.doChecksum(rw.reader, csType, c.done)
		}()
	}

	// target pipe
	rw := rwStruct{}
	rw.reader, rw.writer = io.Pipe()
	c.rws["_"] = rw
	go func() {
		defer func() { c.done <- true }()
		_, err := io.Copy(dst, rw.reader)
		if err != nil {
			c.setError(errors.Wrap(err, "cannot copy to target destination"))
			return
		}
	}()

	// create list of writer
	allWriters := []io.Writer{}
	for _, rw := range c.rws {
		allWriters = append(allWriters, rw.writer)
	}

	c.rwsWriter = io.MultiWriter(allWriters...)
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
		c.setError(errors.Wrapf(err, "cannot create checkum %s", csType))
		return
	}
	csString := fmt.Sprintf("%x", sink.Sum(nil))
	c.setResult(csType, csString)
}

func (c *ChecksumWriter) Write(p []byte) (n int, err error) {
	return c.rwsWriter.Write(p)
}

func (c *ChecksumWriter) Close() error {
	if !c.open {
		return errors.New("writer already closed")
	}
	// wait until all checksums and destination done
	for cnt := 0; cnt < len(c.rws); cnt++ {
		<-c.done
	}

	for key, rw := range c.rws {
		if err := rw.writer.Close(); err != nil {
			c.errors = append(c.errors, errors.Wrapf(err, "error closing pipe '%s'", key))
		}
	}

	return errors.Combine(c.errors...)
}

func (c *ChecksumWriter) GetChecksums() map[DigestAlgorithm]string {
	if c.open {
		return map[DigestAlgorithm]string{}
	}
	return c.cs
}

var (
	_ io.WriteCloser = (*ChecksumWriter)(nil)
)
