package zipfs

import (
	"archive/zip"
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
)

type nopCloserWriter struct {
	io.Writer
}

func (*nopCloserWriter) Close() error { return nil }

type FS struct {
	srcReader io.ReaderAt
	dstWriter io.Writer
	r         *zip.Reader
	w         *zip.Writer
	noCopy    []string
	logger    *logging.Logger
	closed    *bool
	close     func() error
}

func NewFS(src io.ReaderAt, srcSize int64, dst io.Writer, close func() error, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	var err error
	var isClosed bool
	zfs := &FS{
		noCopy:    []string{},
		srcReader: src,
		dstWriter: dst,
		logger:    logger,
		closed:    &isClosed,
		close:     close,
	}
	if src != nil && src != (*os.File)(nil) {
		if zfs.r, err = zip.NewReader(src, srcSize); err != nil {
			return nil, errors.Wrap(err, "cannot create zip reader")
		}
	}
	if dst != nil && dst != (*os.File)(nil) {
		zfs.w = zip.NewWriter(dst)
	}
	return zfs, nil
}

func (zipFS *FS) String() string {
	return fmt.Sprintf("zipfs://")
}

func (zipFS *FS) isClosed() bool {
	return *zipFS.closed
}

func (zipFS *FS) IsNotExist(err error) bool {
	return err == fs.ErrNotExist
}

func (zipFS *FS) Close() error {
	if zipFS.isClosed() {
		return nil
		//return errors.New("zipFS closed")
	}
	zipFS.logger.Debug("Close ZipFS")
	// check whether we have to copy all stuff
	if zipFS.r != nil && zipFS.w != nil {
		// check whether there's a new version of the file
		for _, zipItem := range zipFS.r.File {
			found := false
			for _, added := range zipFS.noCopy {
				if added == zipItem.Name {
					found = true
					zipFS.logger.Debugf("overwriting %s", added)
					break
				}
			}
			if found {
				continue
			}
			zipFS.logger.Debugf("copying %s", zipItem.Name)
			zipItemReader, err := zipItem.OpenRaw()
			if err != nil {
				return errors.Wrapf(err, "cannot open raw source %s", zipItem.Name)
			}
			header := zipItem.FileHeader
			targetItem, err := zipFS.w.CreateRaw(&header)
			if err != nil {
				return errors.Wrapf(err, "cannot create raw target %s", zipItem.Name)
			}
			if _, err := io.Copy(targetItem, zipItemReader); err != nil {
				return errors.Wrapf(err, "cannot raw copy %s", zipItem.Name)
			}
		}
	}
	finalError := []error{}
	if zipFS.w != nil {
		if err := zipFS.w.Flush(); err != nil {
			finalError = append(finalError, err)
		}
		if err := zipFS.w.Close(); err != nil {
			finalError = append(finalError, err)
		}

	}
	finalError = append(finalError, zipFS.close())
	return errors.Combine(finalError...)
}
func (zipFS *FS) Discard() error {
	finalError := []error{}
	if zipFS.w != nil {
		if err := zipFS.w.Flush(); err != nil {
			finalError = append(finalError, err)
		}
		if err := zipFS.w.Close(); err != nil {
			finalError = append(finalError, err)
		}
	}
	return errors.Combine(finalError...)
}

func (zipFS *FS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	if zipFS.isClosed() {
		return nil, errors.New("zipFS closed")
	}

	name = filepath.ToSlash(filepath.Clean(name))
	//name = strings.TrimPrefix(name, "./")
	if zipFS.r == nil {
		return nil, fs.ErrNotExist
	}
	name = filepath.ToSlash(name)
	zipFS.logger.Debugf("%s", name)
	// check whether file is newly created
	for _, newItem := range zipFS.noCopy {
		if newItem == name {
			return nil, fs.ErrInvalid // new files cannot be opened
		}
	}
	for _, zipItem := range zipFS.r.File {
		if zipItem.Name == name {
			finfo, err := NewFileInfoFile(zipItem)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create zipfs.FileInfo for %s", zipItem.Name)
			}
			f, err := NewFile(finfo)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create zipfs.File from zipfs.FileInfo for %s", finfo.Name())
			}
			return f, nil
		}
	}
	zipFS.logger.Debugf("%s not found", name)
	return nil, fs.ErrNotExist
}

func (zipFS *FS) Open(name string) (fs.File, error) {
	return zipFS.OpenSeeker(name)
}

