package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/helper"
	"io"
	"io/fs"
	"strings"
)

func NewMultiZIPFS(fsys fs.FS, names []string, logger zLogger.ZLogger) (fs.FS, io.Closer, error) {
	if len(names) == 0 {
		return nil, nil, errors.New("no files provided for ZIPFSReaderAt")
	}
	mpr, err := helper.NewMultipartFileReaderAt(fsys, names)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot create MultipartFileReaderAt")
	}
	zfs, err := zipfs.NewFS(mpr, mpr.GetSize(), strings.Join(names, ", "), logger)
	if err != nil {
		mpr.Close()
		return nil, nil, errors.Wrap(err, "cannot create ZIPFS from MultipartFileReaderAt")
	}
	return zfs, mpr, nil
}
