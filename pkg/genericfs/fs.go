package genericfs

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

type FS struct {
	folder string
	logger *logging.Logger
	fs     fs.FS
}

func NewGenericFS(fs fs.FS, folder string, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	folder = strings.Trim(filepath.ToSlash(filepath.Clean(folder)), "/")
	osfs := &FS{
		folder: folder,
		fs:     os.DirFS(folder),
		logger: logger,
	}
	return osfs, nil
}

func (ofs *FS) String() string {
	return fmt.Sprintf("file://%s", ofs.folder)
}

func (ofs *FS) IsNotExist(err error) bool {
	err = errors.Cause(err)
	return os.IsNotExist(err) || err == syscall.ENOENT || err == fs.ErrNotExist
}

func (ofs *FS) Close() error {
	ofs.logger.Debug("Close OSFS")
	return nil
}

func (ofs *FS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("opening %s", fullpath)
	file, err := ofs.fs.Open(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", fullpath)
	}
	return file, nil
}

func (ofs *FS) Create(name string) (io.WriteCloser, error) {
	return nil, errors.Wrap(fs.ErrPermission, "read only filesystem")
}

func (ofs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("reading entries of %s", fullpath)
	dentries, err := fs.ReadDir(ofs.fs, fullpath)
	if os.IsNotExist(err) {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read folder %s", fullpath)
	}
	result := []fs.DirEntry{}
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

func (ofs *FS) WalkDir(root string, fn fs.WalkDirFunc) error {
	basepath := filepath.Join(ofs.folder, root)
	lb := len(ofs.folder)
	return fs.WalkDir(ofs.fs, basepath, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		/*
			if d.IsDir() {
				return nil
			}
		*/
		if len(path) <= lb {
			return errors.Errorf("path \"%s\" not a subpath of \"%s\"", path, basepath)
		}
		path = path[lb+1:]
		return fn(filepath.ToSlash(path), d, err)
	})
}

func (ofs *FS) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(ofs.folder, name)
	ofs.logger.Debugf("stat %s", fullpath)

	fi, err := fs.Stat(ofs.fs, fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot stat %s", fullpath)
	}
	return fi, nil
}

func (ofs *FS) SubFS(name string) ocfl.OCFLFS {
	if name == "." {
		name = ""
	}
	if name == "" {
		return ofs
	}
	// error not possible, since base-folder is ok
	sfs, _ := NewGenericFS(ofs.fs, filepath.Join(ofs.folder, name), ofs.logger)
	return sfs
}