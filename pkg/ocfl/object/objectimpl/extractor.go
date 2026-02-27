package objectimpl

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

func NewExtractor(ctx context.Context, factory factory.Factory, logger ocfllogger.OCFLLogger) object.Extractor {
	return &extractor{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "extractor"),
	}
}

type extractor struct {
	object.Object
	sourceFS fs.FS
	objectFS streamfs.FS
	ctx      context.Context
	factory  factory.Factory
	logger   ocfllogger.OCFLLogger
}

func (ext *extractor) GetFileReader(pathStr string) (io.ReadCloser, int64, string, error) {
	fp, err := ext.sourceFS.Open(pathStr)
	if err != nil {
		return nil, 0, "", errors.Wrapf(err, "cannot open file %s", pathStr)
	}
	fi, err := fp.Stat()
	if err != nil {
		return nil, 0, "", errors.Wrapf(err, "cannot stat file %s", pathStr)
	}

	mimeReader, err := iou.NewMimeReader(fp)
	if err != nil {
		return nil, 0, "", errors.Wrapf(err, "cannot create mime reader for object %s - %s", ext.GetID(), pathStr)
	}
	contentType, err := mimeReader.DetectContentType()
	if err != nil {
		return nil, 0, "", errors.Wrapf(err, "cannot detect content type for object %s - %s", ext.GetID(), pathStr)
	}
	return fp, fi.Size(), contentType, nil
}

func (ext *extractor) GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error) {
	pathStr := filepath.ToSlash(filepath.Join("extensions", extensionName, path))
	return ext.GetFileReader(pathStr)
}

func (ext *extractor) Extract(version *inventory.VersionNumber, withManifest bool, area string) error {
	var manifest strings.Builder
	var err error
	var inv = ext.GetInventory()
	var digestAlg = inv.GetDigestAlgorithm()
	if err := inv.IterateFiles(version, func(internals, externals []string, digest string) error {
		for _, external := range externals {
			external, err = ext.GetExtensionManager().BuildObjectExtractPath(external, area)
			if err != nil {
				errCause := errors.Cause(err)
				if errors.Is(errCause, object.ExtensionObjectExtractPathWrongAreaError) {
					return nil
				}
				return errors.Wrapf(err, "cannot map path '%s'", external)
			}
			if err := func() error {
				if len(internals) == 0 {
					return errors.Errorf("no internal paths for '%v'", externals)
				}
				internal := internals[0]
				src, err := ext.sourceFS.Open(internal)
				if err != nil {
					return errors.Wrapf(err, "cannot open '%v/%s'", ext.sourceFS, internal)
				}
				defer src.Close()
				target, err := writefs.Create(ext.objectFS, external)
				if err != nil {
					return errors.Wrapf(err, "cannot create '%v/%s'", ext.objectFS, external)
				}
				defer target.Close()
				ext.logger.Debug().Msgf("writing '%v/%s' -> '%v/%s'", ext.sourceFS, internal, ext.objectFS, external)
				copyDigests, err := checksum.Copy([]checksum.DigestAlgorithm{digestAlg}, src, target)
				if err != nil {
					return errors.Wrapf(err, "error copying '%v/%s' -> '%v/%s'", ext.sourceFS, internal, ext.objectFS, external)
				}
				copyDigest, ok := copyDigests[digestAlg]
				if !ok {
					return errors.Errorf("no digest '%s' generated", digestAlg)
				}
				if copyDigest != digest {
					return errors.Errorf("invalid digest for '%s' - [%s] != [%s]", internal, copyDigests, digest)
				}
				return nil
			}(); err != nil {
				return err
			}
			if withManifest {
				manifest.WriteString(fmt.Sprintf("%s %s\n", digest, external))
			}
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot iterate external files")
	}
	if withManifest {
		manifestName := fmt.Sprintf("manifest.%s", digestAlg)
		fp, err := writefs.Create(ext.objectFS, manifestName)
		if err != nil {
			return errors.Wrapf(err, "cannot crate manifest file %v/%s", ext.objectFS, manifestName)
		}
		if _, err := io.WriteString(fp, manifest.String()); err != nil {
			return errors.Wrapf(err, "cannot write manifest file %v/%s", ext.objectFS, manifestName)
		}
		defer fp.Close()
	}
	ext.logger.Debug().Msgf("object '%s' extracted", inv.GetID())
	return nil

}

func (ext *extractor) WithObject(o object.Object) object.Extractor {
	ext.Object = o
	return ext
}

func (ext *extractor) WithFS(sourceFS fs.FS, objectFS streamfs.FS) object.Extractor {
	ext.sourceFS = sourceFS
	ext.objectFS = objectFS
	return ext
}

var _ object.Extractor = (*extractor)(nil)
