package ocfl

import (
	"bytes"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
)

func NewVersionReaderPlain(version string, fsys fs.FS) (*VersionReaderPlain, error) {
	if version == "" {
		return nil, errors.New("version cannot be empty")
	}
	if fsys == nil {
		return nil, errors.New("filesystem cannot be nil")
	}
	return &VersionReaderPlain{
		version: version,
		fsys:    fsys,
	}, nil
}

type VersionReaderPlain struct {
	version string
	fsys    fs.FS
}

func (v *VersionReaderPlain) GetFS() (fs.FS, io.Closer, error) {
	fs, err := fs.Sub(v.fsys, v.version)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot get sub filesystem for version %s", v.version)
	}
	return fs, io.NopCloser(nil), nil
}

func (v *VersionReaderPlain) GetFilenameChecksum(digestAlgs []checksum.DigestAlgorithm, fullContentFiles []string) (map[string]map[checksum.DigestAlgorithm]string, map[string][]byte, error) {
	var contentChecksums = make(map[string]map[checksum.DigestAlgorithm]string)
	var fullContent = make(map[string][]byte)
	root := v.version
	fs.WalkDir(v.fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error walking directory %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil // skip non-regular files
		}
		fp, err := v.fsys.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot open file %s", path)
		}
		defer fp.Close()
		path = strings.TrimPrefix(path, root+"/")

		var contentWriter = []io.Writer{}
		var contentBuffer *bytes.Buffer
		if slices.Contains(fullContentFiles, path) {
			if fi, err := fp.Stat(); err != nil {
				if fi.Size() > 100*1024*1024 { // 100 MB
					return errors.Wrapf(err, "file %s is too large for full content checksum (%dbyte)", path, fi.Size())
				}
			}
			contentBuffer = bytes.NewBuffer(nil)
			contentWriter = append(contentWriter, contentBuffer)
		}

		checksumsWriter, err := checksum.NewChecksumWriter(digestAlgs, contentWriter...)
		if err != nil {
			return errors.Wrapf(err, "cannot create checksum writer for file %s", path)
		}
		if _, err := io.Copy(checksumsWriter, fp); err != nil {
			return errors.Wrapf(err, "cannot copy file %s", path)
		}
		if err := checksumsWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close checksum writer for file %s", path)
		}
		checksums, err := checksumsWriter.GetChecksums()
		if err != nil {
			return errors.Wrapf(err, "cannot get checksums for file %s", path)
		}
		contentChecksums[path] = make(map[checksum.DigestAlgorithm]string)
		for _, alg := range digestAlgs {
			if cs, ok := checksums[alg]; ok {
				contentChecksums[path][alg] = cs
			} else {
				return errors.Errorf("checksum for algorithm %s not found for file %s", alg, path)
			}
		}
		if contentBuffer != nil {
			fullContent[path] = contentBuffer.Bytes()
		}
		return nil
	})
	return contentChecksums, fullContent, nil
}

func (v *VersionReaderPlain) GetContentFilename() ([]string, error) {
	var filenames = []string{}
	root := path.Join(v.version, "content")
	fs.WalkDir(v.fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error walking directory %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil // skip non-regular files
		}
		filenames = append(filenames, strings.TrimLeft(path, root+"/"))
		return nil
	})
	return filenames, nil
}

func (v *VersionReaderPlain) GetVersion() string {
	return v.version
}

var _ VersionReader = &VersionReaderPlain{}