func (zipFS *FS) ReadFile(name string) ([]byte, error) {
	fp, err := zipFS.Open(name)
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

func (zipFS *FS) Create(name string) (io.WriteCloser, error) {
	if zipFS.isClosed() {
		return nil, errors.New("zipFS closed")
	}
	name = filepath.ToSlash(filepath.Clean(name))
	zipFS.logger.Debugf("%s", name)
	wc, err := zipFS.w.Create(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create file %s", name)
	}
	zipFS.noCopy = append(zipFS.noCopy, name)
	return &nopCloserWriter{wc}, nil
}

func (zipFS *FS) Delete(name string) error {
	filename := filepath.ToSlash(filepath.Clean(name))
	zipFS.noCopy = append(zipFS.noCopy, filename)
	return nil
}

func (zipFS *FS) HasContent() bool {
	dirEntries, err := zipFS.ReadDir(".")
	if err != nil {
		return false
	}
	return len(dirEntries) > 0
}

func (zipFS *FS) ReadDir(path string) ([]fs.DirEntry, error) {
	if zipFS.isClosed() {
		return nil, errors.New("zipFS closed")
	}

	name := filepath.ToSlash(filepath.Clean(path))
	zipFS.logger.Debugf("%s", name)
	if zipFS.r == nil {
		return []fs.DirEntry{}, nil
	}

	if name == "." {
		name = ""
	}
	// force slash at the end
	if name != "" {
		name = strings.TrimSuffix(filepath.ToSlash(name), "/") + "/"
	}
	var entries = []*DirEntry{}
	var dirs = []string{}
	for _, zipItem := range zipFS.r.File {
		if name != "" && !strings.HasPrefix(zipItem.Name, name) {
			continue
		}
		fname := zipItem.Name
		if name != "" && !strings.HasPrefix(fname, name) {
			continue
		}
		fname = strings.TrimPrefix(fname, name)
		parts := strings.Split(fname, "/")
		// only files have one part
		if len(parts) == 1 {
			var fi *FileInfo
			var err error
			if zipItem.Name == name {
				continue
			}
			if zipItem.FileInfo().IsDir() {
				fi, err = NewFileInfoDir(strings.TrimLeft(zipItem.Name, name))
				if err != nil {
					return nil, errors.Wrapf(err, "cannot create FileInfo for %s", zipItem.Name)
				}
			} else {
				fi, err = NewFileInfoFile(zipItem)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot create FileInfo for %s", zipItem.Name)
				}
			}
			entries = append(entries, NewDirEntry(fi))
		} else {
			found := false
			for _, d := range dirs {
				if d == parts[0] {
					found = true
				}
			}
			if !found {
				dirs = append(dirs, parts[0])
			}
		}
	}
	for _, d := range dirs {
		fi, err := NewFileInfoDir(d)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create Fileinfo for %s", d)
		}
		entries = append(entries, NewDirEntry(fi))
	}

	var result = []fs.DirEntry{}
	for _, entry := range entries {
		result = append(result, entry)
	}
	// sort on filename
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result, nil
}

func (zipFS *FS) WalkDir(path string, fn fs.WalkDirFunc) error {
	if zipFS.isClosed() {
		return errors.New("zipFS closed")
	}
	path = filepath.ToSlash(filepath.Clean(path))
	name := path
	lr := len(name) + 1
	for _, file := range zipFS.r.File {
		if !strings.HasPrefix(file.Name, name) {
			continue
		}
		var fi *FileInfo
		var err error
		if file.FileInfo().IsDir() {
			fi, err = NewFileInfoDir(file.Name[lr:])
			if err != nil {
				return errors.Wrapf(err, "cannot create FileInfo for %s", file.Name)
			}
		} else {
			fi, err = NewFileInfoFile(file)
			if err != nil {
				return errors.Wrapf(err, "cannot create FileInfo for %s", file.Name)
			}
		}
		if err := fn(fmt.Sprintf("%s/%s", path, file.Name[lr:]), NewDirEntry(fi), nil); err != nil {
			return err
		}
	}
	return nil
}

func (zipFS *FS) Stat(path string) (fs.FileInfo, error) {
	if zipFS.r == nil {
		return nil, fs.ErrNotExist
	}
	if zipFS.isClosed() {
		return nil, errors.New("zipFS closed")
	}

	name := filepath.ToSlash(filepath.Clean(path))
	zipFS.logger.Debugf("%s", name)

	// check whether file is newly created
	for _, newItem := range zipFS.noCopy {
		if newItem == name {
			return nil, fs.ErrInvalid // new files cannot be opened
		}
	}
	for _, zipItem := range zipFS.r.File {
		if zipItem.Name == name {
			finfo, err := NewFileInfoFile(zipItem)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create zipfs.FileInfo for %s", zipItem.Name)
			}
			return finfo, nil
		} else {
			if strings.HasPrefix(zipItem.Name, name) {
				return NewFileInfoDir(name)
			}
		}
	}
	return nil, fs.ErrNotExist
}

func (zipFS *FS) SubFSRW(path string) (ocfl.OCFLFS, error) {
	return NewSubFS(zipFS, path)
}
func (zipFS *FS) SubFS(path string) (ocfl.OCFLFSRead, error) {
	return zipFS.SubFSRW(path)
}

// check interface satisfaction
var (
	_ ocfl.OCFLFS = &FS{}
)
