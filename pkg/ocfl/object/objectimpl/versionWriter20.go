package objectimpl

import (
	"context"
	"io"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
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
	writeFS appendfs.FS
	digest  string
}

func (versionWriter *versionWriter20) WithObject(obj object.Object) object.VersionWriter {
	versionWriter.obj = obj
	return versionWriter
}

func (versionWriter *versionWriter20) WithEcho(echo bool) object.VersionWriter {
	versionWriter.echo = echo
	return versionWriter
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
	writeFS, err := zipfsw.NewFS(
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

	versionWriter.writeFS = writeFS.(appendfs.FS)
	versionWriter.versionWriter.writeFS = versionWriter.writeFS.(appendfs.FS)
	return nil
}

func (versionWriter *versionWriter20) Close() error {
	if versionWriter.writeFS != nil {
		if closer, ok := versionWriter.writeFS.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return errors.Wrap(err, "cannot close version zip file")
			}
		}
	}
	if err := versionWriter.versionWriter.Close(); err != nil {
		return errors.Wrap(err, "cannot close version writer")
	}
	if versionWriter.digest != "" {
		inv := versionWriter.obj.GetInventory()
		if _, err := writefs.WriteFile(
			versionWriter.obj.GetWriteFS(),
			versionWriter.ver.String()+".zip."+strings.ToLower(inv.GetDigestAlgorithm().String()),
			[]byte(versionWriter.digest+" *"+versionWriter.ver.String()+".zip"),
		); err != nil {
			return errors.Wrap(err, "cannot write version zip file checksum")
		}
	} else {
		return errors.New("versionWriter digest is empty")
	}
	return nil
}
