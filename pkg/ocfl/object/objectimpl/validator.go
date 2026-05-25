package objectimpl

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"slices"
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
		ctx:                ctx,
		factory:            factory,
		logger:             logger.With("task", "validator"),
		config:             validatorConfig,
		allowedFilesRegexp: allowedFilesRegexp,
		allowedDirsRegexp:  allowedDirsRegexp,
		versionFSMap:       factory.NewVersionFSMap(ctx),
	}
	val.getVersion = val._getVersion
	return val
}

// validator is the internal implementation of the OCFL object validator.
type validator struct {
	object.Object
	ctx                context.Context
	factory            factory.FactoryObject
	logger             ocfllogger.OCFLLogger
	config             *ValidatorConfig
	versionFSMap       object.VersionFSMap
	allowedFilesRegexp *regexp.Regexp
	getVersion         func(name string) *inventory.VersionNumber
	allowedDirsRegexp  *regexp.Regexp
}

// WithObject attaches an OCFL object to the validator.
func (val *validator) WithObject(obj2 object.Object) object.Validator {
	val.Object = obj2
	val.versionFSMap.WithBaseFS(obj2.GetReadFS())
	return val
}

// Close finalizes the validator.
func (val *validator) Close() error {
	return val.versionFSMap.Close()
}

// Validate performs a full validation of the OCFL object.
func (val *validator) Validate() error {
	fsys := val.GetReadFS()
	if fsys == nil {
		val.logger.Panic().Msg("object FS is not set")
	}
	inv := val.GetInventory()
	//TODO implement me
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	val.logger.Info().Msgf("object '%s' with object version '%s' found", inv.GetID(), val.factory.GetVersion())

	if err := val.checkRootEntries(); err != nil {
		return errors.WithStack(err)
	}

	if err := val.checkFilesAndVersions(); err != nil {
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
func (val *validator) checkRootEntries() error {
	inv := val.GetInventory()
	// check for allowed files and directories
	/*
		for v := range inv.GetVersions().GetVersionNumbers() {
			allowedDirs = append(allowedDirs, v.String())
		}
	*/
	versionCounter := 0
	entries, err := fs.ReadDir(val.GetReadFS(), ".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	lastVersion := inventory.NewVersionNumber()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if val.allowedDirsRegexp.MatchString(name) {
				continue
			}
		} else {
			if val.allowedFilesRegexp.MatchString(name) {
				continue
			}
		}
		//val.logger.ValidationError(validation.E001, "invalid directory '%s' found", entry.Name())
		// could it be a version folder?
		ver := val.getVersion(name)
		if ver == nil || ver.Int() <= 0 {
			val.logger.ValidationError(validation.E001, "invalid entry '%s' found", name)
			continue
		}
		if lastVersion.Less(ver) {
			lastVersion = ver
		}
		versionCounter++
		if err := val.checkVersionFolder(ver); err != nil {
			fmt.Println(err)
		}
	}
	if inv.GetVersions().LatestVersionNumber().Less(lastVersion) {
		val.logger.ValidationError(validation.E046, "root manifest not most recent because of '%s'", lastVersion)
	}
	invVersionCounter := len(util.SeqToSlice(inv.GetVersions().GetVersionNumbers()))
	if versionCounter != invVersionCounter {
		val.logger.ValidationError(validation.E010, "number of version in inventory (%v) does not fit version in filesystem (%v)", invVersionCounter, versionCounter)
	}
	return nil
}

