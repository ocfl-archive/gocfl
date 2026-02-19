package inventoryimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/uri"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"golang.org/x/exp/slices"
)

func NewInventoryBase(ctx context.Context, f factorytypes.Factory, version version.OCFLVersion, spec inventory.InventorySpec, logger ocfllogger.OCFLLogger) *InventoryBase {
	i := &InventoryBase{
		ctx:     ctx,
		factory: f,
		//object:                 object,
		version: version,
		//folder:  objectFolder,
		//paddingLength: 0,
		//fixityDigestAlgorithms: []checksum.DigestAlgorithm{},
		Type:             spec,
		Head:             inventory.NewVersionNumber(),
		ContentDirectory: "content",
		Manifest:         f.NewManifest(ctx),
		Versions:         f.NewVersions(ctx),
		Fixity:           f.NewFixity(ctx),
		logger:           logger,
	}
	return i

}

type InventoryBase struct {
	factory factorytypes.Factory
	ctx     context.Context
	//folder  string
	//object                 ocfl.Object
	version   version.OCFLVersion
	modified  bool
	writeable bool
	//paddingLength int
	//versionValue           map[string]uint
	//fixityDigestAlgorithms []checksum.DigestAlgorithm
	Id               string                   `json:"id"`
	Type             inventory.InventorySpec  `json:"type"`
	DigestAlgorithm  checksum.DigestAlgorithm `json:"digestAlgorithm"`
	Head             *inventory.VersionNumber `json:"head"` // caveat: only for json marshalling
	ContentDirectory string                   `json:"contentDirectory,omitempty"`
	Manifest         inventory.Manifest       `json:"manifest,omitempty"`
	Versions         inventory.Versions       `json:"versions,omitempty"`
	Fixity           inventory.Fixity         `json:"fixity,omitempty"`
	logger           ocfllogger.OCFLLogger
}

func (i *InventoryBase) GetOCFLVersion() version.OCFLVersion {
	return i.version
}

