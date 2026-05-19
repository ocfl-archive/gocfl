package objectimpl

import (
	"context"
	"io"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/filesystem/pkg/zipfsw"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type VersionWriter20Config struct{}

func NewVersionWriter20(ctx context.Context, fact factory.FactoryObject, conf any, logger ocfllogger.OCFLLogger) object.VersionWriter {
	versionWriterConfig, ok := conf.(*VersionWriter20Config)
	if conf != nil && !ok {
		logger.Error().Msg("invalid config type for versionWriter")
	}
	vw := &versionWriter20{
		versionWriter: &versionWriter{
			logger:      logger,
			updateFiles: []string{},
		},
		config: versionWriterConfig,
	}
	vw.versionWriter.addReader = vw._addReader
	return vw
}

type versionWriter20 struct {
	*versionWriter
	config  *VersionWriter20Config
	writeFS fs.FS
	digest  string
}

func (versionWriter *versionWriter20) _addReader(r io.ReadCloser, names *object.NamesStruct, noExtensionHook bool) (string, error) {
	inv := versionWriter.obj.GetInventory()
	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inv.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}

	var digest string

	versionWriter.updateFiles = append(versionWriter.updateFiles, names.ExternalPaths...)

	if !slices.Contains(digestAlgorithms, inv.GetDigestAlgorithm()) {
		digestAlgorithms = append(digestAlgorithms, inv.GetDigestAlgorithm())
	}

	if versionWriter.writeFS == nil {
		return "", errors.New("object write FS is nil")
	}
	writer, err := writefs.Create(versionWriter.writeFS, names.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "cannot create '%s'", names.ManifestPath)
	}
	defer writer.Close()

	var checksums map[checksum.DigestAlgorithm]string
	if noExtensionHook {
		checksums, err = checksum.Copy(digestAlgorithms, r, writer)
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%v' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		pr, pw := io.Pipe()
		extErrors := make(chan error, 1)
		go func() {
			defer wg.Done()
			if err := versionWriter.obj.GetExtensionManager().StreamObject(versionWriter, pr, names.ExternalPaths, names.InternalPath); err != nil {
				extErrors <- err
			}
		}()
		checksums, err = checksum.Copy(digestAlgorithms, r, writer, pw)
		if err := pw.Close(); err != nil {
			versionWriter.logger.Error().Err(err).Msg("cannot close pipe writer")
		}
		wg.Wait()
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%s' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
		close(extErrors)
		select {
		case err, ok := <-extErrors:
			if ok {
				return "", errors.Wrapf(err, "error on StreamObject() extension hook for object '%s'", inv.GetID())
			}
		default:
		}
	}

	if digest == "" {
		var ok bool
		digest, ok = checksums[inv.GetDigestAlgorithm()]
		if !ok {
			return "", errors.Errorf("digest '%s' not generated", inv.GetDigestAlgorithm())
		}
	} else {
		checksums[inv.GetDigestAlgorithm()] = digest
	}
	if err := inv.AddFile(names.ExternalPaths, names.ManifestPath, checksums); err != nil {
		return "", errors.Wrapf(err, "cannot append '%v'/'%s' to inventory", names.ExternalPaths, names.InternalPath)
	}

	return digest, nil
}

func (versionWriter *versionWriter20) Init(msg string, name string, address string) error {
	//todo: the init function calls UpdateObjectBefore make sure, that the correct versionWriter is the parameter...
	if err := versionWriter.versionWriter.Init(msg, name, address); err != nil {
		return errors.Wrap(err, "cannot initialize version writer")
	}
	inv := versionWriter.obj.GetInventory()

	fp, err := writefs.Create(versionWriter.obj.GetWriteFS(), versionWriter.ver.String()+".zip")
	if err != nil {
		return errors.Wrapf(err, "cannot create version zip file '%s'", versionWriter.ver.String()+".zip")
	}
	versionWriter.writeFS, err = zipfsw.NewFS(
		fp,
		true,
		true,
		versionWriter.ver.String(),
		[]checksum.DigestAlgorithm{inv.GetDigestAlgorithm()},
		func(css map[checksum.DigestAlgorithm]string) error {
			versionWriter.digest = css[inv.GetDigestAlgorithm()]
			return nil
		},
		versionWriter.logger.Logger(),
	)
	if err != nil {
		return errors.Wrapf(err, "cannot create version zip file '%s'", versionWriter.ver.String()+".zip")
	}
	return nil
}

func (versionWriter *versionWriter20) Close() error {
	if err := versionWriter.versionWriter.Close(); err != nil {
		return errors.Wrap(err, "cannot close version writer")
	}
	if versionWriter.writeFS != nil {
		if closer, ok := versionWriter.writeFS.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return errors.Wrap(err, "cannot close version zip file")
			}
		}
	}
	if versionWriter.digest != "" {
		inv := versionWriter.obj.GetInventory()
		if _, err := writefs.WriteFile(
			versionWriter.writeFS,
			versionWriter.ver.String()+".zip."+strings.ToLower(inv.GetDigestAlgorithm().String()),
			[]byte(versionWriter.digest+" *"+versionWriter.ver.String()),
		); err != nil {
			return errors.Wrap(err, "cannot write version zip file checksum")
		}
	} else {
		return errors.New("versionWriter digest is empty")
	}
	return nil
}
