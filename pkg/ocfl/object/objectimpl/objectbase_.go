package objectimpl

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"golang.org/x/exp/slices"
)

//const VERSION = "1.0"

//var objectConformanceDeclaration = fmt.Sprintf("0=ocfl_object_%s", VERSION)

// todo: check WithWriteable() and repair incorrect use...

var versionRegexp = regexp.MustCompile("^v(\\d+)/$")

//var inventoryDigestRegexp = regexp.MustCompile(fmt.Sprintf("^(?i)inventory\\.json\\.(%s|%s)$", string(checksum.DigestSHA512), string(checksum.DigestSHA256)))

/*
	func (objectBase *ObjectBase) WithFS(fsys fs.FS) object.Object {
		objectBase.fsys = fsys
		return objectBase
	}
*/
func (objectBase *ObjectBase) GetExtensionManager() object.ExtensionManager {
	return objectBase.extensionManager
}

func (objectBase *ObjectBase) IsModified() bool { return objectBase.i.IsModified() }

func (objectBase *ObjectBase) AddValidationError(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(objectBase.version, errno).AppendDescription(format, a...).AppendContext("object '%s'", objectBase.i.GetID())
	_, file, line, _ := runtime.Caller(1)
	objectBase.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationErrors(objectBase.ctx, valError))
}

func (objectBase *ObjectBase) AddValidationWarning(errno validation.ValidationErrorCode, format string, a ...any) error {
	valError := validation.GetValidationError(objectBase.version, errno).AppendDescription(format, a...).AppendContext("object '%s'", objectBase.i.GetID())
	_, file, line, _ := runtime.Caller(1)
	objectBase.logger.Debug().Msgf("[%s:%v] %s", file, line, valError.Error())
	return errors.WithStack(validation.AddValidationWarnings(objectBase.ctx, valError))
}