func (i *InventoryBase) Bytes() (inventoryBytes []byte, checksumString string, err error) {
	inventoryBytes, err = json.MarshalIndent(i, "", "   ")
	if err != nil {
		return nil, "", errors.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(i.GetDigestAlgorithm())
	if err != nil {
		return nil, "", errors.Wrapf(err, "invalid digest algorithm '%s'", string(i.GetDigestAlgorithm()))
	}
	if _, err := h.Write(inventoryBytes); err != nil {
		return nil, "", errors.Wrapf(err, "cannot create checksum of manifest")
	}
	checksumBytes := h.Sum(nil)
	checksumString = fmt.Sprintf("%x", checksumBytes)
	return

}

func (i *InventoryBase) WithVersions(versions inventory.Versions) inventory.Inventory {
	i.Versions = versions
	return i
}
func (i *InventoryBase) GetVersions() inventory.Versions {
	return i.Versions
}

func (i *InventoryBase) WithManifest(manifest inventory.Manifest) inventory.Inventory {
	i.Manifest = manifest
	return i
}
func (i *InventoryBase) GetManifest() inventory.Manifest {
	return i.Manifest
}

func (i *InventoryBase) WithFixity(fixity inventory.Fixity) inventory.Inventory {
	i.Fixity = fixity
	return i
}
func (i *InventoryBase) GetFixity() inventory.Fixity {
	/*
		if i.Fixity == nil {
			return map[checksum.DigestAlgorithm]map[string][]string{}
		}

	*/
	return i.Fixity
}

func (i *InventoryBase) WithID(id string) inventory.Inventory {
	i.Id = id
	return i
}
func (i *InventoryBase) GetID() string { return i.Id }

func (i *InventoryBase) WithDigestAlgorithm(algorithm checksum.DigestAlgorithm) inventory.Inventory {
	i.DigestAlgorithm = algorithm
	return i
}
func (i *InventoryBase) GetDigestAlgorithm() checksum.DigestAlgorithm { return i.DigestAlgorithm }

func (i *InventoryBase) WithContentDir(contentDir string) inventory.Inventory {
	if contentDir == "" {
		contentDir = "content"
	}
	i.ContentDirectory = contentDir
	return i
}
func (i *InventoryBase) GetContentDir() string {
	if i.ContentDirectory == "" {
		return "content"
	}
	return i.ContentDirectory
}
func (i *InventoryBase) GetRealContentDir() string {
	return i.ContentDirectory
}

func (i *InventoryBase) Equals(invent inventory.Inventory) bool {

	i2, ok := invent.(*InventoryBase)
	if !ok {
		return false
	}

	if i.Type != i2.Type {
		return false
	}
	if !i.Head.Equal(i2.Head) {
		return false
	}
	if i.ContentDirectory != i2.ContentDirectory {
		return false
	}
	if (i.Manifest == nil && i2.Manifest != nil) || (i.Manifest != nil && i2.Manifest == nil) {
		return false
	}
	if i.Manifest != nil {
		if !i.Manifest.Equals(invent.GetManifest()) {
			return false
		}
	}
	if (i.Versions == nil && i2.Versions != nil) || (i.Versions != nil && i2.Versions == nil) {
		return false
	}
	if i.Versions != nil {
		if !i.Versions.Equals(i2.Versions) {
			return false
		}
	}
	return true
}

func (i *InventoryBase) Finalize(inCreation bool) (err error) {
	if i.Manifest == nil {
		if !inCreation {
			i.AddValidationError(validation.E041, "no manifest in inventory")
		}
		i.Manifest = i.factory.NewManifest(nil)
	}
	if err := i.Manifest.Finalize(i, i.factory, inCreation); err != nil {
		return errors.Wrap(err, "error finalizing manifest")
	}

	if i.Versions == nil {
		if !inCreation {
			i.AddValidationError(validation.E041, "no versions in inventory")
		}
		i.Versions = i.factory.NewVersions(nil)
	}
	if err := i.Versions.Finalize(i, i.factory, inCreation); err != nil {
		return errors.Wrap(err, "error finalizing versions")
	}

	if err := i.Fixity.Finalize(inCreation); err != nil {
		return errors.Wrap(err, "error finalizing fixity")
	}

	if !inCreation {
		if err := i.check(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (i *InventoryBase) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	err := validation.GetValidationError(i.version, errno).AppendDescription(format, a...).AppendContext("object '%s'", i.GetID())
	return errors.WithStack(validation.AddValidationErrors(i.ctx, err))
}
func (i *InventoryBase) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	err := validation.GetValidationError(i.version, errno).AppendDescription(format, a...).AppendContext("object '%s'", i.GetID())
	return errors.WithStack(validation.AddValidationWarnings(i.ctx, err))
}
func (i *InventoryBase) GetHead() *inventory.VersionNumber {
	return i.GetVersions().LatestVersionNumber()
}
func (i *InventoryBase) GetSpec() inventory.InventorySpec { return i.Type }

func (i *InventoryBase) WithWriteable() inventory.Inventory {
	i.writeable = true
	return i
}
func (i *InventoryBase) IsWriteable() bool { return i.writeable }
func (i *InventoryBase) IsModified() bool  { return i.modified }

func (i *InventoryBase) IterateFiles(version *inventory.VersionNumber, fn inventory.StateFileCallback) error {
	if !version.IsValid() {
		version = i.GetHead()
	}
	ver := i.Versions.GetVersion(version)
	if ver == nil {
		return errors.Errorf("invalid version '%s'", version)
	}
	state := ver.GetState()
	if state == nil {
		return errors.Errorf("cannot get state for '%s'", version)
	}
	for digest, externalNames := range state.Iterate() {
		internalNames, err := i.Manifest.GetFiles(digest)
		if err != nil {
			return errors.Wrapf(err, "no manifest for [%s]%v", digest, externalNames)
		}
		if len(internalNames) == 0 {
			return errors.Errorf("invalid manifest for digest [%s]", digest)
		}
		if err := errors.WithStack(fn(internalNames, externalNames, digest)); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (i *InventoryBase) check() error {
	var manifestDigest = []string{}
	var alg = i.GetDigestAlgorithm()
	var fileManifestDigest = map[checksum.DigestAlgorithm]map[string][]string{alg: {}}
	for digest, files := range i.GetManifest().Iterate() {
		manifestDigest = append(manifestDigest, digest)
		fileManifestDigest[alg] = map[string][]string{digest: files}
	}
	if err := i.GetVersions().Check(i, manifestDigest); err != nil {
		return errors.WithStack(err)
	}
	versionDigests := []string{}
	for versionNumber, ver := range i.Versions.Iterate() {
		state := ver.GetState()
		if state == nil {
			return errors.Errorf("cannot get state for version '%s'", versionNumber)
		}
		for digest, _ := range state.Iterate() {
			versionDigests = append(versionDigests, digest)
		}
	}

	/*
		todo: repair this
			if err := i.GetManifest().Check(i, nil, versionDigests); err != nil {
				return errors.WithStack(err)
			}
	*/
	if err := i.GetFixity().Check(i, fileManifestDigest); err != nil {
		return errors.WithStack(err)
	}
	if i.Id == "" {
		i.AddValidationError(validation.E036, "invalid field \"id\" for object")
	}
	if i.Id != "" {
		if _, err := uri.Parse(i.Id); err != nil {
			i.AddValidationWarning(validation.W005, "cannot parse uri id '%s': %v", i.Id, err)
		} /* else {
			if u.Scheme == "" {
				i.AddValidationWarning(W005, "id '%s' is not an uri", i.Id)
			}
		}
		*/
	}
	if !i.Head.IsValid() {
		i.AddValidationError(validation.E040, "invalid field \"head\" for object")
	}
	if i.Type == "" {
		i.AddValidationError(validation.E036, "invalid field \"type\" for object")
	}
	if i.DigestAlgorithm == "" {
		i.AddValidationError(validation.E036, "invalid field \"digestAlgorithm\" for object")
	}

	if !slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA512, checksum.DigestSHA256}, i.DigestAlgorithm) {
		i.AddValidationError(validation.E025, "invalid digest algorithm '%s'", i.DigestAlgorithm)
	} else {
		if slices.Contains([]checksum.DigestAlgorithm{checksum.DigestSHA256}, i.DigestAlgorithm) {
			i.AddValidationError(validation.W004, "digest algorithm '%s' not suggested", i.DigestAlgorithm)
		}
	}

	if i.ContentDirectory != "" {
		if slices.Contains([]string{"", ".", ".."}, i.ContentDirectory) || strings.Contains(i.ContentDirectory, "/") {
			i.AddValidationError(validation.E017, "invalid content directory '%s'", i.ContentDirectory)
		}
	}

	return nil
}

func (i *InventoryBase) CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error {
	i.logger.Debug().Msgf("[%s] checkFiles", i.GetID())
	defer i.logger.Debug().Msgf("[%s] checkFiles done", i.GetID())
	csFiles, ok := fileManifest[i.GetDigestAlgorithm()]
	if !ok {
		if len(fileManifest) == 0 {
			return nil
		}
		return errors.Errorf("checksum for '%s' not created", i.GetDigestAlgorithm())
	}
	_ = csFiles
	/*
		// todo: repair this
			if err := i.Manifest.Check(i, csFiles, nil); err != nil {
				return errors.Wrap(err, "manifest check failed")
			}
	*/
	if err := i.Fixity.Check(i, fileManifest); err != nil {
		return errors.Wrap(err, "fixity check failed")
	}
	return nil
}

func (i *InventoryBase) BuildManifestName(stateFilename string) string {
	return i.BuildManifestNameVersion(stateFilename, i.GetHead())
}

func (i *InventoryBase) BuildManifestNameVersion(stateFilename string, version *inventory.VersionNumber) string {
	return filepath.ToSlash(filepath.Clean(filepath.Join(version.String(), i.GetContentDir(), stateFilename)))
}

var vRegexp *regexp.Regexp = regexp.MustCompile("^v(\\d+)$")

func (i *InventoryBase) AlreadyExists(stateFilename, checksum string) (bool, error) {
	return i.Versions.FileExists(stateFilename, checksum)
}

func (i *InventoryBase) IsUpdate(virtualFilename, checksum string) (bool, error) {
	found, err := i.Versions.FileExists(virtualFilename, checksum)
	return !found, errors.WithStack(err)
}

func (i *InventoryBase) EchoDelete(existing []string, pathPrefix string) error {
	modified, err := i.Versions.EchoDelete(existing, pathPrefix)
	if err != nil {
		return errors.WithStack(err)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) DeleteFile(stateFilename string) error {
	modified, err := i.Versions.DeleteFile(stateFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", stateFilename)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) RenameFile(stateSource, stateDest string) error {
	modified, err := i.Versions.RenameFile(stateSource, stateDest)
	if err != nil {
		return errors.Wrapf(err, "cannot rename '%s' to '%s'", stateSource, stateDest)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) CopyFile(dest string, digest string) error {
	i.logger.Info().Msgf("[%s] copying '%s' -> '%s'", i.GetID(), digest, dest)

	modified, err := i.Versions.CopyFile(dest, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot copy '%s' to '%s'", digest, dest)
	}
	i.modified = i.modified || modified
	return nil
}

func (i *InventoryBase) AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error {
	i.logger.Debug().Msgf("[%s] adding '%s' -> '%s'", i.GetID(), stateFilenames, manifestFilename)
	digest, ok := checksums[i.GetDigestAlgorithm()]
	if !ok {
		return errors.Errorf("no digest for '%s' in checksums", i.GetDigestAlgorithm())
	}
	digest = strings.ToLower(digest) // paranoia

	fixitydigests := map[checksum.DigestAlgorithm]string{}
	for alg, cs := range checksums {
		if alg == i.GetDigestAlgorithm() {
			continue
		}
		fixitydigests[alg] = cs
	}
	modified, err := i.Fixity.AddFile(manifestFilename, fixitydigests)
	if err != nil {
		return errors.Wrapf(err, "cannot add fixity '%s' to '%s'", digest, manifestFilename)
	}
	i.modified = i.modified || modified

	if manifestFilename != "" {
		modified, err := i.Manifest.AddFile(manifestFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add manifest '%s' to '%s'", digest, manifestFilename)
		}
		i.modified = i.modified || modified
	}

	for _, virtualFilename := range stateFilenames {
		dup, err := i.AlreadyExists(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add for duplicate of '%s' [%s]", stateFilenames, digest)
		}
		if dup {
			i.logger.Debug().Msgf("'%s' is a duplicate", stateFilenames)
			// return nil
		}

		modfied, err := i.Versions.AddFile(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot add state '%s' to '%s'", digest, virtualFilename)
		}
		i.modified = i.modified || modfied

		upd, err := i.IsUpdate(virtualFilename, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check for update of '%s' [%s]", stateFilenames, digest)
		}
		if upd {
			i.logger.Debug().Msgf("'%s' is an update - removing old version", stateFilenames)
			if err := i.DeleteFile(virtualFilename); err != nil {
				return errors.Wrapf(err, "cannot delete old version of '%s' [%s]", stateFilenames, digest)
			}
			i.modified = true
		}

		if !dup {
			modified, err := i.Versions.AddFile(virtualFilename, digest)
			if err != nil {
				return errors.Wrapf(err, "cannot add version of '%s' [%s]", stateFilenames, digest)
			}
			i.modified = i.modified || modified
		}
	}

	return nil
}

// clear unmodified version
func (i *InventoryBase) Clean() error {
	i.logger.Debug()
	// read only means nothing to do
	if i.IsModified() {
		return nil
	}
	// only one version. could be empty
	if i.GetHead().String() == "v1" {
		return nil
	}
	i.logger.Debug().Msgf("deleting %v", i.GetHead())
	i.Versions.Delete(i.GetHead())
	lastVersion := i.Versions.LatestVersionNumber()
	if !lastVersion.IsValid() {
		return errors.New("cannot get last version")
	}
	i.Head = lastVersion
	return nil
}

func (i *InventoryBase) MarshalJSON() ([]byte, error) {
	i.Head = i.GetHead()
	type Alias InventoryBase
	return json.Marshal((*Alias)(i))
}

var _ inventory.Inventory = (*InventoryBase)(nil)
var _ validation.Validation = (*InventoryBase)(nil)
