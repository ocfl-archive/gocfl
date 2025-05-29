package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"io"
)

func newVersionPackageWriterPlain(object *ObjectBase, version string) *VersionPackageWriterPlain {
	return &VersionPackageWriterPlain{
		ObjectBase: object,
		version:    version,
	}
}

type VersionPackageWriterPlain struct {
	*ObjectBase
	version string
}

func (version *VersionPackageWriterPlain) GetFileDigest() (map[string]string, error) {
	return map[string]string{}, nil
}

func (version *VersionPackageWriterPlain) WriteFile(name string, r io.Reader) (int64, error) {
	object := version.GetObject()
	writer, err := writefs.Create(object.fsys, name)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot create '%s'", name)
	}
	defer writer.Close()
	num, err := io.Copy(writer, r)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot write to '%s'", name)
	}
	return num, nil
}

func (version *VersionPackageWriterPlain) Version() string {
	return version.version
}

func (version *VersionPackageWriterPlain) GetObject() *ObjectBase {
	return version.ObjectBase
}

func (version *VersionPackageWriterPlain) Type() VersionPackagesType {
	return VersionPlain
}

func (version *VersionPackageWriterPlain) Close() error {
	return nil
}

func (version *VersionPackageWriterPlain) addReader(r io.ReadCloser, names *NamesStruct, noExtensionHook bool) (string, error) {

	object := version.GetObject()

	writer, err := writefs.Create(object.fsys, names.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "cannot create '%s'", names.ManifestPath)
	}
	defer writer.Close()

	digest, err := object.addReader(r, writer, names, noExtensionHook)

	return digest, nil
}

var _ VersionPackageWriter = (*VersionPackageWriterPlain)(nil)
