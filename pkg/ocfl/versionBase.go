package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io/fs"
	"path"
	"slices"
	"strings"
)

func newVersionBase(
	objectID string,
	version string,
	ctx context.Context,
	fsys fs.FS,
	ocflVersion OCFLVersion,
	inventory Inventory,
	packages VersionPackages,
	manager ExtensionManager,
	logger zLogger.ZLogger,
	factory *archiveerror.Factory,
) (*VersionBase, error) {
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
	partsChecksum        map[string]string
}

func (v *VersionBase) Check() error {
	if err := v.prepareContentFiles(); err != nil {
		return errors.Wrapf(err, "cannot prepare content files for version %s", v.version)
	}
	folders, files, err := v.getContent(".")
	if err != nil {
		return errors.Wrapf(err, "cannot get folders for version %s", v.version)
	}
	// only one folder is allowed, which must match the contentDir
	if len(folders) != 1 || !slices.Contains(folders, v.inventory.GetContentDir()) {
		addValidationWarning(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, W002, "version '%s' has unexpected content folders: %v", v.version, folders)
	}
	if len(files) != 0 && len(files) != 2 && !slices.Contains(files, "inventory.json") && !slices.Contains(files, fmt.Sprintf("inventory.json.%s", v.inventory.GetDigestAlgorithm())) {
		addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E015, "version '%s' has unexpected content files: %v", v.version, files)
	}
	folders, files, err = v.getContent(v.inventory.GetContentDir())
	if err != nil {
		return errors.Wrapf(err, "cannot get content for version %s", v.version)
	}
	if len(folders) == 0 && len(files) == 0 {
		addValidationWarning(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, W003, "empty content in version '%s'", v.version)
		return nil
	}

	if versionInventory, err := v.getInventory(); err != nil {
		addValidationWarning(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, W010, "error getting inventory for version '%s': %v", v.version, err)
		v.logger.Error().Err(err).Str("objectID", v.objectID).Str("version", v.version).Msg("Error getting inventory for version")
	} else if versionInventory != nil {
		if !v.inventory.Contains(versionInventory) {
			addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E066, "inventory for version '%s' does not match the expected inventory", v.version)
		}
		if !SpecIsLessOrEqual(v.inventory.GetSpec(), versionInventory.GetSpec()) {
			addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", v.version, v.inventory.GetSpec(), versionInventory.GetSpec())
		}
		if v.objectID != versionInventory.GetID() {
			addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E037, "object ID in version '%s' (%s) does not match the expected object ID (%s)", v.version, versionInventory.GetID(), v.objectID)
		}
		if versionInventory.GetHead() != "" && versionInventory.GetHead() != v.version {
			addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E040, "head in version '%s' (%s) does not match the expected head (%s)", v.version, versionInventory.GetHead(), v.version)
		}
		if versionInventory.GetDigestAlgorithm() != v.inventory.GetDigestAlgorithm() {
			addValidationWarning(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, W000, "digest algorithm in version '%s' (%s) does not match the expected digest algorithm (%s)", v.version, versionInventory.GetDigestAlgorithm(), v.inventory.GetDigestAlgorithm())
		}
		versions := v.inventory.GetVersions()
		for verVer, verVersion := range versionInventory.GetVersions() {
			testV, ok := versions[verVer]
			if !ok {
				addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E066, "version '%s' in version folder '%s' not found in object root manifest", v.version, verVer)
			}
			if !testV.EqualState(verVersion) {
				addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E066, "version '%s' in version folder '%s' not equal to version in object root manifest", v.version, verVer)
			}
			if !testV.EqualMeta(verVersion) {
				addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E066, "version '%s' in version folder '%s' has different metadata as version in object root manifest", v.version, verVer)
			}
		}
	}
	csDigestFiles := map[checksum.DigestAlgorithm]map[string][]string{}
	for name, digestAlgorithms := range v.contentFileChecksums {
		for digestAlgorithm, checksumValue := range digestAlgorithms {
			if _, ok := csDigestFiles[digestAlgorithm]; !ok {
				csDigestFiles[digestAlgorithm] = make(map[string][]string)
			}
			if _, ok := csDigestFiles[digestAlgorithm][checksumValue]; !ok {
				csDigestFiles[digestAlgorithm][checksumValue] = []string{}
			}
			csDigestFiles[digestAlgorithm][checksumValue] = append(csDigestFiles[digestAlgorithm][checksumValue], v.version+"/"+name)
		}
	}
	if err := v.inventory.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrapf(err, "cannot check files for version %s", v.version)
	}
	//realContentDir := v.inventory.GetRealContentDir()

	return nil
}