func (objectBase *ObjectBase) GetMetadata() (*inventory.Metadata, error) {
	inv := objectBase.i
	if inv == nil {
		return nil, errors.Errorf("inventory is nil")
	}

	result := &inventory.Metadata{
		ID:              objectBase.i.GetID(),
		Head:            inv.GetHead(),
		Files:           map[string]*inventory.FileMetadata{},
		DigestAlgorithm: objectBase.i.GetDigestAlgorithm(),
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
	extensionMetadata, err := objectBase.extensionManager.GetMetadata(objectBase)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get extension metadata for object '%s'", objectBase.i.GetID())
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

func (objectBase *ObjectBase) Stat(w io.Writer, statInfo []object.StatInfo) error {
	fmt.Fprintf(w, "[%s] Path: %s\n", objectBase.i.GetID(), objectBase.i.GetDigestAlgorithm())
	i := objectBase.i
	fmt.Fprintf(w, "[%s] Head: %s\n", objectBase.i.GetID(), i.GetHead())
	fixity := i.GetFixity()
	algs := []string{}
	for alg := range fixity.GetDigestAlgorithms() {
		algs = append(algs, string(alg))
	}
	fmt.Fprintf(w, "[%s] Fixity: %s\n", objectBase.i.GetID(), strings.Join(algs, ", "))
	manifest := i.GetManifest()
	cnt := 0
	for _, fs := range manifest.Iterate() {
		cnt += len(fs)
	}

	var uniqueFileCount int
	for _, _ = range manifest.Iterate() {
		uniqueFileCount++
	}
	fmt.Fprintf(w, "[%s] Manifest: %v files (%v unique files)\n", objectBase.i.GetID(), cnt, uniqueFileCount)
	if slices.Contains(statInfo, object.StatObjectVersions) || len(statInfo) == 0 {
		for vString, ver := range i.GetVersions().Iterate() {
			fmt.Fprintf(w, "[%s] Version %s\n", objectBase.i.GetID(), vString)
			fmt.Fprintf(w, "[%s]     User: %s (%s)\n", objectBase.i.GetID(), ver.GetUser().GetName(), ver.GetUser().GetAddress())
			fmt.Fprintf(w, "[%s]     Created: %s\n", objectBase.i.GetID(), ver.GetCreated().String())
			fmt.Fprintf(w, "[%s]     Message: %s\n", objectBase.i.GetID(), ver.GetMessage())
			if slices.Contains(statInfo, object.StatObjectVersionState) || len(statInfo) == 0 {
				for cs, sList := range ver.GetState().Iterate() {
					for _, s := range sList {
						fmt.Fprintf(w, "[%s]        %s\n", objectBase.i.GetID(), s)
						if slices.Contains(statInfo, object.StatObjectManifest) || len(statInfo) == 0 {
							ms, err := manifest.GetFiles(cs)
							if err != nil {
								if errors.Is(err, inventory.DigestNotFound) {
									continue
								}
								return errors.Wrapf(err, "cannot get files for manifest '%s'", objectBase.i.GetID())
							}
							for _, m := range ms {
								fmt.Fprintf(w, "[%s]           %s\n", objectBase.i.GetID(), m)
							}
						}
					}
				}
			}
		}
	}
	if slices.Contains(statInfo, object.StatObjectExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(objectBase.extensionManager.GetConfig(), "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		fmt.Fprintf(w, "[%s] Initial Extension:\n---\n%s\n---\n", objectBase.i.GetID(), string(data))
		fmt.Fprintf(w, "[%s] Extension Configurations:\n", objectBase.i.GetID())
		for _, ext := range objectBase.extensionManager.GetExtensions() {
			cfg := ext.GetConfig()
			str, _ := json.MarshalIndent(cfg, "", "  ")

			fmt.Fprintf(w, "---\n%s\n", str)
		}
	}
	return nil
}

/*
func (objectBase *ObjectBase) GetFS() fs.FS {
	return objectBase.fsys
}
*/

func (objectBase *ObjectBase) CreateInventory(id string, digestAlg checksum.DigestAlgorithm, fixityAlgs []checksum.DigestAlgorithm) (inventory.Inventory, error) {
	fixity := objectBase.factory.NewFixity(objectBase.ctx).WithAlgorithms(fixityAlgs...)
	inventory := objectBase.factory.NewInventory(objectBase.ctx).
		WithID(id).
		WithDigestAlgorithm(digestAlg).
		WithFixity(fixity)

	/*
		inventory, err := inventory.NewInventory(objectBase.ctx, "new", objectBase.GetOCFLVersion(), objectBase.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
	/*
		if err := inventory.Init(id, digest, fixity); err != nil {
			return nil, errors.Wrap(err, "cannot initialize empty inventory")
		}
	*/

	return inventory, inventory.Finalize(true)
}
func (objectBase *ObjectBase) GetInventory() inventory.Inventory {
	return objectBase.i
}

func (objectBase *ObjectBase) StoreInventory(version bool, objectRoot bool) error {
	if objectBase.fsys == nil {
		return errors.Errorf("read only filesystem '%v'", objectBase.fsys)
	}
	objectBase.logger.Debug()

	// check whether object filesystem is writeable
	if !objectBase.i.IsWriteable() {
		return errors.New("inventory not writeable - not updated")
	}

	// create inventory.json from inventory
	iFileName := "inventory.json"
	jsonBytes, err := json.MarshalIndent(objectBase.i, "", "   ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(objectBase.i.GetDigestAlgorithm())
	if err != nil {
		return errors.Wrapf(err, "invalid digest algorithm '%s'", string(objectBase.i.GetDigestAlgorithm()))
	}
	if _, err := h.Write(jsonBytes); err != nil {
		return errors.Wrapf(err, "cannot create checksum of manifest")
	}
	checksumBytes := h.Sum(nil)
	checksumString := fmt.Sprintf("%x %s", checksumBytes, iFileName)

	if objectRoot {
		iWriter, err := writefs.Create(objectBase.fsys, iFileName)
		if err != nil {
			iWriter.Close()
			return errors.Wrap(err, "cannot create inventory.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			return errors.Wrap(err, "cannot write to inventory.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", objectBase.fsys, iFileName)
		}
		csFileName := fmt.Sprintf("inventory.json.%s", string(objectBase.i.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(objectBase.fsys, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", objectBase.fsys, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			iCSWriter.Close()
			return errors.Wrapf(err, "cannot write to '%v/%s'", objectBase.fsys, csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", objectBase.fsys, csFileName)
		}
	}
	if version {
		iFileName = fmt.Sprintf("%s/inventory.json", objectBase.i.GetHead())
		iWriter, err := writefs.Create(objectBase.fsys, iFileName)
		if err != nil {
			return errors.Wrap(err, "cannot create inventory.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			iWriter.Close()
			return errors.Wrap(err, "cannot write to inventory.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", objectBase.fsys, iFileName)
		}
		csFileName := fmt.Sprintf("%s/inventory.json.%s", objectBase.i.GetHead(), string(objectBase.i.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(objectBase.fsys, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", objectBase.fsys, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			iCSWriter.Close()
			return errors.Wrapf(err, "cannot write to '%s'", csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", objectBase.fsys, csFileName)
		}
	}
	return nil
}

func (objectBase *ObjectBase) StoreExtensions() error {
	objectBase.logger.Debug()

	if err := objectBase.extensionManager.WriteConfig(nil); err != nil {
		return errors.Wrap(err, "cannot store extension configs")
	}
	return nil
}

func (objectBase *ObjectBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return objectBase.i.GetDigestAlgorithm()
}

func (objectBase *ObjectBase) echoDelete() error {
	slices.Sort(objectBase.updateFiles)
	objectBase.updateFiles = slices.Compact(objectBase.updateFiles)
	basePath, err := objectBase.extensionManager.BuildObjectStatePath(objectBase, ".", "")
	if err != nil {
		return errors.Wrap(err, "cannot build external path for '.'")
	}
	if basePath == "." {
		basePath = ""
	}
	version := objectBase.i.GetVersions().GetVersion(objectBase.i.GetHead())
	if _, err := version.EchoDelete(objectBase.updateFiles, basePath); err != nil {
		return errors.Wrap(err, "cannot remove deleted files from inventory")
	}
	return nil
}

func (objectBase *ObjectBase) Close() error {
	objectBase.logger.Info().Msgf(fmt.Sprintf("Closing object '%s'", objectBase.i.GetID()))
	if !(objectBase.i.IsWriteable()) {
		return nil
	}

	if !objectBase.i.IsModified() {
		return nil
	}
	//object.storageRoot.setModified()
	if err := objectBase.i.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := objectBase.StoreInventory(false, true); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}
	if err := objectBase.StoreExtensions(); err != nil {
		return errors.Wrap(err, "cannot store extensions")
	}
	return nil
}

func (objectBase *ObjectBase) StartUpdate(sourceFS fs.FS, msg string, UserName string, UserAddress string, echo bool) (fs.FS, error) {
	objectBase.logger.Debug().Msgf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)
	objectBase.echo = echo

	subfs, err := writefs.SubFSCreate(objectBase.fsys, "extensions")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", objectBase.fsys, "extensions")
	}
	// todo: bad style, changes extensionManager in factory
	objectBase.extensionManager.SetFS(subfs, true)

	if err := objectBase.i.GetVersions().NewVersion(objectBase.i.GetHead(), msg, UserName, UserAddress); err != nil {
		return nil, errors.Wrap(err, "cannot create new object version")
	}
	if err := objectBase.extensionManager.UpdateObjectBefore(objectBase); err != nil {
		return nil, errors.Wrapf(err, "cannot execute ext.UpdateObjectBefore()")
	}
	var versionFS fs.FS
	return versionFS, nil
}

func (objectBase *ObjectBase) EndUpdate() error {
	objectBase.logger.Info().Msgf(fmt.Sprintf("EndUpdate of object '%s'", objectBase.i.GetID()))
	if !(objectBase.i.IsWriteable()) {
		objectBase.logger.Warn().Msgf(fmt.Sprintf("object '%s' not writeable", objectBase.i.GetID()))
		return nil
	}
	if !(objectBase.i.IsModified()) {
		objectBase.logger.Info().Msgf(fmt.Sprintf("object '%s' not modified", objectBase.i.GetID()))
		return nil
	}

	if objectBase.echo {
		if err := objectBase.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot delete files")
		}
	}
	if err := objectBase.extensionManager.UpdateObjectAfter(objectBase); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
	}

	if err := objectBase.i.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := objectBase.StoreInventory(true, false); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}

	if needVersion, err := objectBase.extensionManager.NeedNewVersion(objectBase); err != nil {
		return errors.Wrapf(err, "cannot execute ext.NeedNewVersion()")
	} else if needVersion {
		if _, err := objectBase.StartUpdate(nil, "automated version", "gocfl", "https://github.com/ocfl-archive/gocfl", false); err != nil {
			return errors.Wrap(err, "cannot create new version")
		}
		if err := objectBase.extensionManager.DoNewVersion(objectBase); err != nil {
			return errors.Wrapf(err, "cannot execute ext.DoNewVersion()")
		}
		/*
			if err := objectBase.extensionManager.UpdateObjectAfter(objectBase); err != nil {
				return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
			}
		*/
		if err := objectBase.EndUpdate(); err != nil {
			return errors.Wrap(err, "cannot end update")
		}
	}
	return nil
}

func (objectBase *ObjectBase) BeginArea(area string) {
	objectBase.area = area
	objectBase.updateFiles = []string{}
}

func (objectBase *ObjectBase) EndArea() error {
	if objectBase.echo {
		if err := objectBase.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot remove files")
		}
	}
	objectBase.updateFiles = []string{}
	objectBase.area = ""
	return nil
}

func (objectBase *ObjectBase) AddFolder(fsys fs.FS, versionFS fs.FS, checkDuplicate bool, area string) error {
	objectBase.logger.Debug().Msgf("walking '%v'", fsys)
	if err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		if info.Name() == "." {
			return nil
		}
		path = filepath.ToSlash(path)
		if err := objectBase.AddFile(fsys, versionFS, path, checkDuplicate, area, false, info.IsDir()); err != nil {
			return errors.Wrapf(err, "cannot add file '%s'", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (objectBase *ObjectBase) addReader(r io.ReadCloser, versionFS fs.FS, names *object.NamesStruct, noExtensionHook bool) (string, error) {

	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range objectBase.i.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}

	var digest string

	objectBase.updateFiles = append(objectBase.updateFiles, names.ExternalPaths...)

	if !slices.Contains(digestAlgorithms, objectBase.i.GetDigestAlgorithm()) {
		digestAlgorithms = append(digestAlgorithms, objectBase.i.GetDigestAlgorithm())
	}

	writer, err := writefs.Create(objectBase.fsys, names.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "cannot create '%s'", names.ManifestPath)
	}
	defer writer.Close()

	var checksums map[checksum.DigestAlgorithm]string
	if noExtensionHook {
		checksums, err = checksum.Copy(digestAlgorithms, r, writer)
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%v' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		pr, pw := io.Pipe()
		extErrors := make(chan error, 1)
		go func() {
			defer wg.Done()
			if err := objectBase.extensionManager.StreamObject(objectBase, pr, names.ExternalPaths, names.InternalPath); err != nil {
				extErrors <- err
			}
		}()
		checksums, err = checksum.Copy(digestAlgorithms, r, writer, pw)
		if err := pw.Close(); err != nil {
			objectBase.logger.Error().Err(err).Msg("cannot close pipe writer")
		}
		wg.Wait()
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%s' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
		close(extErrors)
		select {
		case err, ok := <-extErrors:
			if ok {
				return "", errors.Wrapf(err, "error on StreamObject() extension hook for object '%s'", objectBase.i.GetID())
			}
		default:
		}
	}

	if digest == "" {
		var ok bool
		digest, ok = checksums[objectBase.i.GetDigestAlgorithm()]
		if !ok {
			return "", errors.Errorf("digest '%s' not generated", objectBase.i.GetDigestAlgorithm())
		}
	} else {
		checksums[objectBase.i.GetDigestAlgorithm()] = digest
	}
	if err := objectBase.i.AddFile(names.ExternalPaths, names.ManifestPath, checksums); err != nil {
		return "", errors.Wrapf(err, "cannot append '%v'/'%s' to inventory", names.ExternalPaths, names.InternalPath)
	}

	return digest, nil
}

func (objectBase *ObjectBase) BuildNames(files []string, area string) (*object.NamesStruct, error) {
	var err error
	result := &object.NamesStruct{
		ExternalPaths: []string{},
	}
	for _, file := range files {
		externalPath, err := objectBase.extensionManager.BuildObjectStatePath(objectBase, file, area)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", file)
		}
		result.ExternalPaths = append(result.ExternalPaths, externalPath)
	}
	result.InternalPath, err = objectBase.extensionManager.BuildObjectManifestPath(objectBase, files[0], area)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create manifest path for '%s'", files[0])
	}
	result.ManifestPath = objectBase.i.BuildManifestName(result.InternalPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", result.InternalPath)
	}
	return result, nil
}

func (objectBase *ObjectBase) AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error) {
	if len(files) == 0 {
		return "", errors.New("no files given")
	}
	if !objectBase.i.IsWriteable() {
		return "", errors.New("object not writeable")
	}
	path := files[0]
	names, err := objectBase.BuildNames(files, area)

	objectBase.logger.Info().Msgf("adding file %s:%v", area, files)

	if !noExtensionHook {
		if err := objectBase.extensionManager.AddFileBefore(objectBase, nil, path, names.InternalPath, area, false); err != nil {
			return "", errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var digest string
	if !isDir {
		digest, err = objectBase.addReader(r, nil, names, noExtensionHook)
		if err != nil {
			return "", errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	} else {
		io.Copy(io.Discard, r)
	}

	if !noExtensionHook {
		if err := objectBase.extensionManager.AddFileAfter(objectBase, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return "", errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return digest, nil
}

func (objectBase *ObjectBase) AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	if !objectBase.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	ver := objectBase.i.GetVersions().GetVersion(objectBase.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", objectBase.i.GetHead())
	}

	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range objectBase.i.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}
	var digest string

	names, err := objectBase.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot names for '%s'", path)

	}

	objectBase.logger.Info().Msgf("adding file %s:%s", area, path)

	newPath, err := objectBase.extensionManager.BuildObjectStatePath(objectBase, path, area)
	if err != nil {
		return errors.Wrapf(err, "cannot map external path '%s'", path)
	}

	objectBase.updateFiles = append(objectBase.updateFiles, newPath)

	var dataReader = bytes.NewReader(data)
	if checkDuplicate {
		// do the checksum
		digest, err = checksum.Checksum(dataReader, objectBase.i.GetDigestAlgorithm())
		if err != nil {
			return errors.Wrapf(err, "cannot create digest of '%s'", path)
		}
		// set filepointer to beginning
		if _, err := dataReader.Seek(0, 0); err != nil {
			return errors.Wrapf(err, "cannot seek in datareader")
		}
		// if file is already there we do nothing
		dup, err := objectBase.i.AlreadyExists(newPath, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
		}
		if dup {
			objectBase.logger.Info().Msgf("[%s] '%s' already exists. ignoring", objectBase.i.GetID(), newPath)
			return nil
		}
		// file already ingested, but new virtual name
		if dups, _ := objectBase.i.GetManifest().GetFiles(digest); len(dups) > 0 {
			objectBase.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", objectBase.i.GetID(), newPath)
			if _, err := ver.CopyFile(newPath, digest); err != nil {
				return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
			}
			return nil
		}
	} else {
		if !slices.Contains(digestAlgorithms, objectBase.i.GetDigestAlgorithm()) {
			digestAlgorithms = append(digestAlgorithms, objectBase.i.GetDigestAlgorithm())
		}
	}

	if !noExtensionHook {
		if err := objectBase.extensionManager.AddFileBefore(objectBase, nil, path, names.InternalPath, area, false); err != nil {
			return errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var r = io.NopCloser(dataReader)
	if !isDir {
		digest, err = objectBase.addReader(r, nil, names, noExtensionHook)
		if err != nil {
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	}

	if !noExtensionHook {
		if err := objectBase.extensionManager.AddFileAfter(objectBase, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (objectBase *ObjectBase) AddFile(fsys fs.FS, versionFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	objectBase.logger.Info().Msgf("adding file %s:%s", area, path)

	ver := objectBase.i.GetVersions().GetVersion(objectBase.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", objectBase.i.GetHead())
	}

	path = filepath.ToSlash(path)

	if !objectBase.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	names, err := objectBase.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
	}

	targetFilename := objectBase.i.BuildManifestName(names.InternalPath)

	var digest string
	if !isDir {

		digestAlgorithms := []checksum.DigestAlgorithm{}
		for alg := range objectBase.i.GetFixity().GetDigestAlgorithms() {
			digestAlgorithms = append(digestAlgorithms, alg)
		}

		file, err := fsys.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot open file '%v/%s'", fsys, path)
		}
		newPath, err := objectBase.extensionManager.BuildObjectStatePath(objectBase, path, area)
		if err != nil {
			file.Close()
			return errors.Wrapf(err, "cannot map external path '%s'", path)
		}

		objectBase.updateFiles = append(objectBase.updateFiles, newPath)

		if checkDuplicate {
			// do the checksum
			digest, err = checksum.Checksum(file, objectBase.i.GetDigestAlgorithm())
			if err != nil {
				return errors.Wrapf(err, "cannot create digest of '%s'", path)
			}
			// set filepointer to beginning
			if seeker, ok := file.(io.Seeker); ok {
				// if we have a seeker, we just seek
				if _, err := seeker.Seek(0, 0); err != nil {
					panic(err)
				}
			} else {
				// otherwise reopen it
				file, err = fsys.Open(path)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s'", fsys, path)
				}
			}
			// if file is already there we do nothing
			dup, err := objectBase.i.AlreadyExists(newPath, digest)
			if err != nil {
				return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
			}
			if dup {
				objectBase.logger.Info().Msgf("[%s] '%s' already exists. ignoring", objectBase.i.GetID(), newPath)
				return nil
			}
			// file already ingested, but new virtual name
			if dups, _ := objectBase.i.GetManifest().GetFiles(digest); len(dups) > 0 {
				objectBase.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", objectBase.i.GetID(), newPath)
				if _, err := ver.CopyFile(newPath, digest); err != nil {
					return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
				}
				return nil
			}
		} else {
			if !slices.Contains(digestAlgorithms, objectBase.i.GetDigestAlgorithm()) {
				digestAlgorithms = append(digestAlgorithms, objectBase.i.GetDigestAlgorithm())
			}
		}
		if !noExtensionHook {
			if err := objectBase.extensionManager.AddFileBefore(objectBase, nil, path, names.InternalPath, area, isDir); err != nil {
				return errors.Wrapf(err, "error on AddFileBefore() extension hook")
			}
		}

		digest, err = objectBase.addReader(file, versionFS, names, noExtensionHook)
		if err != nil {
			file.Close()
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
		if err := file.Close(); err != nil {
			return errors.Wrapf(err, "cannot close file '%s'", path)
		}

	}
	if !noExtensionHook {
		if err := objectBase.extensionManager.AddFileAfter(objectBase, fsys, []string{path}, targetFilename, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (objectBase *ObjectBase) DeleteFile(virtualFilename string, digest string) error {
	ver := objectBase.i.GetVersions().GetVersion(objectBase.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", objectBase.i.GetHead())
	}
	virtualFilename = filepath.ToSlash(virtualFilename)
	objectBase.logger.Debug().Msgf("removing '%s' [%s]", virtualFilename, digest)

	if !objectBase.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := objectBase.i.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilename, digest)
	}
	if !dup {
		objectBase.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilename, digest)
		return nil
	}
	if _, err := ver.DeleteFile(virtualFilename); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilename)
	}
	return nil

}

func (objectBase *ObjectBase) RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error {
	ver := objectBase.i.GetVersions().GetVersion(objectBase.i.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", objectBase.i.GetHead())
	}
	virtualFilenameSource = filepath.ToSlash(virtualFilenameSource)
	objectBase.logger.Debug().Msgf("removing '%s' [%s]", virtualFilenameSource, digest)

	if !objectBase.i.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := objectBase.i.AlreadyExists(virtualFilenameSource, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilenameSource, digest)
	}
	if !dup {
		objectBase.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilenameSource, digest)
		return nil
	}
	if _, err := ver.RenameFile(virtualFilenameSource, virtualFilenameDest); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilenameSource)
	}
	return nil

}

func (objectBase *ObjectBase) GetID() string {
	if objectBase.i == nil {
		return ""
	}
	return objectBase.i.GetID()
}

func (objectBase *ObjectBase) GetOCFLVersion() version.OCFLVersion {
	return objectBase.version
}

var allowedFilesRegexp = regexp.MustCompile("^(inventory.json(\\.sha512|\\.sha384|\\.sha256|\\.sha1|\\.md5)?|0=ocfl_object_[0-9]+\\.[0-9]+)$")

func (objectBase *ObjectBase) checkVersionFolder(version string) error {
	versionEntries, err := fs.ReadDir(objectBase.fsys, version)
	if err != nil {
		return errors.Wrapf(err, "cannot read version folder '%s'", version)
	}
	for _, ve := range versionEntries {
		if !ve.IsDir() {
			if !allowedFilesRegexp.MatchString(ve.Name()) {
				objectBase.AddValidationError(validation.E015, "extra file '%s' in version directory '%s'", ve.Name(), version)
			}
		}
	}
	return nil
}

func (objectBase *ObjectBase) checkFilesAndVersions() error {
	// create list of version content directories
	versionContents := map[string]string{}
	versionStrings := ocfl.SeqToSlice(objectBase.i.GetVersions().GetVersionNumbers())

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
		versionContents[ver.String()] = objectBase.i.GetContentDir()
	}

	// load object content files
	objectContentFiles := map[string][]string{}
	objectContentFilesFlat := []string{}
	objectFilesFlat := []string{}
	for ver, cont := range versionContents {
		// load all object version content files
		versionContent := ver + "/" + cont
		//inventoryFile := ver + "/inventory.json"
		if _, ok := objectContentFiles[ver]; !ok {
			objectContentFiles[ver] = []string{}
		}
		fs.WalkDir(
			objectBase.fsys,
			ver,
			func(path string, d fs.DirEntry, err error) error {
				path = filepath.ToSlash(path)
				if d.IsDir() {
					if !strings.HasPrefix(path, versionContent) && path != ver && !strings.HasPrefix(ver+"/"+objectBase.i.GetContentDir(), path) {
						objectBase.AddValidationWarning(validation.W002, "extra dir '%s' in version '%s'", path, ver)
					}
				} else {
					objectFilesFlat = append(objectFilesFlat, path)
					if strings.HasPrefix(path, versionContent) {
						objectContentFiles[ver] = append(objectContentFiles[ver], path)
						objectContentFilesFlat = append(objectContentFilesFlat, path)
					} else {
						/*
							if !strings.HasPrefix(path, inventoryFile) {
								objectBase.AddValidationWarning(W002, "extra file '%s' in version '%s'", path, ver)
							}
						*/
					}
				}
				return nil
			},
		)
		if len(objectContentFiles[ver]) == 0 {
			fi, err := fs.Stat(objectBase.fsys, versionContent)
			if err != nil {
				if !errors.Is(errors.Cause(err), fs.ErrNotExist) {
					return errors.Wrapf(err, "cannot stat '%s'", versionContent)
				}
			} else {
				if fi.IsDir() {
					objectBase.AddValidationWarning(validation.W003, "empty content folder '%s'", versionContent)
				}
			}
		}
	}
	// load all inventories
	versionInventories, err := objectBase.getVersionInventories()
	if err != nil {
		return errors.Wrap(err, "cannot get version inventories")
	}

	csDigestFiles, err := objectBase.createContentManifest()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := objectBase.i.CheckFiles(csDigestFiles); err != nil {
		return errors.Wrap(err, "cannot check file digests for object root")
	}

	contentDir := ""
	if len(versionStrings) > 0 {
		contentDir = versionInventories[versionStrings[0].String()].GetRealContentDir()
	}
	for _, ver := range versionStrings {
		inv := versionInventories[ver.String()]
		if inv == nil {
			continue
		}
		if contentDir != inv.GetRealContentDir() {
			objectBase.AddValidationError(validation.E019, "content directory '%s' of version '%s' not the same as '%s' in version '%s'", inv.GetRealContentDir(), ver, contentDir, versionStrings[0])
		}
		if err := inv.CheckFiles(csDigestFiles); err != nil {
			return errors.Wrapf(err, "cannot check file digests for version '%s'", ver)
		}
		digestAlg := inv.GetDigestAlgorithm()
		allowedFiles := []string{"inventory.json", "inventory.json." + string(digestAlg)}
		allowedDirs := []string{inv.GetContentDir()}
		versionEntries, err := fs.ReadDir(objectBase.fsys, ver.String())
		if err != nil {
			objectBase.AddValidationError(validation.E010, "cannot read version folder '%s'", ver)
			continue
			//			return errors.Wrapf(err, "cannot read dir '%s'", ver)
		}
		for _, entry := range versionEntries {
			if entry.IsDir() {
				if !slices.Contains(allowedDirs, entry.Name()) {
					objectBase.AddValidationWarning(validation.W002, "extra dir '%s' in version directory '%s'", entry.Name(), ver)
				}
			} else {
				if !slices.Contains(allowedFiles, entry.Name()) {
					objectBase.AddValidationError(validation.E015, "extra file '%s' in version directory '%s'", entry.Name(), ver)
				}
			}
		}
	}

	for key := 0; key < len(versionStrings)-1; key++ {
		v1 := versionStrings[key]
		vi1, ok := versionInventories[v1.String()]
		if !ok {
			objectBase.AddValidationWarning(validation.W010, "no inventory for version '%s'", versionStrings[key])
			continue
			// return errors.Errorf("no inventory for version '%s'", versionStrings[key])
		}
		v2 := versionStrings[key+1]
		vi2, ok := versionInventories[v2.String()]
		if !ok {
			objectBase.AddValidationWarning(validation.W000, "no inventory for version '%s'", versionStrings[key+1])
			continue
		}
		if !inventory.SpecIsLessOrEqual(vi1.GetSpec(), vi2.GetSpec()) {
			objectBase.AddValidationError(validation.E103, "spec in version '%s' (%s) greater than spec in version '%s' (%s)", v1, vi1.GetSpec(), v2, vi2.GetSpec())
		}
	}

	if len(versionStrings) > 0 {
		lastVersion := versionStrings[len(versionStrings)-1]
		if lastInv, ok := versionInventories[lastVersion.String()]; ok {
			if !lastInv.Equals(objectBase.i) {
				objectBase.AddValidationError(validation.E064, "root inventory not equal to inventory version '%s'", lastVersion)
			}
		}
	}

	id := objectBase.i.GetID()
	digestAlg := objectBase.i.GetDigestAlgorithm()
	versions := objectBase.i.GetVersions()
	for ver, verInventory := range versionInventories {
		// check for id consistency
		if id != verInventory.GetID() {
			objectBase.AddValidationError(validation.E037, "invalid id - root inventory id '%s' != version '%s' inventory id '%s'", id, ver, verInventory.GetID())
		}
		if verInventory.GetHead().IsValid() && verInventory.GetHead().String() != ver {
			objectBase.AddValidationError(validation.E040, "wrong head '%s' in manifest for version '%s'", verInventory.GetHead(), ver)
		}

		if verInventory.GetDigestAlgorithm() != digestAlg {
			objectBase.AddValidationError(validation.W000, "different digest algorithm '%s' in version '%s'", verInventory.GetDigestAlgorithm(), ver)
		}

		for versionNumber, vVersion := range verInventory.GetVersions().Iterate() {
			testV := versions.GetVersion(versionNumber)
			if testV == nil {
				objectBase.AddValidationError(validation.E066, "version '%s' in version folder '%s' not in object root manifest", vVersion, versionNumber)
			}
			if !testV.Equals(vVersion) {
				objectBase.AddValidationError(validation.E066, "version '%s' in version folder '%s' not equal to version in object root manifest", vVersion, versionNumber)
			}
		}
	}

	//
	// all files in any manifest must belong to a physical file #E092
	//
	for inventoryVersion, inventory := range versionInventories {
		for manifestFile := range inventory.GetManifest().GetFilesFlat() {
			if !slices.Contains(objectFilesFlat, manifestFile) {
				objectBase.AddValidationError(validation.E092, "file '%s' from manifest not in object content (%s/inventory.json)", manifestFile, inventoryVersion)
			}
		}
	}

	for manifestFile := range objectBase.i.GetManifest().GetFilesFlat() {
		if !slices.Contains(objectFilesFlat, manifestFile) {
			objectBase.AddValidationError(validation.E092, "file '%s' manifest not in object content (./inventory.json)", manifestFile)
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
				versionManifestFiles := ocfl.SeqToSlice(versionInventory.GetManifest().GetFilesFlat())
				for _, objectContentVersionFile := range objectContentVersionFiles {
					// check all inventories which are less in version
					if !slices.Contains(versionManifestFiles, objectContentVersionFile) {
						objectBase.AddValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, inventoryVersion)
					}
				}
			}
		}
		rootVersion := objectBase.i.GetHead()
		if objectContentVersionNumber.Less(rootVersion) {
			rootManifestFiles := ocfl.SeqToSlice(objectBase.i.GetManifest().GetFilesFlat())
			for _, objectContentVersionFile := range objectContentVersionFiles {
				// check all inventories which are less in version
				if !slices.Contains(rootManifestFiles, objectContentVersionFile) {
					objectBase.AddValidationError(validation.E023, "file '%s' not in manifest version '%s'", objectContentVersionFile, rootVersion)
				}
			}
		}
	}

	return nil
}

