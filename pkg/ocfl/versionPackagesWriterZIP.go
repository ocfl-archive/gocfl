package ocfl

import (
	"archive/zip"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
)

func newSplitWriter(partSize int64, filename string, baseFS fs.FS, digestAlgorithm checksum.DigestAlgorithm) (*splitWriter, error) {
	s := &splitWriter{
		partSize:        partSize,
		partNo:          0,
		filename:        filename,
		baseFS:          baseFS,
		fileDigest:      map[string]string{},
		digestAlgorithm: digestAlgorithm,
	}
	if err := s.nextPart(); err != nil {
		return nil, errors.Wrapf(err, "cannot create first part '%s'", filename)
	}
	return s, nil
}

type splitWriter struct {
	partSize        int64
	pos             int64
	partNo          int
	filename        string
	baseFS          fs.FS
	writer          io.WriteCloser
	csWriter        *checksum.ChecksumWriter
	fileDigest      map[string]string
	digestAlgorithm checksum.DigestAlgorithm
}

const partNameFormat = "%s.%03d"

func (s *splitWriter) nextPart() error {
	var err error
	if s.writer != nil {
		if s.csWriter != nil {
			if err := s.csWriter.Close(); err != nil {
				return errors.Wrapf(err, "cannot close checksum writer for part %d", s.partNo)
			}
			partName := fmt.Sprintf(partNameFormat, s.filename, s.partNo)
			checksums, err := s.csWriter.GetChecksums()
			if err != nil {
				return errors.Wrapf(err, "cannot get checksums for part %d", s.partNo)
			}
			s.fileDigest[partName] = checksums[s.digestAlgorithm]
			s.csWriter = nil
		}
		if err := s.writer.Close(); err != nil {
			return errors.Wrapf(err, "cannot close part %d", s.partNo)
		}
	}
	s.partNo++
	s.pos = 0
	partName := s.filename
	if s.partSize > 0 {
		partName = fmt.Sprintf(partNameFormat, s.filename, s.partNo)
	}
	s.writer, err = writefs.Create(s.baseFS, partName)
	if err != nil {
		return errors.Wrapf(err, "cannot create part %d", s.partNo)
	}
	s.csWriter, err = checksum.NewChecksumWriter([]checksum.DigestAlgorithm{s.digestAlgorithm}, s.writer)
	if err != nil {
		defer s.writer.Close() // ensure we close the writer even on error
		if closeErr := s.writer.Close(); closeErr != nil {
			return errors.Wrapf(closeErr, "cannot close part %d after error", s.partNo)
		}
		return errors.Wrapf(err, "cannot create checksum writer for part %d", s.partNo)
	}
	return nil
}

func (s *splitWriter) Write(p []byte) (written int, err error) {
	rest := int64(len(p))
	for rest > 0 {
		size := rest
		if s.partSize > 0 {
			size = s.partSize - s.pos
		}
		if size > 0 {
			if size > rest {
				size = rest
			}
			n, err := s.csWriter.Write(p[:size])
			if err != nil {
				return n, errors.Wrapf(err, "cannot write to part %d", s.partNo)
			}
			written += n
			s.pos += int64(n)
			rest -= int64(n)
		} else {
			// we need to create a new part
			if err := s.nextPart(); err != nil {
				return 0, errors.Wrapf(err, "cannot create next part %d", s.partNo)
			}
		}
	}
	return written, nil
}

func (s *splitWriter) Close() error {
	if s.writer == nil {
		return nil
	}
	if err := s.csWriter.Close(); err != nil {
		return errors.Wrapf(err, "cannot close checksum writer for part %d", s.partNo)
	}
	checksums, err := s.csWriter.GetChecksums()
	if err != nil {
		return errors.Wrapf(err, "cannot get checksums for part %d", s.partNo)
	}
	s.csWriter = nil

	if err := s.writer.Close(); err != nil {
		return errors.Wrapf(err, "cannot close part %d", s.partNo)
	}
	s.writer = nil
	partName := s.filename
	if s.partSize > 0 {
		partName = fmt.Sprintf(partNameFormat, s.filename, s.partNo)
	}
	s.fileDigest[partName] = checksums[s.digestAlgorithm]
	return nil
}

var _ io.WriteCloser = (*splitWriter)(nil)

func newVersionPackagesWriterZIP(object *ObjectBase, version string, partSize int64, compress bool) (*VersionPackagesWriterZIP, error) {
	writer, err := newSplitWriter(partSize, fmt.Sprintf("%s.zip", version), object.fsys, object.GetDigestAlgorithm())
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zip writer for version '%s'", version)
	}
	zipWriter := zip.NewWriter(writer)
	return &VersionPackagesWriterZIP{
		ObjectBase: object,
		version:    version,
		writer:     writer,
		zipWriter:  zipWriter,
		compress:   compress,
		fileDigest: map[string]string{},
	}, nil
}

type VersionPackagesWriterZIP struct {
	*ObjectBase
	version    string
	writer     *splitWriter
	zipWriter  *zip.Writer
	compress   bool
	fileDigest map[string]string
}

func (version *VersionPackagesWriterZIP) GetFileDigest() (map[string]string, error) {
	if version == nil {
		return map[string]string{}, nil
	}
	return version.fileDigest, nil
}

func (version *VersionPackagesWriterZIP) WriteFile(name string, r io.Reader) (int64, error) {
	header := &zip.FileHeader{
		Name: name,
	}
	if version.compress {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}
	writer, err := version.zipWriter.CreateHeader(header)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot create '%s'", name)
	}
	num, err := io.Copy(writer, r)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot write file '%s' for version '%s'", name, version.version)
	}
	return num, nil
}

func (version *VersionPackagesWriterZIP) Version() string {
	return "ZIP64"
}

func (version *VersionPackagesWriterZIP) GetObject() *ObjectBase {
	return version.ObjectBase
}

func (version *VersionPackagesWriterZIP) Type() VersionPackagesType {
	return VersionZIP
}

func (version *VersionPackagesWriterZIP) Close() error {
	if version.writer == nil {
		return nil
	}
	var errs = []error{}
	if version.zipWriter != nil {
		if err := version.zipWriter.Close(); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot close zip writer for version '%s'", version.version))
		}
		version.zipWriter = nil
	}
	if err := version.writer.Close(); err != nil {
		errs = append(errs, errors.Wrapf(err, "cannot close writer for version '%s'", version.version))
	}
	version.fileDigest = version.writer.fileDigest
	version.writer = nil
	return errors.Combine(errs...)
}

func (version *VersionPackagesWriterZIP) addReader(r io.ReadCloser, names *NamesStruct, noExtensionHook bool) (string, error) {

	object := version.GetObject()

	writer, err := version.zipWriter.Create(names.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "cannot create '%s'", names.ManifestPath)
	}

	digest, err := object.addReader(r, writer, names, noExtensionHook)
	if err != nil {
		return "", errors.Wrapf(err, "cannot addReader() for '%s'", names.ManifestPath)
	}

	return digest, nil
}

var _ VersionPackageWriter = (*VersionPackagesWriterZIP)(nil)
