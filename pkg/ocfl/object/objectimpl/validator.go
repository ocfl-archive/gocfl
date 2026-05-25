package objectimpl

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/util"
)

// ValidatorConfig holds the configuration for the validator.
type ValidatorConfig struct{}

// NewObjectBaseValidator creates a new instance of an object validator.
func NewObjectBaseValidator(ctx context.Context, factory factory.FactoryObject, config any, logger ocfllogger.OCFLLogger) object.Validator {
	validatorConfig, ok := config.(*ValidatorConfig)
	if config != nil && !ok {
		logger.Error().Msg("invalid config type for validator")
	}
	val := &validator{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "validator"),
		config:  validatorConfig,
	}
	val.getVersionFS = val._getVersionFS
	return val
}

// validator is the internal implementation of the OCFL object validator.
type validator struct {
	object.Object
	ctx          context.Context
	factory      factory.FactoryObject
	logger       ocfllogger.OCFLLogger
	config       *ValidatorConfig
	getVersionFS func(string) (fs.FS, error)
}

// WithObject attaches an OCFL object to the validator.
func (obj *validator) WithObject(obj2 object.Object) object.Validator {
	obj.Object = obj2
	return obj
}

// _getVersionFS is the default implementation for getting a filesystem for a specific version.
func (obj *validator) _getVersionFS(version string) (fs.FS, error) {
	readFS, err := fs.Sub(obj.GetReadFS(), version)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create sub FS '%v'", version)
	}
	return readFS, nil
}

// Validate performs a full validation of the OCFL object.
func (obj *validator) Validate() error {
	var fsMap = map[string]fs.FS{}
	defer func() {
		for k, v := range fsMap {
			if closer, ok := v.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					obj.logger.Error().Err(err).Msgf("cannot close FS '%s'", k)
				}
			}
		}
	}()

	fsys := obj.GetReadFS()
	if fsys == nil {
		obj.logger.Panic().Msg("object FS is not set")
	}
	inv := obj.GetInventory()
	//TODO implement me
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	obj.logger.Info().Msgf("object '%s' with object version '%s' found", inv.GetID(), obj.factory.GetVersion())

	if err := obj.checkRootEntries(fsMap); err != nil {
		return errors.WithStack(err)
	}

	if err := obj.checkFilesAndVersions(fsMap); err != nil {
		return errors.WithStack(err)
	}

	//todo: is there something missing?
	/*
		dAlgs := []checksum.DigestAlgorithm{inv.GetDigestAlgorithm()}
		dAlgs = append(dAlgs, util.SeqToSlice(inv.GetFixity().GetDigestAlgorithms())...)
	*/
	return nil
}

