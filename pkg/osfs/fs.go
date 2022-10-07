package osfs

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FS struct {
	folder string
	logger *logging.Logger
	fs     fs.FS
}

func NewFSIO(folder string, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	folder = strings.Trim(filepath.ToSlash(filepath.Clean(folder)), "/")
	osfs := &FS{
		folder: folder,
		fs:     os.DirFS(folder),
		logger: logger,
	}
	return osfs, nil
}

func (ofs *FS) Close() error {
	ofs.logger.Debug("Close OSFS")
	return nil
}

func (ofs *FS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("opening %s", fullpath)
	file, err := os.Open(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", fullpath)
	}
	return file, nil
}

func (ofs *FS) Create(name string) (io.WriteCloser, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("creating %s", fullpath)
	file, err := os.Create(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create %s", fullpath)
	}
	return file, nil
}

func (ofs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("reading entries of %s", fullpath)
	dentries, err := os.ReadDir(fullpath)
	if os.IsNotExist(err) {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read folder %s", fullpath)
	}
	result := []os.DirEntry{}
	// get rid of pseudo dirs
	for _, dentry := range dentries {
		if dentry.Name() == "." || dentry.Name() == ".." {
			continue
		}
		result = append(result, dentry)
	}
	// sort on filename
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}

func (ofs *FS) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("stat %s", fullpath)

	fi, err := os.Stat(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot stat %s", fullpath)
	}
	return fi, nil
}

func (ofs *FS) SubFS(name string) ocfl.OCFLFS {
	if name == "." {
		name = ""
	}
	// error not possible, since base-folder is ok
	sfs, _ := NewFSIO(filepath.Join(ofs.folder, name), ofs.logger)
	return sfs
}
