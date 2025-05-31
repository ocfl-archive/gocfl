package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io/fs"
	"slices"
	"strings"
)

func newVersionBase(objectID string, version string, ctx context.Context, fsys fs.FS, ocflVersion OCFLVersion, inventory Inventory, packages VersionPackages, manager ExtensionManager, logger zLogger.ZLogger, factory *archiveerror.Factory) (*VersionBase, error) {
	if inventory == nil {
		return nil, errors.New("inventory must not be nil")
	}
	if fsys == nil {
		return nil, errors.New("filesystem must not be nil")
	}
	if version == "" {
		return nil, errors.New("version must not be empty")
	}
	ob := &VersionBase{
		objectID:         objectID,
		version:          version,
		ctx:              ctx,
		ocflVersion:      ocflVersion,
		inventory:        inventory,
		packages:         packages,
		fsys:             fsys,
		extensionManager: manager,
		logger:           logger,
		errorFactory:     factory,
	}
	if slices.Contains([]OCFLVersion{Version1_0, Version1_1}, ocflVersion) && packages != nil {
		return nil, errors.New("packages must be nil for OCFL versions 1.0 and 1.1")
	}
	return ob, nil
}

type VersionBase struct {
	objectID             string
	version              string
	ocflVersion          OCFLVersion
	inventory            Inventory
	packages             VersionPackages
	fsys                 fs.FS
	extensionManager     ExtensionManager
	logger               zLogger.ZLogger
	errorFactory         *archiveerror.Factory
	ctx                  context.Context
	inventoryData        []byte
	inventorySidecar     []byte
	contentFilenames     []string
	contentFileChecksums map[string]map[checksum.DigestAlgorithm]string
}

func (v *VersionBase) getVersionReader() (VersionReader, error) {
	if slices.Contains([]OCFLVersion{Version1_0, Version1_1}, v.ocflVersion) ||
		v.packages == nil {
		return NewVersionReaderPlain(v.version, v.fsys)
	}
	pv, ok := v.packages.GetVersion(v.version)
	if !ok {
		return NewVersionReaderPlain(v.version, v.fsys)
	}
	switch strings.ToLower(pv.Metadata.Format) {
	case "zip":
		return NewVersionReaderZIP(v.version, v.fsys, pv.Packages, v.logger)
	}
	return nil, errors.Errorf("unknown version package format '%s'", pv.Metadata.Format)
}

func (v *VersionBase) prepareContentFiles() error {
	if v.contentFileChecksums != nil && v.contentFilenames != nil {
		// Already prepared
		return nil
	}
	inventoryFixity := v.inventory.GetFixity()
	fixityAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inventoryFixity {
		fixityAlgorithms = append(fixityAlgorithms, alg)
	}
	digestAlgorithm := v.inventory.GetDigestAlgorithm()
	vr, err := v.getVersionReader()
	if err != nil {
		return errors.Wrapf(err, "cannot get version reader for version %s", v.version)
	}
	sidecarFilename := fmt.Sprintf("inventory.json.%s", digestAlgorithm)
	fileChecksums, fullContent, err := vr.GetFilenameChecksum(append(fixityAlgorithms, digestAlgorithm), []string{"inventory.json", sidecarFilename})
	if err != nil {
		return errors.Wrapf(err, "cannot get filename checksums for version %s", v.version)
	}
	v.contentFileChecksums = fileChecksums
	v.contentFilenames = make([]string, 0, len(fileChecksums))
	for file := range fileChecksums {
		v.contentFilenames = append(v.contentFilenames, file)
	}
	if data, ok := fullContent["inventory.json"]; ok {
		v.inventoryData = data
	} else {
		v.inventoryData = nil // No inventory file found
	}
	if data, ok := fullContent[sidecarFilename]; ok {
		v.inventorySidecar = data
	} else {
		v.inventorySidecar = nil // No sidecar file found
	}
	return nil
}

func (v *VersionBase) ValidateChecksums() error {
	// calculate all checksums for version files
	if err := v.prepareContentFiles(); err != nil {
		return errors.Wrapf(err, "cannot prepare content files for version %s", v.version)
	}
	inventoryFixity := v.inventory.GetFixity()
	fixityAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inventoryFixity {
		fixityAlgorithms = append(fixityAlgorithms, alg)
	}
	digestAlgorithm := v.inventory.GetDigestAlgorithm()

	//	inventoryManifest := v.inventory.GetManifest()
	//	inventoryFixity := v.inventory.GetFixity()
	v.inventory.IterateStateFiles(v.version, func(internal []string, external []string, digest string) error {
		for _, file := range internal {
			if checksums, ok := v.contentFileChecksums[file]; ok {
				if cs, ok := checksums[digestAlgorithm]; ok {
					if cs != digest {
						addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E092, "invalid digest for file '%s'", file)
						return errors.Errorf("checksum mismatch for file %s in version %s: expected %s, got %s",
							file, v.version, digest, cs)
					}
				} else {
					return errors.Errorf("checksum for file %s not found in version %s", file, v.version)
				}
			}
		}
		return nil
	})
	return nil
}