func (v *VersionBase) getVersionReader() (VersionReader, error) {
	if slices.Contains([]OCFLVersion{Version1_0, Version1_1}, v.ocflVersion) ||
		v.packages == nil {
		v.logger.Debug().Str("objectID", v.objectID).Str("version", v.version).Msg("Using plain version reader for OCFL versions 1.0 and 1.1 or when no packages are defined")
		return NewVersionReaderPlain(v.version, v.fsys, v.logger)
	}
	pv, ok := v.packages.GetVersion(v.version)
	if !ok {
		v.logger.Debug().Str("objectID", v.objectID).Str("version", v.version).Msg("Using plain version reader for OCFL versions 2.0 without packages")
		return NewVersionReaderPlain(v.version, v.fsys, v.logger)
	}
	switch strings.ToLower(pv.Metadata.Format) {
	case "zip":
		v.logger.Debug().Str("objectID", v.objectID).Str("version", v.version).Msgf("Using ZIP version reader for OCFL version %s with packages %v", v.ocflVersion, pv.Packages)
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
	fileChecksums, fullContent, partsChecksum, err := vr.GetFilenameChecksum(digestAlgorithm, fixityAlgorithms, []string{"inventory.json", sidecarFilename})
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
	v.partsChecksum = partsChecksum
	return nil
}

func (v *VersionBase) getContent(root string) ([]string, []string, error) {
	if err := v.prepareContentFiles(); err != nil {
		return nil, nil, errors.Wrapf(err, "cannot prepare content files for version %s", v.version)
	}
	root = path.Clean(root) + "/"
	root = strings.TrimPrefix(root, "./") // Ensure root is relative
	folders := []string{}
	files := []string{}
	for _, file := range v.contentFilenames {
		if strings.HasPrefix(file, root) {
			folder := strings.TrimPrefix(file, root)
			if idx := strings.Index(folder, "/"); idx != -1 {
				folder = folder[:idx]
				if !slices.Contains(folders, folder) {
					folders = append(folders, folder)
				}
			} else {
				if !slices.Contains(files, folder) {
					files = append(files, folder)
				}
			}
		}
	}
	return folders, files, nil
}

func (v *VersionBase) getInventory() (Inventory, error) {
	if v.inventoryData == nil {
		return nil, nil
	}
	if v.inventorySidecar == nil {
		addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E058, "inventory sidecar file not found for version '%s'", v.version)
		return nil, nil
	}
	digestString := strings.TrimSpace(string(v.inventorySidecar))
	//if !strings.HasSuffix(digestString, " inventory.json") {
	matches := inventorySideCarFormat.FindStringSubmatch(digestString)
	if len(matches) == 0 {
		addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E061, "no suffix \" inventory.json\" found in inventory sidecar file for version '%s'", v.version)
		return nil, nil
	}
	digestString = matches[1]
	h, err := checksum.GetHash(v.inventory.GetDigestAlgorithm())
	if err != nil {
		//v.logger.Error().Err(err).Msgf("cannot get hash for digest algorithm '%s'", v.inventory.GetDigestAlgorithm())
		return nil, errors.Wrapf(err, "cannot get hash for digest algorithm '%s'", v.inventory.GetDigestAlgorithm())
	}
	h.Reset()
	h.Write(v.inventoryData)
	sumBytes := h.Sum(nil)
	inventoryDigestString := fmt.Sprintf("%x", sumBytes)
	if inventoryDigestString != digestString {
		addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E060, "inventory sidecar file for version '%s' has invalid digest '%s', expected '%s'", v.version, digestString, inventoryDigestString)
		return nil, nil
	}
	anyMap := map[string]any{}
	if err := json.Unmarshal(v.inventoryData, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(v.inventoryData))
	}
	var ocflVersion OCFLVersion
	t, ok := anyMap["type"]
	if !ok {
		return nil, errors.New("no type in inventory")
	}
	sStr, ok := t.(string)
	if !ok {
		return nil, errors.Errorf("type not a string in inventory - '%v'", t)
	}
	switch InventorySpec(sStr) {
	case InventorySpec1_1:
		ocflVersion = Version1_1
	case InventorySpec1_0:
		ocflVersion = Version1_0
	case InventorySpec2_0:
		ocflVersion = Version2_0
	default:
		// if we don't know anything use the old stuff
		ocflVersion = Version1_0
	}
	inventory, err := newInventory(v.ctx, v.version, ocflVersion, v.logger, v.errorFactory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := json.Unmarshal(v.inventoryData, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(v.inventoryData, &jsonMap); err2 != nil {
			addValidationErrors(v.ctx, GetValidationError(ocflVersion, E033).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", v.objectID))
			addValidationErrors(v.ctx, GetValidationError(ocflVersion, E034).AppendDescription("json syntax error: %v", err2).AppendContext("object '%v'", v.objectID))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(v.ctx, GetValidationError(ocflVersion, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]).AppendContext("object '%v'", v.objectID))
			}
		}
		return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(v.inventoryData))
	}

	return inventory, inventory.Finalize(false)

}

func (v *VersionBase) ValidateChecksums() error {
	// calculate all checksums for version files
	if err := v.prepareContentFiles(); err != nil {
		return errors.Wrapf(err, "cannot prepare content files for version %s", v.version)
	}
	//	inventoryManifest := v.inventory.GetManifest()
	inventoryFixity := v.inventory.GetFixity()
	fixityAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inventoryFixity {
		fixityAlgorithms = append(fixityAlgorithms, alg)
	}
	digestAlgorithm := v.inventory.GetDigestAlgorithm()

	v.inventory.IterateStateFiles(v.version, func(internal []string, external []string, digest string) error {
		for _, file := range internal {
			if fileChecksums, ok := v.contentFileChecksums[file]; ok {
				if cs, ok := fileChecksums[digestAlgorithm]; ok {
					if cs != digest {
						addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E092, "invalid digest for file '%s'", file)
					}
				} else {
					addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E092, "checksum for file '%s' not found in version '%s'", file, v.version)
				}
				fixityChecksums := inventoryFixity.Checksums(file)
				for _, alg := range fixityAlgorithms {
					fileChecksum, fileExists := fileChecksums[alg]
					fixityChecksum, fixityExists := fixityChecksums[alg]

					if !fileExists {
						addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E093, "checksum for file '%s' with algorithm '%s' not found in version '%s'", file, alg, v.version)
					}
					if !fixityExists {
						addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E093, "fixity checksum for file '%s' with algorithm '%s' not found in inventory", file, alg)
					}
					if fileChecksum != fixityChecksum {
						addValidationError(v.ctx, v.logger, v.ocflVersion, v.objectID, v.fsys, E093, "invalid fixity checksum for file '%s' with algorithm '%s'", file, alg)
					}
				}
			}
		}
		return nil
	})
	return nil
}

var _ Version = &VersionBase{}
