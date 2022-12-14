package s3fs

import (
	"emperror.dev/errors"
	"github.com/minio/minio-go/v7"
	"io"
)

type UploadInfo struct {
	uploadInfo minio.UploadInfo
	err        error
}

type WriteCloser struct {
	*io.PipeWriter
	pr *io.PipeReader
	c  chan UploadInfo
}

func (wc *WriteCloser) Close() error {
	errs := []error{}
	errs = append(errs, wc.pr.Close())
	errs = append(errs, wc.PipeWriter.Close())
	uploadInfo := <-wc.c
	errs = append(errs, uploadInfo.err)
	return errors.Combine(errs...)
}

func NewWriteCloser() *WriteCloser {
	pr, pw := io.Pipe()
	return &WriteCloser{
		PipeWriter: pw,
		pr:         pr,
		c:          make(chan UploadInfo),
	}
}
