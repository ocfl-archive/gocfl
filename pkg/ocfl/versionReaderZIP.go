package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io"
	"io/fs"
	"strings"
)

func NewVersionReaderZIP(version string, fsys fs.FS, names []string, logger zLogger.ZLogger) (VersionReader, error) {
	if version == "" {
		return nil, errors.New("version must not be empty")
	}
	if fsys == nil {
		return nil, errors.New("filesystem must not be nil")
	}
	if len(names) == 0 {
		return nil, errors.New("names must not be empty")
	}
	for _, name := range names {
		if name == "" {
			return nil, errors.New("name in names must not be empty")
		}
	}
	return &VersionReaderZIP{
		version: version,
		fsys:    fsys,
		names:   names,
		logger:  logger,
	}, nil
}

type VersionReaderZIP struct {
	version string
	fsys    fs.FS
	names   []string
	logger  zLogger.ZLogger
}

func (v *VersionReaderZIP) GetFS() (fs.FS, io.Closer, error) {
	fsys, closer, err := NewMultiZIPFS(v.fsys, v.names, v.logger)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot create filesystem for version %s", v.version)
	}
	fsys2, err := fs.Sub(fsys, v.version)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot create sub-filesystem for version %s", v.version)
	}
	return fsys2, closer, nil
}

func (v *VersionReaderZIP) GetFilenameChecksum(digestAlgs []checksum.DigestAlgorithm, fullContentFiles []string) (map[string]map[checksum.DigestAlgorithm]string, map[string][]byte, error) {
	var contentChecksums = make(map[string]map[checksum.DigestAlgorithm]string)
	var fullContent = make(map[string][]byte)
	fsys, closer, err := v.GetFS()
	if err != nil {
		return nil, fullContent, errors.Wrapf(err, "cannot get filesystem for version %s", v.version)
	}
	defer closer.Close()

	root := "content"
	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
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
		checksumsWriter, err := checksum.NewChecksumWriter(digestAlgs)
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

		return nil
	}); err != nil {
		return nil, fullContent, errors.Wrapf(err, "error walking content directory %s", root)
	}
	return contentChecksums, fullContent, nil
}

func (v *VersionReaderZIP) GetContentFilename() ([]string, error) {
	var filenames = []string{}

	fsys, closer, err := v.GetFS()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get filesystem for version %s", v.version)
	}
	defer closer.Close()

	root := "content"
	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
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
	}); err != nil {
		return nil, errors.Wrapf(err, "error walking content directory %s", root)
	}
	return filenames, nil
}

func (v *VersionReaderZIP) GetVersion() string {
	return v.version
}

var _ VersionReader = &VersionReaderZIP{}