// checkRootEntries validates the entries in the object's root directory.
func (obj *validator) checkRootEntries(fsMap map[string]fs.FS) error {
	inv := obj.GetInventory()
	// check for allowed files and directories
	allowedDirs := []string{"logs", "extensions"}
	for v := range inv.GetVersions().GetVersionNumbers() {
		allowedDirs = append(allowedDirs, v.String())
	}
	versionCounter := 0
	entries, err := fs.ReadDir(obj.GetReadFS(), ".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !slices.Contains(allowedDirs, entry.Name()) {
				obj.logger.ValidationError(validation.E001, "invalid directory '%s' found", entry.Name())
				// could it be a version folder?
				if _, err := strconv.Atoi(strings.TrimLeft(entry.Name(), "v0")); err == nil {
					if err2 := obj.checkVersionFolder(fsMap, entry.Name()); err2 == nil {
						obj.logger.ValidationError(validation.E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			for v := range inv.GetVersions().GetVersionNumbers() {
				if v.String() == entry.Name() {
					if err := obj.checkVersionFolder(fsMap, entry.Name()); err != nil {
						return errors.WithStack(err)
					}
					versionCounter++
					break
				}
			}
		} else {
			if !allowedFilesRegexp.MatchString(entry.Name()) {
				obj.logger.ValidationError(validation.E001, "invalid file '%s' found", entry.Name())
			}
		}
	}

	invVersionCounter := len(util.SeqToSlice(inv.GetVersions().GetVersionNumbers()))
	if versionCounter != invVersionCounter {
		obj.logger.ValidationError(validation.E010, "number of version in inventory (%v) does not fit version in filesystem (%v)", versionCounter, invVersionCounter)
	}
	return nil
}

// allowedFilesRegexp matches the filenames allowed in an OCFL object's root or version directory.
var allowedFilesRegexp = regexp.MustCompile(`^(inventory.json(\.sha512|\.sha384|\.sha256|\.sha1|\.md5)?|0=ocfl_object_[0-9]+\.[0-9]+)$`)

// getVersionInventories loads the inventory files for all versions of the object.
func (obj *validator) getVersionInventories(fsMap map[string]fs.FS) (map[string]inventory.Inventory, string, error) {
	inv := obj.GetInventory()
	if inv.GetVersions().IsEmpty() {
		return map[string]inventory.Inventory{}, "", nil
	}

	versionStrings := util.SeqToSlice(inv.GetVersions().GetVersionNumbers())

	// sort in ascending order
	slices.SortFunc(versionStrings, func(a, b *inventory.VersionNumber) int {
		if a.Less(b) {
			return -1
		}

		if a.Equal(b) {
			return 0
		}

		return 1
	})
	versionInventories := map[string]inventory.Inventory{}
	var lastDigestString string
	for _, ver := range versionStrings {
		versionName := ver.String()
		if _, ok := fsMap[versionName]; !ok {
			vFS, err := obj.getVersionFS(versionName)
			if err != nil {
				return nil, "", errors.Wrapf(err, "cannot get version FS for '%s'", versionName)
			}
			fsMap[versionName] = vFS
		}
		vi, digestString, err := loadInventoryFile(obj.ctx, fsMap[versionName], "inventory.json", obj.GetOCFLVersion(), obj.factory, obj.logger)
		if err != nil {
			if errors.Is(errors.Cause(err), fs.ErrNotExist) {
				obj.logger.ValidationError(validation.E010, "inventory file '%s' does not exist", ver.String())
				continue
			}
			return nil, "", errors.Wrapf(err, "cannot load inventory from folder '%s'", ver)
		}
		versionInventories[ver.String()] = vi
		lastDigestString = digestString
	}
	return versionInventories, lastDigestString, nil
}

// checkVersionFolder validates the contents of a specific version directory.
func (obj *validator) checkVersionFolder(fsMap map[string]fs.FS, version string) error {
	if _, ok := fsMap[version]; !ok {
		vFS, err := obj.getVersionFS(version)
		if err != nil {
			return errors.Wrapf(err, "cannot get version FS for '%s'", version)
		}
		fsMap[version] = vFS
	}
	versionFS := fsMap[version]
	versionEntries, err := fs.ReadDir(versionFS, ".")
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder '%s'", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !allowedFilesRegexp.MatchString(ve.Name()) {
				obj.logger.ValidationError(validation.E015, "found extra file '%s' in version directory '%s'", ve.Name(), version)
			}
		}
	}
	obj.logger.Debug().Msgf("found %d correct version inventory files", len(versionEntries))
	return nil
}

// checkFilesAndVersions orchestrates the validation of files and their versions across the object.
func (obj *validator) checkFilesAndVersions(fsMap map[string]fs.FS) error {
	inv := obj.GetInventory()
	versionStrings := util.SeqToSlice(inv.GetVersions().GetVersionNumbers())
	if len(versionStrings) > 1 {
		slices.SortFunc(versionStrings, func(a, b *inventory.VersionNumber) int {
			if a.Less(b) {
				return -1
			}
			if a.Equal(b) {
				return 0
			}
			return 1
		})
	}

	// load object content files
	objectContentFiles, objectFilesFlat, err := obj.walkVersionFiles(versionStrings)
	if err != nil {
		return errors.WithStack(err)
	}

	// load all inventories
	versionInventories, lastDigestString, err := obj.getVersionInventories(fsMap)
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}

	obj.checkInventoryConsistency(versionStrings, versionInventories, lastDigestString)

	csDigestFiles, err := obj.createContentManifest(fsMap)
	if err != nil {
		return errors.Wrap(err, "cannot create content manifest")
	}
	if err := inv.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrap(err, "cannot check file digests for object root")
	}

	if err := obj.checkVersionInventories(versionStrings, versionInventories, csDigestFiles); err != nil {
		return errors.WithStack(err)
	}

	obj.checkManifestFiles(versionInventories, objectFilesFlat)

	obj.checkContentFiles(objectContentFiles, versionInventories)

	return nil
}

