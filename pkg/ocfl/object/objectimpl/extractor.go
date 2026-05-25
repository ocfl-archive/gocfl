package objectimpl

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type ExtractorConfig struct{}

func NewExtractor(ctx context.Context, factory factory.FactoryObject, config any, logger ocfllogger.OCFLLogger) object.Extractor {
	extractorConfig, ok := config.(*ExtractorConfig)
	if config != nil && !ok {
		logger.Error().Msg("invalid config type for extractor")
	}
	ext := &extractor{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "extractor"),
		config:  extractorConfig,
	}
	ext.getVersionFS = ext._getVersionFS
	return ext
}

type extractor struct {
	object.Object
	destFS       appendfs.FS
	ctx          context.Context
	factory      factory.FactoryObject
	logger       ocfllogger.OCFLLogger
	config       *ExtractorConfig
	objectReadFS fs.FS
	getVersionFS func(string) (fs.FS, error)
}

func (ext *extractor) GetMetadata() (*inventory.Metadata, error) {
	inv := ext.GetInventory()
	if inv == nil {
		return nil, errors.Errorf("inventory is nil")
	}

	result := &inventory.Metadata{
		ID:              inv.GetID(),
		Head:            inv.GetHead(),
		Files:           map[string]*inventory.FileMetadata{},
		DigestAlgorithm: inv.GetDigestAlgorithm(),
		Versions:        map[string]*inventory.VersionMetadata{},
	}
	versions := inv.GetVersions()
	versionStrings := []string{}
	for v, ver := range versions.Iterate() {
		result.Versions[v.String()] = &inventory.VersionMetadata{
			Created: ver.GetCreated(),
			Message: ver.GetMessage(),
			Name:    ver.GetUser().GetName(),
			Address: ver.GetUser().GetAddress(),
		}
		versionStrings = append(versionStrings, v.String())
	}
	// sort version strings in ascending order
	slices.SortFunc(versionStrings, func(a, b string) int {
		a = strings.TrimPrefix(a, "v0")
		b = strings.TrimPrefix(b, "v0")
		ia, _ := strconv.Atoi(a)
		ib, _ := strconv.Atoi(b)
		return cmp.Compare(ia, ib)
	})
	extensionMetadata, err := ext.GetExtensionManager().GetMetadata(ext.GetReadFS(), ext)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get extension metadata for object '%s'", inv.GetID())
	}
	if objectMeta, ok := extensionMetadata[""]; ok {
		/*
			for key, val := range objectMeta {
				result.Extension[key] = val
			}

		*/
		result.Extension = objectMeta
	}
	manifest := inv.GetManifest()
	fixity := inv.GetFixity()
	for digest, fnames := range manifest.Iterate() {
		if len(fnames) == 0 {
			continue
		}
		fm := &inventory.FileMetadata{
			Checksums:    map[checksum.DigestAlgorithm]string{},
			InternalName: fnames,
			VersionName:  map[string][]string{},
			Extension:    map[string]any{},
		}
		fm.Checksums = fixity.Checksums(fnames[0])
		for v, ver := range versions.Iterate() {
			for d, externalNames := range ver.GetState().Iterate() {
				if digest == d {
					if _, ok := fm.VersionName[v.String()]; !ok {
						fm.VersionName[v.String()] = []string{}
					}
					fm.VersionName[v.String()] = append(fm.VersionName[v.String()], externalNames...)
					break
				}
			}
		}
		if emAny, ok := extensionMetadata[digest]; ok {
			if em, ok := emAny.(map[string]any); ok {
				fm.Extension = em
			}
		}
		result.Files[digest] = fm
	}
	return result, nil
}

func (ext *extractor) GetFileReader(pathStr string) (io.ReadCloser, int64, string, error) {
	fp, err := ext.GetReadFS().Open(pathStr)
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
	return mimeReader, fi.Size(), contentType, nil
}

func (ext *extractor) GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error) {
	pathStr := filepath.ToSlash(filepath.Join("extensions", extensionName, path))
	return ext.GetFileReader(pathStr)
}

func (ext *extractor) _getVersionFS(version string) (fs.FS, error) {
	readFS, err := fs.Sub(ext.objectReadFS, version)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create sub FS '%v'", version)
	}
	return readFS, nil
}

func (ext *extractor) Extract(version *inventory.VersionNumber, withManifest bool, area string) error {
	var err error
	var fsMap = map[string]fs.FS{}
	defer func() {
		for k, v := range fsMap {
			if closer, ok := v.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					ext.logger.Error().Err(err).Msgf("cannot close FS '%s'", k)
				}
			}
		}
	}()
	if ext.destFS == nil {
		return errors.New("destination FS is not set")
	}
	var manifest strings.Builder
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
				var internal string
				for _, internal = range internals {
					parts := strings.SplitN(internal, "/", 2)
					if _, ok := fsMap[parts[0]]; ok {
						break
					}
				}
				parts := strings.SplitN(internal, "/", 2)
				if _, ok := fsMap[parts[0]]; !ok {
					readFS, err := ext.getVersionFS(parts[0])
					if err != nil {
						return errors.Wrapf(err, "cannot get version FS for '%s'", parts[0])
					}
					fsMap[parts[0]] = readFS
				}
				src, err := fsMap[parts[0]].Open(parts[1])
				if err != nil {
					return errors.Wrapf(err, "cannot open '%s'", internal)
				}
				defer func(src fs.File) {
					err := src.Close()
					if err != nil {
						ext.logger.Error().Err(err).Msgf("cannot close '%s'", internal)
					}
				}(src)
				target, err := writefs.Create(ext.destFS, external)
				if err != nil {
					return errors.Wrapf(err, "cannot create '%v/%s'", ext.destFS, external)
				}
				defer func(target writefs.FileWrite) {
					err := target.Close()
					if err != nil {
						ext.logger.Error().Err(err).Msgf("cannot close '%v'", target)
					}
				}(target)
				ext.logger.Debug().Msgf("writing '%s' -> '%v/%s'", internal, ext.destFS, external)
				copyDigests, err := checksum.Copy([]checksum.DigestAlgorithm{digestAlg}, src, target)
				if err != nil {
					return errors.Wrapf(err, "error copying '%s' -> '%v/%s'", internal, ext.destFS, external)
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
		fp, err := writefs.Create(ext.destFS, manifestName)
		if err != nil {
			return errors.Wrapf(err, "cannot crate manifest file %v/%s", ext.destFS, manifestName)
		}
		if _, err := io.WriteString(fp, manifest.String()); err != nil {
			if err := fp.Close(); err != nil {
				return errors.Wrapf(err, "cannot close manifest file %v/%s", ext.destFS, manifestName)
			}
			return errors.Wrapf(err, "cannot write manifest file %v/%s", ext.destFS, manifestName)
		}
		if err := fp.Close(); err != nil {
			return errors.Wrapf(err, "cannot close manifest file %v/%s", ext.destFS, manifestName)
		}
	}
	ext.logger.Debug().Msgf("object '%s' extracted", inv.GetID())
	return nil

}

func (ext *extractor) WithObject(o object.Object) object.Extractor {
	ext.Object = o
	ext.objectReadFS = o.GetReadFS()
	return ext
}

func (ext *extractor) WithDestFS(destFS appendfs.FS) object.Extractor {
	ext.destFS = destFS
	return ext
}

var _ object.Extractor = (*extractor)(nil)