// getVersionInventories loads the inventory files for all versions of the object.
func (val *validator) getVersionInventories() (map[string]inventory.Inventory, string, error) {
	inv := val.GetInventory()
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
		vFS, err := val.versionFSMap.GetVersionFS(ver)
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot get version FS for '%s'", versionName)
		}
		vi, digestString, err := loadInventoryFile(val.ctx, vFS, "inventory.json", val.GetOCFLVersion(), val.factory, val.logger)
		if err != nil {
			if errors.Is(errors.Cause(err), fs.ErrNotExist) {
				val.logger.ValidationError(validation.E010, "inventory file '%s' does not exist", ver.String())
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
func (val *validator) checkVersionFolder(version *inventory.VersionNumber) error {
	if version == nil {
		return errors.New("version cannot be nil")
	}
	if version.Int() <= 0 {
		return errors.New("version number cannot zero or be negative")
	}
	versionFS, err := val.versionFSMap.GetVersionFS(version)
	if err != nil {
		return errors.Wrapf(err, "cannot get version FS for '%s'", version)
	}
	versionEntries, err := fs.ReadDir(versionFS, ".")
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder '%s'", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !val.allowedFilesRegexp.MatchString(ve.Name()) {
				val.logger.ValidationError(validation.E015, "found extra file '%s' in version directory '%s'", ve.Name(), version)
			}
		}
	}
	val.logger.Debug().Msgf("found %d correct version inventory files", len(versionEntries))
	return nil
}

// checkFilesAndVersions orchestrates the validation of files and their versions across the object.
func (val *validator) checkFilesAndVersions() error {
	inv := val.GetInventory()
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
	objectContentFiles, objectFilesFlat, err := val.walkVersionFiles(versionStrings)
	if err != nil {
		return errors.WithStack(err)
	}

	// load all inventories
	versionInventories, lastDigestString, err := val.getVersionInventories()
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}

	val.checkInventoryConsistency(versionStrings, versionInventories, lastDigestString)

	csDigestFiles, err := val.createContentManifest()
	if err != nil {
		return errors.Wrap(err, "cannot create content manifest")
	}
	if err := inv.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrap(err, "cannot check file digests for object root")
	}

	if err := val.checkVersionInventories(versionStrings, versionInventories, csDigestFiles); err != nil {
		return errors.WithStack(err)
	}

	val.checkManifestFiles(versionInventories, objectFilesFlat)

	val.checkContentFiles(objectContentFiles, versionInventories)

	return nil
}

// walkVersionFiles traverses all version directories to collect information about present files.
func (val *validator) walkVersionFiles(versionStrings []*inventory.VersionNumber) (map[string][]string, []string, error) {
	inv := val.GetInventory()
	objectContentFiles := map[string][]string{}
	objectFilesFlat := []string{}

	for _, ver := range versionStrings {
		verStr := ver.String()
		versionContent := path.Join(verStr, inv.GetContentDir())
		if _, ok := objectContentFiles[verStr]; !ok {
			objectContentFiles[verStr] = []string{}
		}
		vFS, err := val.versionFSMap.GetVersionFS(ver)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot get version FS for '%s'", verStr)
		}

		if err := fs.WalkDir(
			vFS,
			".",
			func(fpath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				fpath = filepath.ToSlash(fpath)
				var fullPath string
				if fpath == "." {
					fullPath = verStr
				} else {
					fullPath = path.Join(verStr, fpath)
				}
				if d.IsDir() {
					if !strings.HasPrefix(fullPath, versionContent) && fullPath != verStr && !strings.HasPrefix(verStr+"/"+inv.GetContentDir(), fullPath) {
						val.logger.ValidationError(validation.W002, "extra dir '%s' in version '%s'", fullPath, verStr)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, fullPath)
					if strings.HasPrefix(fullPath, versionContent) {
						objectContentFiles[verStr] = append(objectContentFiles[verStr], fullPath)
					}
				}
				return nil
			},
		); err != nil {
			return nil, nil, errors.Wrapf(err, "cannot walk version '%s'", verStr)
		}
		// leerer content ordner
		if len(objectContentFiles[verStr]) == 0 {
			fi, err := fs.Stat(vFS, inv.GetContentDir())
			if err == nil && fi.IsDir() {
				val.logger.ValidationError(validation.W003, "empty content folder '%s'", versionContent)
			}
		}
	}
	return objectContentFiles, objectFilesFlat, nil
}

