package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
)

func NewVersionReaderPlain(version string, fsys fs.FS, logger zLogger.ZLogger) (*VersionReaderPlain, error) {
	if version == "" {
		return nil, errors.New("version cannot be empty")
	}
	if fsys == nil {
		return nil, errors.New("filesystem cannot be nil")
	}
	return &VersionReaderPlain{
		version: version,
		fsys:    fsys,
		logger:  logger,
	}, nil
}

type VersionReaderPlain struct {
	version string
	fsys    fs.FS
	logger  zLogger.ZLogger
}

func (v *VersionReaderPlain) GetFS() (fs.FS, io.Closer, error) {
	fs, err := fs.Sub(v.fsys, v.version)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot get sub filesystem for version %s", v.version)
	}
	return fs, io.NopCloser(nil), nil
}

func (v *VersionReaderPlain) GetFilenameChecksum(digestAlgorithm checksum.DigestAlgorithm, fixityAlgorithms []checksum.DigestAlgorithm, fn gfcCallback) (map[string]map[checksum.DigestAlgorithm]string, map[string]string, error) {
	var contentChecksums = make(map[string]map[checksum.DigestAlgorithm]string)

	var digestAlgs = append(fixityAlgorithms, digestAlgorithm)

	root := v.version
	fs.WalkDir(v.fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error walking directory %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			v.logger.Debug().Msgf("skipping non-regular file %s", path)
			return nil // skip non-regular files
		}
		v.logger.Debug().Msgf("processing file %s in version %s", path, v.version)
		fp, err := v.fsys.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot open file %s", path)
		}
		defer fp.Close()
		path = strings.TrimPrefix(path, root+"/")

		var contentWriter = []io.Writer{}
		pr, pw := io.Pipe()
		contentWriter = append(contentWriter, pw)
		checksumWriter, err := checksum.NewChecksumWriter(digestAlgs, contentWriter...)
		if err != nil {
			return errors.Wrapf(err, "cannot create checksum writer for file %s", path)
		}
		wg := sync.WaitGroup{}
		var errs = []error{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(path, pr); err != nil {
				errs = append(errs, errors.Wrapf(err, "error processing file %s in version %s", path, v.version))
			}
			io.Copy(io.Discard, pr) // drain the pipe to avoid deadlock
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := io.Copy(checksumWriter, fp); err != nil {
				errs = append(errs, errors.Wrapf(err, "error copying file %s in version %s", path, v.version))
			}
			if err := checksumWriter.Close(); err != nil {
				errs = append(errs, errors.Wrapf(err, "error closing checksum writer for file %s in version %s", path, v.version))
			}
		}()
		wg.Wait()
		if err := pw.Close(); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot close pipe writer for file %s in version %s", path, v.version))
		}
		if len(errs) > 0 {
			return errors.Wrapf(errors.Combine(errs...), "error processing file %s in version %s", path, v.version)
		}
		checksums, err := checksumWriter.GetChecksums()
		if err != nil {
			return errors.Wrapf(err, "cannot get checksums for file %s", path)
		}
		contentChecksums[path] = make(map[checksum.DigestAlgorithm]string)
		for _, alg := range fixityAlgorithms {
			if cs, ok := checksums[alg]; ok {
				contentChecksums[path][alg] = cs
			} else {
				return errors.Errorf("checksum for algorithm %s not found for file %s", alg, path)
			}
		}
		return nil
	})
	return contentChecksums, nil, nil
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
