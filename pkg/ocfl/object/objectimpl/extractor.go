package objectimpl

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
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

func (extractor *extractor) Extract(version *inventory.VersionNumber, withManifest bool, area string) error {
	var manifest strings.Builder
	var err error
	var inv = extractor.GetInventory()
	var digestAlg = inv.GetDigestAlgorithm()
	if err := inv.IterateFiles(version, func(internals, externals []string, digest string) error {
		for _, external := range externals {
			external, err = extractor.GetExtensionManager().BuildObjectExtractPath(external, area)
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
				src, err := extractor.sourceFS.Open(internal)
				if err != nil {
					return errors.Wrapf(err, "cannot open '%v/%s'", extractor.sourceFS, internal)
				}
				defer src.Close()
				target, err := writefs.Create(extractor.objectFS, external)
				if err != nil {
					return errors.Wrapf(err, "cannot create '%v/%s'", extractor.objectFS, external)
				}
				defer target.Close()
				extractor.logger.Debug().Msgf("writing '%v/%s' -> '%v/%s'", extractor.sourceFS, internal, extractor.objectFS, external)
				copyDigests, err := checksum.Copy([]checksum.DigestAlgorithm{digestAlg}, src, target)
				if err != nil {
					return errors.Wrapf(err, "error copying '%v/%s' -> '%v/%s'", extractor.sourceFS, internal, extractor.objectFS, external)
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
		fp, err := writefs.Create(extractor.objectFS, manifestName)
		if err != nil {
			return errors.Wrapf(err, "cannot crate manifest file %v/%s", extractor.objectFS, manifestName)
		}
		if _, err := io.WriteString(fp, manifest.String()); err != nil {
			return errors.Wrapf(err, "cannot write manifest file %v/%s", extractor.objectFS, manifestName)
		}
		defer fp.Close()
	}
	extractor.logger.Debug().Msgf("object '%s' extracted", inv.GetID())
	return nil

}

func (extractor *extractor) WithObject(o object.Object) object.Extractor {
	extractor.Object = o
	return extractor
}

func (extractor *extractor) WithFS(sourceFS fs.FS, objectFS streamfs.FS) object.Extractor {
	extractor.sourceFS = sourceFS
	extractor.objectFS = objectFS
	return extractor
}

var _ object.Extractor = (*extractor)(nil)
