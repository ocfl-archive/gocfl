package ocfl

import (
	"bytes"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/krolaw/zipstream"
	"github.com/ocfl-archive/gocfl/v2/pkg/helper"
	"io"
	"io/fs"
	"slices"
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

func (v *VersionReaderZIP) _GetFilenameChecksum(digestAlgs []checksum.DigestAlgorithm, fullContentFiles []string) (map[string]map[checksum.DigestAlgorithm]string, map[string][]byte, error) {
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
func (v *VersionReaderZIP) GetFilenameChecksum(digestAlgorithm checksum.DigestAlgorithm, fixityAlgorithms []checksum.DigestAlgorithm, fullContentFiles []string) (map[string]map[checksum.DigestAlgorithm]string, map[string][]byte, map[string]string, error) {
	var contentChecksums = make(map[string]map[checksum.DigestAlgorithm]string)
	var fullContent = make(map[string][]byte)
	var partsChecksum = make(map[string]string)

	var digestAlgs = append(fixityAlgorithms, digestAlgorithm)

	mpr, err := helper.NewMultipartFileReader(v.fsys, v.names, digestAlgorithm)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "cannot create multipart file reader for version %s", v.version)
	}

	zipStream := zipstream.NewReader(mpr)
	for {
		meta, err := zipStream.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // end of zip stream
			}
			return nil, nil, nil, errors.Wrapf(err, "error reading next file in zip stream for version %s", v.version)
		}
		v.logger.Debug().Msgf("processing file %s in version %s", meta.Name, v.version)
		name := strings.TrimPrefix(meta.Name, v.version+"/")
		var writers []io.Writer
		var contentBuffer *bytes.Buffer
		if slices.Contains(fullContentFiles, name) {
			v.logger.Debug().Msgf("reading full content for file %s in version %s", name, v.version)
			if meta.UncompressedSize64 > 100*1024*1024 { // 100 MB
				return nil, nil, nil, errors.Errorf("file %s is too large for full content (%d bytes)", name, meta.UncompressedSize64)
			}
			contentBuffer = bytes.NewBuffer(nil)
			writers = append(writers, contentBuffer)
		}
		checksumWriter, err := checksum.NewChecksumWriter(digestAlgs, writers...)
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "cannot create checksum writer for file %s in version %s", meta.Name, v.version)
		}
		if _, err := io.Copy(checksumWriter, zipStream); err != nil {
			return nil, nil, nil, errors.Wrapf(err, "cannot copy file %s in version %s", meta.Name, v.version)
		}
		if err := checksumWriter.Close(); err != nil {
			return nil, nil, nil, errors.Wrapf(err, "cannot close checksum writer for file %s in version %s", meta.Name, v.version)
		}
		if contentBuffer != nil {
			fullContent[name] = contentBuffer.Bytes()
		}
		checksums, err := checksumWriter.GetChecksums()
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "cannot get checksums for file %s in version %s", meta.Name, v.version)
		}
		contentChecksums[name] = make(map[checksum.DigestAlgorithm]string)
		for alg, cs := range checksums {
			contentChecksums[name][alg] = cs
		}
	}
	partsChecksum = mpr.GetPartsChecksum()
	return contentChecksums, fullContent, partsChecksum, nil
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