// walkVersionFiles traverses all version directories to collect information about present files.
func (obj *validator) walkVersionFiles(versionStrings []*inventory.VersionNumber) (map[string][]string, []string, error) {
	inv := obj.GetInventory()
	objectContentFiles := map[string][]string{}
	objectFilesFlat := []string{}

	for _, ver := range versionStrings {
		verStr := ver.String()
		versionContent := path.Join(verStr, inv.GetContentDir())
		if _, ok := objectContentFiles[verStr]; !ok {
			objectContentFiles[verStr] = []string{}
		}
		if err := fs.WalkDir(
			obj.GetReadFS(),
			verStr,
			func(path string, d fs.DirEntry, err error) error {
				path = filepath.ToSlash(path)
				if d == nil || d.IsDir() {
					if !strings.HasPrefix(path, versionContent) && path != verStr && !strings.HasPrefix(verStr+"/"+inv.GetContentDir(), path) {
						obj.logger.ValidationError(validation.W002, "extra dir '%s' in version '%s'", path, verStr)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[verStr] = append(objectContentFiles[verStr], path)
					}
				}
				return nil
			},
		); err != nil {
			return nil, nil, errors.Wrapf(err, "cannot walk version '%v/%s'", obj.GetReadFS(), verStr)
		}
		// leerer content ordner
		if len(objectContentFiles[verStr]) == 0 {
			fi, err := fs.Stat(obj.GetReadFS(), versionContent)
			if err == nil && fi.IsDir() {
				obj.logger.ValidationError(validation.W003, "empty content folder '%s'", versionContent)
			}
		}
	}
	return objectContentFiles, objectFilesFlat, nil
}

// checkInventoryConsistency ensures that inventories in version folders are consistent with each other and the root inventory.
func (obj *validator) checkInventoryConsistency(versionStrings []*inventory.VersionNumber, versionInventories map[string]inventory.Inventory, lastDigestString string) {
	inv := obj.GetInventory()
	var lastInventory inventory.Inventory
	var lastNumber string
	for _, ver := range versionStrings {
		versionInventory, ok := versionInventories[ver.String()]
		if !ok {
			continue
		}
		if lastInventory == nil {
			lastInventory = versionInventory
			lastNumber = ver.String()
			continue
		}
		allVersions := versionInventory.GetVersions()
		for lastVerNumber, lastVer := range lastInventory.GetVersions().Iterate() {
			if !lastVer.Equals(allVersions.GetVersion(lastVerNumber)) {
				obj.logger.ValidationError(validation.W011, "version inventory %s/%s does not match %s/%s", ver.String(), lastVerNumber, lastNumber, lastVerNumber)
			}
		}
		lastInventory = versionInventory
		lastNumber = ver.String()
	}
	if lastDigestString != "" {
		sidecarPath := fmt.Sprintf("%s.%s", "inventory.json", inv.GetDigestAlgorithm())
		digestString, err := getInventorySidecarChecksum(obj.GetReadFS(), sidecarPath, obj.logger)
		if err == nil && lastDigestString != digestString {
			obj.logger.ValidationError(validation.E064, "checksum of latest version inventory and root inventory are different")
		}
	}
}