func (objectBase *ObjectBase) Check() error {
	// https://ocfl.io/1.0/spec/#object-structure
	//object.fs
	objectBase.logger.Info().Msgf("object '%s' with object version '%s' found", objectBase.i.GetID(), objectBase.GetOCFLVersion())
	// check folders

	// check for allowed files and directories
	allowedDirs := []string{"logs", "extensions"}
	for v := range objectBase.i.GetVersions().GetVersionNumbers() {
		allowedDirs = append(allowedDirs, v.String())
	}
	versionCounter := 0
	entries, err := fs.ReadDir(objectBase.fsys, ".")
	if err != nil {
		return errors.Wrap(err, "cannot read object folder")
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !slices.Contains(allowedDirs, entry.Name()) {
				objectBase.AddValidationError(validation.E001, "invalid directory '%s' found", entry.Name())
				// could it be a version folder?
				if _, err := strconv.Atoi(strings.TrimLeft(entry.Name(), "v0")); err == nil {
					if err2 := objectBase.checkVersionFolder(entry.Name()); err2 == nil {
						objectBase.AddValidationError(validation.E046, "root manifest not most recent because of '%s'", entry.Name())
					} else {
						fmt.Println(err2)
					}
				}
			}

			// check version directories
			for v := range objectBase.i.GetVersions().GetVersionNumbers() {
				if v.String() == entry.Name() {
					if err := objectBase.checkVersionFolder(entry.Name()); err != nil {
						return errors.WithStack(err)
					}
					versionCounter++
					break
				}
			}
		} else {
			if !allowedFilesRegexp.MatchString(entry.Name()) {
				objectBase.AddValidationError(validation.E001, "invalid file '%s' found", entry.Name())
			}
		}
	}

	invVersionCounter := len(ocfl.SeqToSlice(objectBase.i.GetVersions().GetVersionNumbers()))
	if versionCounter != invVersionCounter {
		objectBase.AddValidationError(validation.E010, "number of version in inventory (%v) does not fit version in filesystem (%v)", versionCounter, invVersionCounter)
	}

	if err := objectBase.checkFilesAndVersions(); err != nil {
		return errors.WithStack(err)
	}

	dAlgs := []checksum.DigestAlgorithm{objectBase.i.GetDigestAlgorithm()}
	dAlgs = append(dAlgs, ocfl.SeqToSlice(objectBase.i.GetFixity().GetDigestAlgorithms())...)
	return nil
}