// checkInventoryConsistency ensures that inventories in version folders are consistent with each other and the root inventory.
func (val *validator) checkInventoryConsistency(versionStrings []*inventory.VersionNumber, versionInventories map[string]inventory.Inventory, lastDigestString string) {
	inv := val.GetInventory()
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
				val.logger.ValidationError(validation.W011, "version inventory %s/%s does not match %s/%s", ver.String(), lastVerNumber, lastNumber, lastVerNumber)
			}
		}
		lastInventory = versionInventory
		lastNumber = ver.String()
	}
	if lastDigestString != "" {
		sidecarPath := fmt.Sprintf("%s.%s", "inventory.json", inv.GetDigestAlgorithm())
		digestString, err := getInventorySidecarChecksum(val.GetReadFS(), sidecarPath, val.logger)
		if err == nil && lastDigestString != digestString {
			val.logger.ValidationError(validation.E064, "checksum of latest version inventory and root inventory are different")
		}
	}
}

// checkVersionInventories validates individual version inventories against the overall object state.
func (val *validator) checkVersionInventories(versionStrings []*inventory.VersionNumber, versionInventories map[string]inventory.Inventory, csDigestFiles map[checksum.DigestAlgorithm]map[string][]string) error {
	inv := val.GetInventory()
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
			val.logger.ValidationError(validation.E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", vi.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := vi.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", verStr)
		}

		// check for extra files in version folder
		digestAlg := vi.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{vi.GetContentDir()}
		vFS, err := val.versionFSMap.GetVersionFS(ver)
		if err != nil {
			return errors.Wrapf(err, "cannot get version FS for '%s'", verStr)
		}
		versionEntries, err := fs.ReadDir(vFS, ".")
		if err != nil {
			val.logger.ValidationError(validation.E010, "cannot read version folder '%s'", verStr)
		} else {
			for _, entry := range versionEntries {
				if entry.IsDir() {
					if !slices.Contains(allowedDirs, entry.Name()) {
						val.logger.ValidationError(validation.W002, "extra dir '%s' in version directory '%s'", entry.Name(), verStr)
					}
				} else {
					if !slices.Contains(allowedFiles, entry.Name()) {
						val.logger.ValidationError(validation.E015, "extra file '%s' in version directory '%s'", entry.Name(), verStr)
					}
				}
			}
		}

		// check spec consistency
		if i < len(versionStrings)-1 {
			nextVer := versionStrings[i+1]
			if viNext, ok := versionInventories[nextVer.String()]; ok {
				if !inventory.SpecIsLessOrEqual(vi.GetSpec(), viNext.GetSpec()) {
					val.logger.ValidationError(validation.E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", ver, vi.GetSpec(), nextVer, viNext.GetSpec())
				}
			}
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion.String()]; ok {
			if !lastInv.Equals(inv) {
				val.logger.ValidationError(validation.E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}
	return nil
}

// checkManifestFiles verifies that all files listed in manifests actually exist in the object content.
func (val *validator) checkManifestFiles(versionInventories map[string]inventory.Inventory, objectFilesFlat []string) {
	inv := val.GetInventory()
	for inventoryVersion, vi := range versionInventories {
		for manifestFile := range vi.GetManifest().GetFilesFlat() {
			if !slices.Contains(objectFilesFlat, manifestFile) {
				val.logger.ValidationError(validation.E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}
	for manifestFile := range inv.GetManifest().GetFilesFlat() {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			val.logger.ValidationError(validation.E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
		}
	}
}

// checkContentFiles ensures that all physical content files are correctly referenced by the inventories.
func (val *validator) checkContentFiles(objectContentFiles map[string][]string, versionInventories map[string]inventory.Inventory) {
	inv := val.GetInventory()
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
						val.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", file, inventoryVersion)
					}
				}
			}
		}
		// check against root inventory
		if ocvNumber.Less(rootVersion) {
			for _, file := range files {
				if !slices.Contains(rootManifestFiles, file) {
					val.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", file, rootVersion)
				}
			}
		}

		// check id consistency and head in version inventories
		if vi, ok := versionInventories[objectContentVersion]; ok {
			if inv.GetID() != vi.GetID() {
				val.logger.ValidationError(validation.E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", inv.GetID(), objectContentVersion, vi.GetID())
			}
			if vi.GetHead().IsValid() && vi.GetHead().String() != objectContentVersion {
				val.logger.ValidationError(validation.E040, "wrong head '%s' in manifest for version '%s'", vi.GetHead(), objectContentVersion)
			}
			if vi.GetDigestAlgorithm() != inv.GetDigestAlgorithm() {
				val.logger.ValidationError(validation.W000, "different digest algorithm '%s' in version '%s'", vi.GetDigestAlgorithm(), objectContentVersion)
			}
			for vNum, vVer := range vi.GetVersions().Iterate() {
				testV := inv.GetVersions().GetVersion(vNum)
				if testV == nil {
					val.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not in object root manifest", vVer, vNum)
				} else if !testV.Equals(vVer) {
					val.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not equal to version in object root manifest", vVer, vNum)
				}
			}
		}
	}
}

// createContentManifest generates a manifest of all files found in the object's version directories.
// It iterates through all directories in the object root, identifies version folders,
// and collects checksums for all files within their respective content directories
// using the object's primary digest algorithm and any additional fixity algorithms.
func (val *validator) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	inv := val.GetInventory()
	digestAlgorithms := append(util.SeqToSlice(inv.GetFixity().GetDigestAlgorithms()), inv.GetDigestAlgorithm())
	result := map[checksum.DigestAlgorithm]map[string][]string{}

	for ver := range inv.GetVersions().GetVersionNumbers() {
		if err := val.processVersionContent(ver, digestAlgorithms, result); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return result, nil
}

// processVersionContent handles the validation of content for a specific version.
// It ensures the version's filesystem is available in the provided fsMap,
// traverses the version's content directory, and triggers checksum calculations
// for every file found.
func (val *validator) processVersionContent(version *inventory.VersionNumber, digestAlgorithms []checksum.DigestAlgorithm, result map[checksum.DigestAlgorithm]map[string][]string) error {
	inv := val.GetInventory()
	versionFS, err := val.versionFSMap.GetVersionFS(version)
	if err != nil {
		return errors.Wrapf(err, "cannot get version FS for '%s'", version)
	}

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
			return val.calculateFileChecksums(versionFS, version, fpath, digestAlgorithms, result)
		})
}

