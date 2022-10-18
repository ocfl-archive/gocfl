package zipfs

import (
	"archive/zip"
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

type nopCloserWriter struct {
	io.Writer
}

func (*nopCloserWriter) Close() error { return nil }

type FS struct {
	srcReader io.ReaderAt
	dstWriter io.Writer
	r         *zip.Reader
	w         *zip.Writer
	newFiles  []string
	logger    *logging.Logger
}

func NewFSIO(src io.ReaderAt, srcSize int64, dst io.Writer, logger *logging.Logger) (*FS, error) {
	logger.Debug("instantiating FS")
	var err error
	zfs := &FS{
		newFiles:  []string{},
		srcReader: src,
		dstWriter: dst,
		logger:    logger,
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

func (zf *FS) String() string {
	return "zipfs://"
}
func (zf *FS) Close() error {
	zf.logger.Debug("Close ZipFS")
	// check whether we have to copy all stuff
	if zf.r != nil && zf.w != nil {
		// check whether there's a new version of the file
		for _, zipItem := range zf.r.File {
			found := false
			for _, added := range zf.newFiles {
				if added == zipItem.Name {
					found = true
					zf.logger.Debugf("overwriting %s", added)
					break
				}
			}
			if found {
				continue
			}
			zf.logger.Debugf("copying %s", zipItem.Name)
			zipItemReader, err := zipItem.OpenRaw()
			if err != nil {
				return errors.Wrapf(err, "cannot open raw source %s", zipItem.Name)
			}
			header := zipItem.FileHeader
			targetItem, err := zf.w.CreateRaw(&header)
			if err != nil {
				return errors.Wrapf(err, "cannot create raw target %s", zipItem.Name)
			}
			if _, err := io.Copy(targetItem, zipItemReader); err != nil {
				return errors.Wrapf(err, "cannot raw copy %s", zipItem.Name)
			}
		}
	}
	finalError := []error{}
	if zf.w != nil {
		if err := zf.w.Flush(); err != nil {
			finalError = append(finalError, err)
		}
		if err := zf.w.Close(); err != nil {
			finalError = append(finalError, err)
		}
	}
	return errors.Combine(finalError...)
}

func (zfs *FS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(name, "./")
	if zfs.r == nil {
		return nil, fs.ErrNotExist
	}
	name = filepath.ToSlash(name)
	zfs.logger.Debugf("%s", name)
	// check whether file is newly created
	for _, newItem := range zfs.newFiles {
		if newItem == name {
			return nil, fs.ErrInvalid // new files cannot be opened
		}
	}
	for _, zipItem := range zfs.r.File {
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
	zfs.logger.Debugf("%s not found", name)
	return nil, fs.ErrNotExist
}

func (zfs *FS) Create(name string) (io.WriteCloser, error) {
	zfs.logger.Debugf("%s", name)
	wc, err := zfs.w.Create(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create file %s", name)
	}
	zfs.newFiles = append(zfs.newFiles, name)
	return &nopCloserWriter{wc}, nil
}

func (zf *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	zf.logger.Debugf("%s", name)
	if zf.r == nil {
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
	for _, zipItem := range zf.r.File {
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
			fi, err := NewFileInfoFile(zipItem)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create FileInfo for %s", zipItem.Name)
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

func (zf *FS) Stat(name string) (fs.FileInfo, error) {
	name = filepath.ToSlash(name)
	zf.logger.Debugf("%s", name)

	// check whether file is newly created
	for _, newItem := range zf.newFiles {
		if newItem == name {
			return nil, fs.ErrInvalid // new files cannot be opened
		}
	}
	for _, zipItem := range zf.r.File {
		if zipItem.Name == name {
			finfo, err := NewFileInfoFile(zipItem)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create zipfs.FileInfo for %s", zipItem.Name)
			}
			return finfo, nil
		}
	}
	return nil, fs.ErrNotExist
}

func (zf *FS) SubFS(name string) ocfl.OCFLFS {
	if name == "." {
		name = ""
	}
	return &SubFS{
		FS:         zf,
		pathPrefix: filepath.ToSlash(filepath.Clean(name)),
	}
}
