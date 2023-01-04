package concurrentWriter

import (
	"emperror.dev/errors"
	"io"
	"sync"
)

type WriterRunner interface {
	Do(reader io.Reader, done chan bool)
	GetName() string
}

type pipe struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

type ConcurrentWriter struct {
	errors    []error
	rwsWriter io.Writer
	rws       map[string]pipe
	dataLock  sync.Mutex
	done      chan bool
	open      bool
	runners   []WriterRunner
}

func NewConcurrentWriter(runners []WriterRunner, writer ...io.Writer) *ConcurrentWriter {
	c := &ConcurrentWriter{
		errors:   []error{},
		rws:      map[string]pipe{},
		dataLock: sync.Mutex{},
		done:     make(chan bool),
		open:     true,
		runners:  runners,
	}
	c.start(writer...)
	return c
}

func (c *ConcurrentWriter) start(writers ...io.Writer) {

	// create the map of all ChecksumCopy-pipes and start async process
	for _, runner := range c.runners {
		rw := pipe{}
		rw.reader, rw.writer = io.Pipe()
		c.rws[runner.GetName()] = rw
		var r = runner
		go r.Do(rw.reader, c.done)
	}

	if len(writers) > 0 {
		// all destinations
		var dst io.Writer
		if len(writers) > 1 {
			dst = io.MultiWriter(writers...)
		} else {
			dst = writers[0]
		}

		// target pipe
		rw := pipe{}
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
	}
	// create list of writer
	allWriters := []io.Writer{}
	for _, rw := range c.rws {
		allWriters = append(allWriters, rw.writer)
	}

	c.rwsWriter = io.MultiWriter(allWriters...)
}

func (c *ConcurrentWriter) setError(err error) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	c.errors = append(c.errors, err)
}

func (c *ConcurrentWriter) Write(p []byte) (n int, err error) {
	return c.rwsWriter.Write(p)
}

func (c *ConcurrentWriter) Close() error {
	if !c.open {
		return errors.New("writer already closed")
	}
	defer func() { c.open = false }()
	for key, rw := range c.rws {
		if err := rw.writer.Close(); err != nil {
			c.errors = append(c.errors, errors.Wrapf(err, "error closing pipe '%s'", key))
		}
	}

	// wait until all checksums and destination done
	for cnt := 0; cnt < len(c.rws); cnt++ {
		<-c.done
	}

	return errors.Combine(c.errors...)
}

var (
	_ io.WriteCloser = (*ConcurrentWriter)(nil)
)
