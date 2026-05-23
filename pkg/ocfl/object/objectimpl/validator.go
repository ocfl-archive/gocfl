package objectimpl

import (
	"context"
	"fmt"
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

type ValidatorConfig struct{}

func NewObjectBaseValidator(ctx context.Context, factory factory.FactoryObject, config any, logger ocfllogger.OCFLLogger) object.Validator {
	validatorConfig, ok := config.(*ValidatorConfig)
	if config != nil && !ok {
		logger.Error().Msg("invalid config type for validator")
	}
	return &validator{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "validator"),
		config:  validatorConfig,
	}
}

type validator struct {
	object.Object
	ctx     context.Context
	factory factory.FactoryObject
	logger  ocfllogger.OCFLLogger
	config  *ValidatorConfig
}

func (obj *validator) WithObject(obj2 object.Object) object.Validator {
	obj.Object = obj2
	return obj
}

func (obj *validator) Validate() error {
	fsys := obj.GetReadFS()
	if fsys == nil {
		obj.logger.Panic().Msg("object FS is not set")
	}
	inv := obj.GetInventory()
	//TODO implement me
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	obj.logger.Info().Msgf("object '%s' with object version '%s' found", inv.GetID(), obj.factory.GetVersion())
	// check folders

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
					if err2 := obj.checkVersionFolder(entry.Name()); err2 == nil {
						obj.logger.ValidationError(validation.E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			for v := range inv.GetVersions().GetVersionNumbers() {
				if v.String() == entry.Name() {
					if err := obj.checkVersionFolder(entry.Name()); err != nil {
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

	if err := obj.checkFilesAndVersions(); err != nil {
		return errors.WithStack(err)
	}

	//todo: is there something missing?
	/*
		dAlgs := []checksum.DigestAlgorithm{inv.GetDigestAlgorithm()}
		dAlgs = append(dAlgs, util.SeqToSlice(inv.GetFixity().GetDigestAlgorithms())...)
	*/
	return nil

}

var allowedFilesRegexp = regexp.MustCompile(`^(inventory.json(\.sha512|\.sha384|\.sha256|\.sha1|\.md5)?|0=ocfl_object_[0-9]+\.[0-9]+)$`)

func (obj *validator) getVersionInventories() (map[string]inventory.Inventory, string, error) {
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
		vi, digestString, err := loadInventoryFile(obj.ctx, obj.GetReadFS(), path.Join(ver.String(), "inventory.json"), obj.GetOCFLVersion(), obj.factory, obj.logger)
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

func (obj *validator) checkVersionFolder(version string) error {
	versionEntries, err := fs.ReadDir(obj.GetReadFS(), version)
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

func (obj *validator) checkFilesAndVersions() error {
	inv := obj.GetInventory()
	//ocflVersion := inv.GetOCFLVersion()
	// create list of version content directories
	versionContents := map[string]string{}
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

	for _, ver := range versionStrings {
		versionContents[ver.String()] = inv.GetContentDir()
	}

	// load object content files
	objectContentFiles := map[string][]string{}
	objectContentFilesFlat := []string{}
	objectFilesFlat := []string{}
	for ver, cont := range versionContents {
		// load all object version content files
		versionContent := path.Join(ver, cont)
		//inventoryFile := ver + "/inventory.json"
		if _, ok := objectContentFiles[ver]; !ok {
			objectContentFiles[ver] = []string{}
		}
		// erstelle eine liste aller dateien im versionsordner und eine liste aller dateien im contentordner der version
		if err := fs.WalkDir(
			obj.GetReadFS(),
			ver,
			func(path string, d fs.DirEntry, err error) error {
				path = filepath.ToSlash(path)
				if d == nil || d.IsDir() {
					if !strings.HasPrefix(path, versionContent) && path != ver && !strings.HasPrefix(ver+"/"+inv.GetContentDir(), path) {
						obj.logger.ValidationError(validation.W002, "extra dir '%s' in version '%s'", path, ver)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[ver] = append(objectContentFiles[ver], path)
						objectContentFilesFlat = append(objectContentFilesFlat, path)
					} /* else {

							if !strings.HasPrefix(path, inventoryFile) {
								obj.AddValidationWarning(W002, "extra file '%s' in version '%s'", path, ver)
							}

					}*/
				}
				return nil
			},
		); err != nil {
			return errors.Wrapf(err, "cannot walk version '%v/%s'", obj.GetReadFS(), ver)
		}
		// leerer content ordner
		if len(objectContentFiles[ver]) == 0 {
			fi, err := fs.Stat(obj.GetReadFS(), versionContent)
			if err != nil {
				if !errors.Is(errors.Cause(err), fs.ErrNotExist) {
					return errors.Wrapf(err, "cannot stat '%s'", versionContent)
				}
			} else {
				if fi.IsDir() {
					obj.logger.ValidationError(validation.W003, "empty content folder '%s'", versionContent)
				}
			}
		}
	}

	// load all inventories
	versionInventories, lastDigestString, err := obj.getVersionInventories()
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}
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
		// load root inventory checksum
		sidecarPath := fmt.Sprintf("%s.%s", "inventory.json", inv.GetDigestAlgorithm())
		digestString, err := getInventorySidecarChecksum(obj.GetReadFS(), sidecarPath, obj.logger)
		if err == nil {
			if lastDigestString != digestString {
				obj.logger.ValidationError(validation.E064, "checksum of latest version inventory and root inventory are different")
			}
		}

	}

	csDigestFiles, err := obj.createContentManifest()
	if err != nil {
		return errors.Wrap(err, "cannot create content manifest")
	}
	if err := inv.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrap(err, "cannot check file digests for object root")
	}

	contentDir := ""
	if len(versionStrings) > 0 {
		versionInventory, ok := versionInventories[versionStrings[0].String()]
		if !ok {
			obj.logger.ValidationError(validation.W010, "version inventory '%s' not found", versionStrings[0])
		} else {
			contentDir = versionInventory.GetRealContentDir()
		}
	}
	for _, ver := range versionStrings {
		inv := versionInventories[ver.String()]
		if inv == nil {
			continue
		}
		if contentDir != inv.GetRealContentDir() {
			obj.logger.ValidationError(validation.E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", inv.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := inv.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", ver)
		}
		digestAlg := inv.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{inv.GetContentDir()}
		versionEntries, err := fs.ReadDir(obj.GetReadFS(), ver.String())
		if err != nil {
			obj.logger.ValidationError(validation.E010, "cannot read version folder '%s'", ver)
			continue
			//			return errors.Wrapf(err, "cannot read dir '%s'", ver)
		}
		for _, entry := range versionEntries {
			if entry.IsDir() {
				if !slices.Contains(allowedDirs, entry.Name()) {
					obj.logger.ValidationError(validation.W002, "extra dir '%s' in version directory '%s'", entry.Name(), ver)
				}
			} else {
				if !slices.Contains(allowedFiles, entry.Name()) {
					obj.logger.ValidationError(validation.E015, "extra file '%s' in version directory '%s'", entry.Name(), ver)
				}
			}
		}
	}

	for key := 0; key < len(versionStrings)-1; key++ {
		v1 := versionStrings[key]
		vi1, ok := versionInventories[v1.String()]
		if !ok {
			obj.logger.ValidationError(validation.W010, "no inventory for version '%s'", versionStrings[key])
			continue
			// return errors.Errorf("no inventory for version '%s'", versionStrings[key])
		}
		v2 := versionStrings[key+1]
		vi2, ok := versionInventories[v2.String()]
		if !ok {
			obj.logger.ValidationError(validation.W000, "no inventory for version '%s'", versionStrings[key+1])
			continue
		}
		if !inventory.SpecIsLessOrEqual(vi1.GetSpec(), vi2.GetSpec()) {
			obj.logger.ValidationError(validation.E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", v1, vi1.GetSpec(), v2, vi2.GetSpec())
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion.String()]; ok {
			if !lastInv.Equals(obj.GetInventory()) {
				obj.logger.ValidationError(validation.E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}

	id := inv.GetID()
	digestAlg := inv.GetDigestAlgorithm()
	versions := inv.GetVersions()
	for ver, verInventory := range versionInventories {
		// check for id consistency
		if id != verInventory.GetID() {
			obj.logger.ValidationError(validation.E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", id, ver, verInventory.GetID())
		}
		if verInventory.GetHead().IsValid() && verInventory.GetHead().String() != ver {
			obj.logger.ValidationError(validation.E040, "wrong head '%s' in manifest for version '%s'", verInventory.GetHead(), ver)
		}

		if verInventory.GetDigestAlgorithm() != digestAlg {
			obj.logger.ValidationError(validation.W000, "different digest algorithm '%s' in version '%s'", verInventory.GetDigestAlgorithm(), ver)
		}

		for versionNumber, vVersion := range verInventory.GetVersions().Iterate() {
			testV := versions.GetVersion(versionNumber)
			if testV == nil {
				obj.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not in object root manifest", vVersion, versionNumber)
			}
			if !testV.Equals(vVersion) {
				obj.logger.ValidationError(validation.E066, "version '%s' in version folder '%s' not equal to version in object root manifest", vVersion, versionNumber)
			}
		}
	}

	//
	// all files in any manifest must belong to a physical file #E092
	//
	for inventoryVersion, inv2 := range versionInventories {
		for manifestFile := range inv2.GetManifest().GetFilesFlat() {
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

	//
	// all object content files must belong to manifest
	//

	latestVersion := inventory.NewVersionNumber()

	for objectContentVersion, objectContentVersionFiles := range objectContentFiles {
		objectContentVersionNumber := inventory.NewVersionNumber().WithString(objectContentVersion)
		if !latestVersion.IsValid() {
			latestVersion = objectContentVersionNumber
		}
		if latestVersion.Less(objectContentVersionNumber) {
			latestVersion = objectContentVersionNumber
		}
		// check version inventories
		for inventoryVersion, versionInventory := range versionInventories {
			inventoryVersionNumber := inventory.NewVersionNumber().WithString(inventoryVersion)
			if objectContentVersionNumber.Less(inventoryVersionNumber) {
				versionManifestFiles := util.SeqToSlice(versionInventory.GetManifest().GetFilesFlat())
				for _, objectContentVersionFile := range objectContentVersionFiles {
					// check all inventories which are less in version
					if !slices.Contains(versionManifestFiles, objectContentVersionFile) {
						obj.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, inventoryVersion)
					}
				}
			}
		}
		rootVersion := inv.GetHead()
		if objectContentVersionNumber.Less(rootVersion) {
			rootManifestFiles := util.SeqToSlice(inv.GetManifest().GetFilesFlat())
			for _, objectContentVersionFile := range objectContentVersionFiles {
				// check all inventories which are less in version
				if !slices.Contains(rootManifestFiles, objectContentVersionFile) {
					obj.logger.ValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, rootVersion)
				}
			}
		}
	}

	return nil
}

func (obj *validator) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	inv := obj.GetInventory()
	// get all possible digest algs
	digestAlgorithms := append(util.SeqToSlice(inv.GetFixity().GetDigestAlgorithms()), inv.GetDigestAlgorithm())

	result := map[checksum.DigestAlgorithm]map[string][]string{}
	//	versionNumbers := ocfl.SeqToSlice(inv.GetVersions().GetVersionNumbers())
	entries, err := fs.ReadDir(obj.GetReadFS(), ".")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read dir '%v'", obj.GetReadFS())
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "extensions" {
			continue
		}
		versionNumber := inventory.NewVersionNumber().WithString(entry.Name())
		if versionNumber.Int() == 0 {
			obj.logger.ValidationError(validation.E012, "folder '%v' is not a version", entry.Name())
			continue
		}
		if err := fs.WalkDir(
			obj.GetReadFS(),
			//fmt.Sprintf("%s/%s", version, inv.GetContentDir()),
			path.Join(versionNumber.String(), inv.GetContentDir()),
			func(path string, d fs.DirEntry, dirErr error) error {
				if dirErr != nil {
					return errors.Wrapf(dirErr, "error walking into %s", path)
				}
				//obj.logger.Debug(path)
				if d == nil || d.IsDir() {
					return nil
				}
				fname := path // filepath.ToSlash(filepath.Join(version, path))
				fp, err := obj.GetReadFS().Open(fname)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%s'", fname)
				}
				defer fp.Close()
				css, err := checksum.Copy(digestAlgorithms, fp, &checksum.NullWriter{})
				if err != nil {
					return errors.Wrapf(err, "cannot read and create checksums for file '%s'", fname)
				}
				for d, cs := range css {
					if _, ok := result[d]; !ok {
						result[d] = map[string][]string{}
					}
					if _, ok := result[d][cs]; !ok {
						result[d][cs] = []string{}
					}
					result[d][cs] = append(result[d][cs], fname)
				}
				obj.logger.Debug().Msgf("calculated %d checksums for file '%s'", len(css), fname)
				return nil
			}); err != nil {
			return nil, errors.Wrapf(err, "cannot walk content dir '%s'", inv.GetContentDir())
		}
	}
	return result, nil
}

var _ object.Validator = (*validator)(nil)