// calculateFileChecksums computes checksums for a specific file using a list of algorithms.
// It opens the file from the version's filesystem, calculates the requested digests,
// and updates the result map with the mapping of checksums to their root-relative file paths.
func (val *validator) calculateFileChecksums(versionFS fs.FS, version *inventory.VersionNumber, fpath string, digestAlgorithms []checksum.DigestAlgorithm, result map[checksum.DigestAlgorithm]map[string][]string) error {
	fp, err := versionFS.Open(fpath)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", fpath)
	}
	defer fp.Close()

	css, err := checksum.Copy(digestAlgorithms, fp, &checksum.NullWriter{})
	if err != nil {
		return errors.Wrapf(err, "cannot read and create checksums for file '%s'", fpath)
	}

	fnameInRoot := path.Join(version.String(), fpath)
	for d, cs := range css {
		if _, ok := result[d]; !ok {
			result[d] = map[string][]string{}
		}
		if _, ok := result[d][cs]; !ok {
			result[d][cs] = []string{}
		}
		result[d][cs] = append(result[d][cs], fnameInRoot)
	}
	val.logger.Debug().Msgf("calculated %d checksums for file '%s'", len(css), fnameInRoot)
	return nil
}

func (val *validator) _getVersion(name string) *inventory.VersionNumber {
	vn := inventory.NewVersionNumber().WithString(name)
	if vn.Int() == 0 {
		return nil
	}
	return vn
}

// allowedFilesRegexp matches the filenames allowed in an OCFL object's root or version directory.
var allowedFilesRegexp = regexp.MustCompile(`^(inventory.json(\.sha512|\.sha384|\.sha256|\.sha1|\.md5)?|0=ocfl_object_[0-9]+\.[0-9]+)$`)
var allowedDirsRegexp = regexp.MustCompile(`^extensions|logs$`)
var _ object.Validator = (*validator)(nil)