// create checksums of all content files
func (objectBase *ObjectBase) createContentManifest() (map[checksum.DigestAlgorithm]map[string][]string, error) {
	// get all possible digest algs
	digestAlgorithms := append(ocfl.SeqToSlice(objectBase.i.GetFixity().GetDigestAlgorithms()), objectBase.i.GetDigestAlgorithm())

	result := map[checksum.DigestAlgorithm]map[string][]string{}
	for versionNumber := range objectBase.i.GetVersions().GetVersionNumbers() {
		if err := fs.WalkDir(
			objectBase.fsys,
			//fmt.Sprintf("%s/%s", version, objectBase.i.GetContentDir()),
			versionNumber.String(),
			func(path string, d fs.DirEntry, err error) error {
				//objectBase.logger.Debug(path)
				if d.IsDir() {
					return nil
				}
				fname := path // filepath.ToSlash(filepath.Join(version, path))
				fp, err := objectBase.fsys.Open(fname)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s'", objectBase.fsys, fname)
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
				return nil
			}); err != nil {
			return nil, errors.Wrapf(err, "cannot walk content dir '%s'", objectBase.i.GetContentDir())
		}
	}
	return result, nil
}

// helper functions

func (objectBase *ObjectBase) getVersionInventories() (map[string]inventory.Inventory, error) {
	if len(objectBase.versionInventories) > 0 {
		return objectBase.versionInventories, nil
	}

	versionStrings := ocfl.SeqToSlice(objectBase.i.GetVersions().GetVersionNumbers())

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
	for _, ver := range versionStrings {
		vi, err := objectBase.joadInventory(ver.String())
		if err != nil {
			if errors.Is(errors.Cause(err), fs.ErrNotExist) {
				objectBase.AddValidationWarning(validation.W010, "no inventory for version '%s'", ver)
				continue
			}
			return nil, errors.Wrapf(err, "cannot load inventory from folder '%s'", ver)
		}
		versionInventories[ver.String()] = vi
	}
	objectBase.versionInventories = versionInventories
	return objectBase.versionInventories, nil
}

