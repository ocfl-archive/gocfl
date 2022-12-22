package osfs

import (
	"bytes"
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
	//fs     fs.FS
}

func (osFS *FS) Delete(name string) error {
	filename := filepath.ToSlash(filepath.Join(osFS.folder, name))
	return os.Remove(filename)
}

func NewFS(folder string, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	folder = strings.Trim(filepath.ToSlash(filepath.Clean(folder)), "/")
	osfs := &FS{
		folder: folder,
		//fs:     os.DirFS(folder),
		logger: logger,
	}
	return osfs, nil
}

func (osFS *FS) String() string {
	return fmt.Sprintf("file://%s", osFS.folder)
}

func (osFS *FS) IsNotExist(err error) bool {
	err = errors.Cause(err)
	return os.IsNotExist(err) || err == syscall.ENOENT
}

func (osFS *FS) Close() error {
	osFS.logger.Debug("Close OSFS ('%s')", osFS.folder)
	return nil
}

func (osFS *FS) Discard() error {
	osFS.logger.Debug("Discard OSFS")

	return nil
}

func (osFS *FS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(osFS.folder, name)
	osFS.logger.Debugf("opening %s", fullpath)
	file, err := os.Open(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", fullpath)
	}
	return file, nil
}

func (osFS *FS) Open(name string) (fs.File, error) {
	return osFS.OpenSeeker(name)
}

func (osFS *FS) ReadFile(name string) ([]byte, error) {
	fp, err := osFS.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", name)
	}
	defer fp.Close()
	data := bytes.NewBuffer(nil)
	if _, err := io.Copy(data, fp); err != nil {
		return nil, errors.Wrapf(err, "cannot read '%s'", name)
	}
	return data.Bytes(), nil
}

func (osFS *FS) Create(name string) (io.WriteCloser, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(osFS.folder, name)
	osFS.logger.Debugf("creating %s", fullpath)
	dir := filepath.Dir(fullpath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, errors.Wrapf(err, "cannot create folder '%s'", dir)
	}
	file, err := os.Create(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create %s", fullpath)
	}
	return file, nil
}

func (osFS *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(osFS.folder, name)
	osFS.logger.Debugf("reading entries of %s", fullpath)
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

func (osFS *FS) HasContent() bool {
	f, err := os.Open(osFS.folder)
	if err != nil {
		return false
	}
	defer f.Close()

	names, err := f.Readdirnames(3) // Or f.Readdir(1)
	if err != nil {
		return false
	}
	var hasContent bool
	for _, name := range names {
		if name == "." || name == ".." {
			continue
		}
		hasContent = true
		break
	}
	return hasContent
}

func (osFS *FS) WalkDir(root string, fn fs.WalkDirFunc) error {
	basepath := filepath.Join(osFS.folder, root)
	lb := len(osFS.folder)
	return filepath.WalkDir(basepath, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		/*
			if d.IsDir() {
				return nil
			}
		*/
		if len(path) <= lb {
			return errors.Errorf("path '%s' not a subpath of '%s'", path, basepath)
		}
		path = path[lb+1:]
		return fn(filepath.ToSlash(path), d, err)
	})
}

func (osFS *FS) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(osFS.folder, name)
	osFS.logger.Debugf("stat %s", fullpath)

	fi, err := os.Stat(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot stat %s", fullpath)
	}
	return fi, nil
}

func (osFS *FS) SubFS(name string) (ocfl.OCFLFSRead, error) {
	if name == "" || name == "." || name == "./" {
		return osFS, nil
	}
	return NewFS(filepath.Join(osFS.folder, name), osFS.logger)
}

func (osFS *FS) SubFSRW(name string) (ocfl.OCFLFS, error) {
	if name == "" || name == "." || name == "./" {
		return osFS, nil
	}
	return NewFS(filepath.Join(osFS.folder, name), osFS.logger)
}

// check interface satisfaction
var (
	_ ocfl.OCFLFS = &FS{}
)
