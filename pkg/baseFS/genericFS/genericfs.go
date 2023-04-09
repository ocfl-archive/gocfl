package genericfs

import (
	"bytes"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/op/go-logging"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

func NewFS(fsys fs.FS, folder string, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	folder = strings.Trim(filepath.ToSlash(filepath.Clean(folder)), "/")
	osfs := &FS{
		folder: folder,
		fs:     fsys,
		logger: logger,
	}
	return osfs, nil
}

type FS struct {
	folder string
	logger *logging.Logger
	fs     fs.FS
}

func (genericFS *FS) String() string {
	return fmt.Sprintf("file://%s", genericFS.folder)
}
func (genericFS *FS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	file, err := genericFS.Open(name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	rsc, ok := file.(ocfl.FileSeeker)
	if !ok {
		return nil, errors.Errorf("no ReadSeekCloser for '%s'", name)
	}
	return rsc, nil
}

func (genericFS *FS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.ToSlash(filepath.Join(genericFS.folder, name))
	genericFS.logger.Debugf("opening %s", fullpath)
	file, err := genericFS.fs.Open(fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", fullpath)
	}
	return file, nil
}

func (genericFS *FS) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(genericFS.folder, name)
	genericFS.logger.Debugf("stat %s", fullpath)

	fi, err := fs.Stat(genericFS.fs, fullpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot stat %s", fullpath)
	}
	return fi, nil
}
func (genericFS *FS) ReadFile(name string) ([]byte, error) {
	fp, err := genericFS.Open(name)
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
func (genericFS *FS) Close() error {
	genericFS.logger.Debug("Close OSFS")
	return nil
}
func (genericFS *FS) IsNotExist(err error) bool {
	err = errors.Cause(err)
	return os.IsNotExist(err) || err == syscall.ENOENT || err == fs.ErrNotExist
}
func (genericFS *FS) WalkDir(root string, fn fs.WalkDirFunc) error {
	basepath := filepath.Join(genericFS.folder, root)
	lb := len(genericFS.folder)
	return fs.WalkDir(genericFS.fs, basepath, func(path string, d fs.DirEntry, err error) error {
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
func (genericFS *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = strings.TrimPrefix(filepath.ToSlash(filepath.Clean(name)), "./")
	fullpath := filepath.Join(genericFS.folder, name)
	genericFS.logger.Debugf("reading entries of %s", fullpath)
	dentries, err := fs.ReadDir(genericFS.fs, fullpath)
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
func (genericFS *FS) HasContent() bool {
	dirEntries, err := genericFS.ReadDir(".")
	if err != nil {
		return false
	}
	var hasContent bool
	for _, de := range dirEntries {
		if de.Name() == "." || de.Name() == ".." {
			continue
		}
		hasContent = true
		break
	}
	return hasContent
}
func (genericFS *FS) SubFS(name string) (ocfl.OCFLFSRead, error) {
	if name == "." {
		name = ""
	}
	if name == "" {
		return genericFS, nil
	}
	return NewFS(genericFS.fs, filepath.ToSlash(filepath.Join(genericFS.folder, name)), genericFS.logger)
}

func (genericFS *FS) Sub(dir string) (fs.FS, error) {
	return genericFS.SubFS(dir)
}

// check interface satisfaction
var (
	_ fs.FS         = &FS{}
	_ fs.StatFS     = &FS{}
	_ fs.ReadFileFS = &FS{}
	_ fs.ReadDirFS  = &FS{}
	_ fs.SubFS      = &FS{}
)