/*
func (objectBase *ObjectBase) getAllDigests() ([]checksum.DigestAlgorithm, error) {
	versionInventories, err := objectBase.getVersionInventories()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get version inventories")
	}
	allDigestAlgs := []checksum.DigestAlgorithm{objectBase.i.GetDigestAlgorithm()}
	for _, vi := range versionInventories {
		allDigestAlgs = append(allDigestAlgs, vi.GetDigestAlgorithm())
		for digestAlg := range vi.GetFixity().GetDigestAlgorithms() {
			allDigestAlgs = append(allDigestAlgs, digestAlg)
		}
	}
	slices.Sort(allDigestAlgs)
	allDigestAlgs = slices.Compact(allDigestAlgs)
	return allDigestAlgs, nil
}
*/

func (objectBase *ObjectBase) Extract(fsys fs.FS, version *inventory.VersionNumber, withManifest bool, area string) error {
	var manifest strings.Builder
	var err error
	var digestAlg = objectBase.i.GetDigestAlgorithm()
	if err := objectBase.i.IterateFiles(version, func(internals, externals []string, digest string) error {
		for _, external := range externals {
			external, err = objectBase.extensionManager.BuildObjectExtractPath(objectBase, external, area)
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
				src, err := objectBase.fsys.Open(internal)
				if err != nil {
					return errors.Wrapf(err, "cannot open '%v/%s'", objectBase.fsys, internal)
				}
				defer src.Close()
				target, err := writefs.Create(fsys, external)
				if err != nil {
					return errors.Wrapf(err, "cannot create '%v/%s'", fsys, external)
				}
				defer target.Close()
				objectBase.logger.Debug().Msgf("writing '%v/%s' -> '%v/%s'", objectBase.fsys, internal, fsys, external)
				copyDigests, err := checksum.Copy([]checksum.DigestAlgorithm{digestAlg}, src, target)
				if err != nil {
					return errors.Wrapf(err, "error copying '%v/%s' -> '%v/%s'", objectBase.fsys, internal, fsys, external)
				}
				copyDigest, ok := copyDigests[digestAlg]
				if !ok {
					return errors.Errorf("no digest '%s' generatied", digestAlg)
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
		fp, err := writefs.Create(fsys, manifestName)
		if err != nil {
			return errors.Wrapf(err, "cannot crate manifest file %v/%s", fsys, manifestName)
		}
		if _, err := io.WriteString(fp, manifest.String()); err != nil {
			return errors.Wrapf(err, "cannot write manifest file %v/%s", fsys, manifestName)
		}
		defer fp.Close()
	}
	objectBase.logger.Debug().Msgf("object '%s' extracted", objectBase.i.GetID())
	return nil
}

func (objectBase *ObjectBase) GetAreaPath(area string) (string, error) {
	path, err := objectBase.extensionManager.GetAreaPath(objectBase, area)
	return path, errors.WithStack(err)
}

var _ object.Object = (*ObjectBase)(nil)