// checkVersionInventories validates individual version inventories against the overall object state.
func (obj *validator) checkVersionInventories(versionStrings []*inventory.VersionNumber, versionInventories map[string]inventory.Inventory, csDigestFiles map[checksum.DigestAlgorithm]map[string][]string) error {
	inv := obj.GetInventory()
	contentDir := ""
	if len(versionStrings) > 0 {
		if vi, ok := versionInventories[versionStrings[0].String()]; ok {
			contentDir = vi.GetRealContentDir()
		}
	}

	for i, ver := range versionStrings {
		verStr := ver.String()
		vi := versionInventories[verStr]
		if vi == nil {
			continue
		}
		if contentDir != vi.GetRealContentDir() {
			obj.logger.ValidationError(validation.E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", vi.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := vi.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", verStr)
		}

		// check for extra files in version folder
		digestAlg := vi.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{vi.GetContentDir()}
		versionEntries, err := fs.ReadDir(obj.GetReadFS(), verStr)
		if err != nil {
			obj.logger.ValidationError(validation.E010, "cannot read version folder '%s'", verStr)
		} else {
			for _, entry := range versionEntries {
				if entry.IsDir() {
					if !slices.Contains(allowedDirs, entry.Name()) {
						obj.logger.ValidationError(validation.W002, "extra dir '%s' in version directory '%s'", entry.Name(), verStr)
					}
				} else {
					if !slices.Contains(allowedFiles, entry.Name()) {
						obj.logger.ValidationError(validation.E015, "extra file '%s' in version directory '%s'", entry.Name(), verStr)
					}
				}
			}
		}

		// check spec consistency
		if i < len(versionStrings)-1 {
			nextVer := versionStrings[i+1]
			if viNext, ok := versionInventories[nextVer.String()]; ok {
				if !inventory.SpecIsLessOrEqual(vi.GetSpec(), viNext.GetSpec()) {
					obj.logger.ValidationError(validation.E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", ver, vi.GetSpec(), nextVer, viNext.GetSpec())
				}
			}
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion.String()]; ok {
			if !lastInv.Equals(inv) {
				obj.logger.ValidationError(validation.E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}
	return nil
}

// checkManifestFiles verifies that all files listed in manifests actually exist in the object content.
func (obj *validator) checkManifestFiles(versionInventories map[string]inventory.Inventory, objectFilesFlat []string) {
	inv := obj.GetInventory()
	for inventoryVersion, vi := range versionInventories {
		for manifestFile := range vi.GetManifest().GetFilesFlat() {
			if !slices.Contains(objectFilesFlat, manifestFile) {
				obj.logger.ValidationError(validation.E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}
	for manifestFile := range inv.GetManifest().GetFilesFlat() {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			obj.logger.ValidationError(validation.E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
		}
	}
}

// checkContentFiles ensures that all physical content files are correctly referenced by the inventories.
func (obj *validator) checkContentFiles(objectContentFiles map[string][]string, versionInventories map[string]inventory.Inventory) {
	inv := obj.GetInventory()
	rootVersion := inv.GetHead()
	rootManifestFiles := util.SeqToSlice(inv.GetManifest().GetFilesFlat())

	for objectContentVersion, files := range objectContentFiles {
		ocvNumber := inventory.NewVersionNumber().WithString(objectContentVersion)
		// check against other version inventories
		for inventoryVersion, vi := range versionInventories {
			ivNumber := inventory.NewVersionNumber().WithString(inventoryVersion)
			if ocvNumber.Less(ivNumber) {
				versionManifestFiles := util.SeqToSlice(vi.GetManifest().GetFilesFlat())
				for _, file := range files {
					if !slices.Contains(versionManifestFiles, file) {
						obj.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", file, inventoryVersion)
					}
				}
			}
		}
		// check against root inventory
		if ocvNumber.Less(rootVersion) {
			for _, file := range files {
				if !slices.Contains(rootManifestFiles, file) {
					obj.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", file, rootVersion)
				}
			}
		}

		// check id consistency and head in version inventories
		if vi, ok := versionInventories[objectContentVersion]; ok {
			if inv.GetID() != vi.GetID() {
				obj.logger.ValidationError(validation.E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", inv.GetID(), objectContentVersion, vi.GetID())
			}
			if vi.GetHead().IsValid() && vi.GetHead().String() != objectContentVersion {
				obj.logger.ValidationError(validation.E040, "wrong head '%s' in manifest for version '%s'", vi.GetHead(), objectContentVersion)
			}
			if vi.GetDigestAlgorithm() != inv.GetDigestAlgorithm() {
				obj.logger.ValidationError(validation.W000, "different digest algorithm '%s' in version '%s'", vi.GetDigestAlgorithm(), objectContentVersion)
			}
			for vNum, vVer := range vi.GetVersions().Iterate() {
				testV := inv.GetVersions().GetVersion(vNum)
				if testV == nil {
					obj.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not in object root manifest", vVer, vNum)
				} else if !testV.Equals(vVer) {
					obj.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not equal to version in object root manifest", vVer, vNum)
				}
			}
		}
	}
}

// createContentManifest generates a manifest of all files found in the object's version directories.
// It iterates through all directories in the object root, identifies version folders,
// and collects checksums for all files within their respective content directories
// using the object's primary digest algorithm and any additional fixity algorithms.
func (obj *validator) createContentManifest(fsMap map[string]fs.FS) (map[checksum.DigestAlgorithm]map[string][]string, error) {
	inv := obj.GetInventory()
	digestAlgorithms := append(util.SeqToSlice(inv.GetFixity().GetDigestAlgorithms()), inv.GetDigestAlgorithm())
	result := map[checksum.DigestAlgorithm]map[string][]string{}

	entries, err := fs.ReadDir(obj.GetReadFS(), ".")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read dir '%v'", obj.GetReadFS())
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "extensions" {
			continue
		}
		versionNumber := inventory.NewVersionNumber().WithString(entry.Name())
		if versionNumber.Int() == 0 {
			obj.logger.ValidationError(validation.E012, "folder '%v' is not a version", entry.Name())
			continue
		}
		if err := obj.processVersionContent(fsMap, entry.Name(), digestAlgorithms, result); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return result, nil
}

// processVersionContent handles the validation of content for a specific version.
// It ensures the version's filesystem is available in the provided fsMap,
// traverses the version's content directory, and triggers checksum calculations
// for every file found.
func (obj *validator) processVersionContent(fsMap map[string]fs.FS, version string, digestAlgorithms []checksum.DigestAlgorithm, result map[checksum.DigestAlgorithm]map[string][]string) error {
	inv := obj.GetInventory()
	if _, ok := fsMap[version]; !ok {
		vFS, err := obj.getVersionFS(version)
		if err != nil {
			return errors.Wrapf(err, "cannot get version FS for '%s'", version)
		}
		fsMap[version] = vFS
	}
	versionFS := fsMap[version]

	return fs.WalkDir(
		versionFS,
		inv.GetContentDir(),
		func(fpath string, d fs.DirEntry, dirErr error) error {
			if dirErr != nil {
				return errors.Wrapf(dirErr, "error walking into %s", fpath)
			}
			if d == nil || d.IsDir() {
				return nil
			}
			return obj.calculateFileChecksums(versionFS, version, fpath, digestAlgorithms, result)
		})
}

// calculateFileChecksums computes checksums for a specific file using a list of algorithms.
// It opens the file from the version's filesystem, calculates the requested digests,
// and updates the result map with the mapping of checksums to their root-relative file paths.
func (obj *validator) calculateFileChecksums(versionFS fs.FS, version, fpath string, digestAlgorithms []checksum.DigestAlgorithm, result map[checksum.DigestAlgorithm]map[string][]string) error {
	fp, err := versionFS.Open(fpath)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", fpath)
	}
	defer fp.Close()

	css, err := checksum.Copy(digestAlgorithms, fp, &checksum.NullWriter{})
	if err != nil {
		return errors.Wrapf(err, "cannot read and create checksums for file '%s'", fpath)
	}

	fnameInRoot := path.Join(version, fpath)
	for d, cs := range css {
		if _, ok := result[d]; !ok {
			result[d] = map[string][]string{}
		}
		if _, ok := result[d][cs]; !ok {
			result[d][cs] = []string{}
		}
		result[d][cs] = append(result[d][cs], fnameInRoot)
	}
	obj.logger.Debug().Msgf("calculated %d checksums for file '%s'", len(css), fnameInRoot)
	return nil
}

var _ object.Validator = (*validator)(nil)
