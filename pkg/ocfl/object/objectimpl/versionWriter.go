package objectimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

func NewVersionWriter(obj object.Object, objectFS streamfs.FS, echo bool, logger ocfllogger.OCFLLogger) (object.VersionWriter, error) {
	vw := &versionWriter{
		Object:      obj,
		objectFS:    objectFS,
		logger:      logger,
		echo:        echo,
		updateFiles: []string{},
	}
	if err := vw.init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize version writer")
	}
	return vw, nil
}

type versionWriter struct {
	object.Object
	objectFS    streamfs.FS
	versionFS   streamfs.FS
	ver         *inventory.VersionNumber
	logger      ocfllogger.OCFLLogger
	updateFiles []string
	echo        bool
	area        string
}

func (versionWriter *versionWriter) BeginArea(area string) {
	versionWriter.area = area
	versionWriter.updateFiles = []string{}
}

func (versionWriter *versionWriter) EndArea() error {
	if versionWriter.echo {
		if err := versionWriter.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot remove files")
		}
	}
	versionWriter.updateFiles = []string{}
	versionWriter.area = ""
	return nil
}

func (versionWriter *versionWriter) init() error {
	var err error
	versionWriter.ver = versionWriter.GetInventory().GetHead()
	versionWriter.versionFS, err = streamfs.Sub(versionWriter.versionFS, versionWriter.ver.String())
	if err != nil {
		return errors.Wrapf(err, "failed to open version stream %v/%s", versionWriter.versionFS, versionWriter.ver.String())
	}
	if err := versionWriter.GetExtensionManager().UpdateObjectBefore(versionWriter); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectBefore()")
	}
	if versionWriter.ver.Int() <= 1 {
		if err := versionWriter.storeExtensions(); err != nil {
			return errors.Wrap(err, "failed to store extensions")
		}
	}

	return nil
}

func (versionWriter *versionWriter) storeInventory(version bool, objectRoot bool) error {
	inv := versionWriter.GetInventory()

	// check whether object filesystem is writeable
	if !inv.IsWriteable() {
		return errors.New("inventory not writeable - not updated")
	}

	// create inv.json from inventory
	iFileName := "inv.json"
	jsonBytes, err := json.MarshalIndent(inv, "", "   ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal inventory")
	}
	h, err := checksum.GetHash(inv.GetDigestAlgorithm())
	if err != nil {
		return errors.Wrapf(err, "invalid digest algorithm '%s'", string(inv.GetDigestAlgorithm()))
	}
	if _, err := h.Write(jsonBytes); err != nil {
		return errors.Wrapf(err, "cannot create checksum of manifest")
	}
	checksumBytes := h.Sum(nil)
	checksumString := fmt.Sprintf("%x %s", checksumBytes, iFileName)

	if objectRoot {
		iWriter, err := writefs.Create(versionWriter.objectFS, iFileName)
		if err != nil {
			return errors.Wrap(err, "cannot create inv.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			if err := iWriter.Close(); err != nil {
				versionWriter.logger.Error().Err(err).Msg("cannot close version writer")
			}
			return errors.Wrap(err, "cannot write to inv.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", versionWriter.objectFS, iFileName)
		}
		csFileName := fmt.Sprintf("inv.json.%s", string(inv.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(versionWriter.objectFS, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", versionWriter.objectFS, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			if err := iCSWriter.Close(); err != nil {
				versionWriter.logger.Error().Err(err).Msg("cannot close iCSWriter writer")
			}
			return errors.Wrapf(err, "cannot write to '%v/%s'", versionWriter.objectFS, csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", versionWriter.objectFS, csFileName)
		}
	}
	if version {
		iFileName = fmt.Sprintf("%s/inv.json", inv.GetHead())
		iWriter, err := writefs.Create(versionWriter.versionFS, iFileName)
		if err != nil {
			return errors.Wrap(err, "cannot create inv.json")
		}
		if _, err := iWriter.Write(jsonBytes); err != nil {
			if err := iWriter.Close(); err != nil {
				versionWriter.logger.Error().Err(err).Msg("cannot close version writer")
			}
			return errors.Wrap(err, "cannot write to inv.json")
		}
		if err := iWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", versionWriter.versionFS, iFileName)
		}
		csFileName := fmt.Sprintf("%s/inv.json.%s", inv.GetHead(), string(inv.GetDigestAlgorithm()))
		iCSWriter, err := writefs.Create(versionWriter.versionFS, csFileName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", versionWriter.versionFS, csFileName)
		}
		if _, err := iCSWriter.Write([]byte(checksumString)); err != nil {
			if err := iCSWriter.Close(); err != nil {
				versionWriter.logger.Error().Err(err).Msg("cannot close iCSWriter writer")
			}
			return errors.Wrapf(err, "cannot write to '%s'", csFileName)
		}
		if err := iCSWriter.Close(); err != nil {
			return errors.Wrapf(err, "cannot close '%v/%s'", versionWriter.versionFS, csFileName)
		}
	}
	return nil
}

func (versionWriter *versionWriter) storeExtensions() error {
	versionWriter.logger.Debug()

	subFS, err := streamfs.Sub(versionWriter.versionFS, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create sub filesystem %v/extensions", versionWriter.versionFS)
	}
	if err := versionWriter.GetExtensionManager().WriteConfig(subFS); err != nil {
		return errors.Wrap(err, "cannot store extension configs")
	}
	return nil

}

func (versionWriter *versionWriter) Close() error {
	inv := versionWriter.GetInventory()
	versionWriter.logger.Info().Msgf(fmt.Sprintf("Closing Version %s of object '%s'", versionWriter.ver.String(), inv.GetID()))
	if !(inv.IsWriteable()) {
		return nil
	}

	if !inv.IsModified() {
		return nil
	}
	if versionWriter.echo {
		if err := versionWriter.echoDelete(); err != nil {
			return errors.Wrap(err, "cannot delete files")
		}
	}

	//object.storageRoot.setModified()

	if err := versionWriter.GetExtensionManager().UpdateObjectAfter(versionWriter); err != nil {
		return errors.Wrapf(err, "cannot execute ext.UpdateObjectAfter()")
	}
	if err := inv.Clean(); err != nil {
		return errors.Wrap(err, "cannot clean inventory")
	}
	if err := versionWriter.storeInventory(true, false); err != nil {
		return errors.Wrap(err, "cannot store inventory")
	}
	if needVersion, err := versionWriter.GetExtensionManager().NeedNewVersion(versionWriter); err != nil {
		return errors.Wrapf(err, "cannot execute ext.NeedNewVersion()")
	} else if needVersion {
		nextVersion, err := versionWriter.StartUpdate(nil, "automated version", "gocfl", "https://github.com/ocfl-archive/gocfl", false)
		if err != nil {
			return errors.Wrap(err, "cannot create new version")
		}
		if err := versionWriter.GetExtensionManager().DoNewVersion(nextVersion); err != nil {
			return errors.Wrapf(err, "cannot execute ext.DoNewVersion()")
		}
		if err := nextVersion.Close(); err != nil {
			return errors.Wrap(err, "cannot close next version")
		}
	}
	return nil
}

func (versionWriter *versionWriter) BuildNames(files []string, area string) (*object.NamesStruct, error) {
	var err error
	result := &object.NamesStruct{
		ExternalPaths: []string{},
	}
	for _, file := range files {
		externalPath, err := versionWriter.GetExtensionManager().BuildObjectStatePath(file, area)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", file)
		}
		result.ExternalPaths = append(result.ExternalPaths, externalPath)
	}
	result.InternalPath, err = versionWriter.GetExtensionManager().BuildObjectManifestPath(files[0], area)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create manifest path for '%s'", files[0])
	}
	result.ManifestPath = versionWriter.GetInventory().BuildManifestName(result.InternalPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create virtual filename for '%s'", result.InternalPath)
	}
	return result, nil
}

func (versionWriter *versionWriter) echoDelete() error {
	inv := versionWriter.GetInventory()
	slices.Sort(versionWriter.updateFiles)
	versionWriter.updateFiles = slices.Compact(versionWriter.updateFiles)
	basePath, err := versionWriter.GetExtensionManager().BuildObjectStatePath(".", "")
	if err != nil {
		return errors.Wrap(err, "cannot build external path for '.'")
	}
	if basePath == "." {
		basePath = ""
	}
	version := inv.GetVersions().GetVersion(inv.GetHead())
	if _, err := version.EchoDelete(versionWriter.updateFiles, basePath); err != nil {
		return errors.Wrap(err, "cannot remove deleted files from inventory")
	}
	return nil
}
func (versionWriter *versionWriter) AddFolder(sourceFS fs.FS, checkDuplicate bool, area string) error {
	versionWriter.logger.Debug().Msgf("walking '%v'", sourceFS)
	if err := fs.WalkDir(sourceFS, ".", func(path string, info fs.DirEntry, err error) error {
		if info.Name() == "." {
			return nil
		}
		path = filepath.ToSlash(path)
		if err := versionWriter.AddFile(sourceFS, path, checkDuplicate, area, false, info.IsDir()); err != nil {
			return errors.Wrapf(err, "cannot add file '%s'", path)
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot walk filesystem")
	}

	return nil
}

func (versionWriter *versionWriter) AddFile(sourceFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	versionWriter.logger.Info().Msgf("adding file %s:%s", area, path)

	/*
		ver := inv.GetVersions().GetVersion(inv.GetHead())
		if ver == nil {
			return errors.Errorf("version %s not found", inv.GetHead())
		}
	*/

	path = filepath.ToSlash(path)
	inv := versionWriter.GetInventory()
	if !inv.IsWriteable() {
		return errors.New("object inventory not writeable")
	}

	names, err := versionWriter.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot create virtual filename for '%s'", path)
	}

	targetFilename := inv.BuildManifestName(names.InternalPath)

	var digest string
	if !isDir {

		digestAlgorithms := []checksum.DigestAlgorithm{}
		for alg := range inv.GetFixity().GetDigestAlgorithms() {
			digestAlgorithms = append(digestAlgorithms, alg)
		}

		file, err := sourceFS.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot open file '%v/%s'", sourceFS, path)
		}
		newPath, err := versionWriter.GetExtensionManager().BuildObjectStatePath(path, area)
		if err != nil {
			if err := file.Close(); err != nil {
				versionWriter.logger.Error().Err(err).Msgf("cannot close file '%s'", path)
			}
			return errors.Wrapf(err, "cannot map external path '%s'", path)
		}

		versionWriter.updateFiles = append(versionWriter.updateFiles, newPath)

		if checkDuplicate {
			// do the checksum
			digest, err = checksum.Checksum(file, inv.GetDigestAlgorithm())
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
				if err := file.Close(); err != nil {
					versionWriter.logger.Error().Err(err).Msgf("cannot close file '%s'", path)
				}
				file, err = sourceFS.Open(path)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s'", sourceFS, path)
				}
			}
			// if file is already there we do nothing
			dup, err := inv.AlreadyExists(newPath, digest)
			if err != nil {
				return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
			}
			if dup {
				versionWriter.logger.Info().Msgf("[%s] '%s' already exists. ignoring", inv.GetID(), newPath)
				return nil
			}
			// file already ingested, but new virtual name
			if dups, _ := inv.GetManifest().GetFiles(digest); len(dups) > 0 {
				versionWriter.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", inv.GetID(), newPath)
				ver := inv.GetVersions().GetVersion(inv.GetVersions().LatestVersionNumber())
				if ver == nil {
					return errors.Errorf("version %s not found", inv.GetVersions().LatestVersionNumber())
				}
				if _, err := ver.CopyFile(newPath, digest); err != nil {
					return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
				}
				return nil
			}
		} else {
			if !slices.Contains(digestAlgorithms, inv.GetDigestAlgorithm()) {
				digestAlgorithms = append(digestAlgorithms, inv.GetDigestAlgorithm())
			}
		}
		if !noExtensionHook {
			if err := versionWriter.GetExtensionManager().AddFileBefore(versionWriter, nil, path, names.InternalPath, area, isDir); err != nil {
				return errors.Wrapf(err, "error on AddFileBefore() extension hook")
			}
		}

		digest, err = versionWriter.addReader(file, names, noExtensionHook)
		if err != nil {
			file.Close()
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
		if err := file.Close(); err != nil {
			return errors.Wrapf(err, "cannot close file '%s'", path)
		}

	}
	if !noExtensionHook {
		if err := versionWriter.GetExtensionManager().AddFileAfter(versionWriter, sourceFS, []string{path}, targetFilename, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (versionWriter *versionWriter) addReader(r io.ReadCloser, names *object.NamesStruct, noExtensionHook bool) (string, error) {
	inv := versionWriter.GetInventory()
	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inv.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}

	var digest string

	versionWriter.updateFiles = append(versionWriter.updateFiles, names.ExternalPaths...)

	if !slices.Contains(digestAlgorithms, inv.GetDigestAlgorithm()) {
		digestAlgorithms = append(digestAlgorithms, inv.GetDigestAlgorithm())
	}

	writer, err := writefs.Create(versionWriter.versionFS, names.ManifestPath)
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
			if err := versionWriter.GetExtensionManager().StreamObject(versionWriter, pr, names.ExternalPaths, names.InternalPath); err != nil {
				extErrors <- err
			}
		}()
		checksums, err = checksum.Copy(digestAlgorithms, r, writer, pw)
		if err := pw.Close(); err != nil {
			versionWriter.logger.Error().Err(err).Msg("cannot close pipe writer")
		}
		wg.Wait()
		if err != nil {
			return "", errors.Wrapf(err, "cannot copy '%s' -> '%s'", names.ExternalPaths, names.ManifestPath)
		}
		close(extErrors)
		select {
		case err, ok := <-extErrors:
			if ok {
				return "", errors.Wrapf(err, "error on StreamObject() extension hook for object '%s'", inv.GetID())
			}
		default:
		}
	}

	if digest == "" {
		var ok bool
		digest, ok = checksums[inv.GetDigestAlgorithm()]
		if !ok {
			return "", errors.Errorf("digest '%s' not generated", inv.GetDigestAlgorithm())
		}
	} else {
		checksums[inv.GetDigestAlgorithm()] = digest
	}
	if err := inv.AddFile(names.ExternalPaths, names.ManifestPath, checksums); err != nil {
		return "", errors.Wrapf(err, "cannot append '%v'/'%s' to inventory", names.ExternalPaths, names.InternalPath)
	}

	return digest, nil
}

func (versionWriter *versionWriter) AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error {
	inv := versionWriter.GetInventory()
	if !inv.IsWriteable() {
		return errors.New("object not writeable")
	}

	ver := inv.GetVersions().GetVersion(inv.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", inv.GetHead())
	}

	digestAlgorithms := []checksum.DigestAlgorithm{}
	for alg := range inv.GetFixity().GetDigestAlgorithms() {
		digestAlgorithms = append(digestAlgorithms, alg)
	}
	var digest string

	names, err := versionWriter.BuildNames([]string{path}, area)
	if err != nil {
		return errors.Wrapf(err, "cannot names for '%s'", path)

	}

	versionWriter.logger.Info().Msgf("adding file %s:%s", area, path)

	newPath, err := versionWriter.GetExtensionManager().BuildObjectStatePath(path, area)
	if err != nil {
		return errors.Wrapf(err, "cannot map external path '%s'", path)
	}

	versionWriter.updateFiles = append(versionWriter.updateFiles, newPath)

	var dataReader = bytes.NewReader(data)
	if checkDuplicate {
		// do the checksum
		digest, err = checksum.Checksum(dataReader, inv.GetDigestAlgorithm())
		if err != nil {
			return errors.Wrapf(err, "cannot create digest of '%s'", path)
		}
		// set filepointer to beginning
		if _, err := dataReader.Seek(0, 0); err != nil {
			return errors.Wrapf(err, "cannot seek in datareader")
		}
		// if file is already there we do nothing
		dup, err := inv.AlreadyExists(newPath, digest)
		if err != nil {
			return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", names.InternalPath, digest)
		}
		if dup {
			versionWriter.logger.Info().Msgf("[%s] '%s' already exists. ignoring", inv.GetID(), newPath)
			return nil
		}
		// file already ingested, but new virtual name
		if dups, _ := inv.GetManifest().GetFiles(digest); len(dups) > 0 {
			versionWriter.logger.Info().Msgf("[%s] file with same content as '%s' already exists. creating virtual copy", inv.GetID(), newPath)
			if _, err := ver.CopyFile(newPath, digest); err != nil {
				return errors.Wrapf(err, "cannot append '%s' to inventory as '%s'", path, names.InternalPath)
			}
			return nil
		}
	} else {
		if !slices.Contains(digestAlgorithms, inv.GetDigestAlgorithm()) {
			digestAlgorithms = append(digestAlgorithms, inv.GetDigestAlgorithm())
		}
	}

	if !noExtensionHook {
		if err := versionWriter.GetExtensionManager().AddFileBefore(versionWriter, nil, path, names.InternalPath, area, false); err != nil {
			return errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var r = io.NopCloser(dataReader)
	if !isDir {
		digest, err = versionWriter.addReader(r, names, noExtensionHook)
		if err != nil {
			return errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	}

	if !noExtensionHook {
		if err := versionWriter.GetExtensionManager().AddFileAfter(versionWriter, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return nil
}

func (versionWriter *versionWriter) AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error) {
	inv := versionWriter.GetInventory()
	if len(files) == 0 {
		return "", errors.New("no files given")
	}
	if !inv.IsWriteable() {
		return "", errors.New("object not writeable")
	}
	path := files[0]
	names, err := versionWriter.BuildNames(files, area)

	versionWriter.logger.Info().Msgf("adding file %s:%v", area, files)

	if !noExtensionHook {
		if err := versionWriter.GetExtensionManager().AddFileBefore(versionWriter, nil, path, names.InternalPath, area, false); err != nil {
			return "", errors.Wrapf(err, "error on AddFileBefore() extension hook")
		}
	}

	var digest string
	if !isDir {
		digest, err = versionWriter.addReader(r, names, noExtensionHook)
		if err != nil {
			return "", errors.Wrapf(err, "cannot add file '%s' to object", path)
		}
	} else {
		io.Copy(io.Discard, r)
	}

	if !noExtensionHook {
		if err := versionWriter.GetExtensionManager().AddFileAfter(versionWriter, nil, names.ExternalPaths, names.ManifestPath, digest, area, isDir); err != nil {
			return "", errors.Wrapf(err, "error on AddFileAfter() extension hook")
		}
	}

	return digest, nil
}

func (versionWriter *versionWriter) DeleteFile(virtualFilename string, digest string) error {
	inv := versionWriter.GetInventory()
	ver := inv.GetVersions().GetVersion(inv.GetVersions().LatestVersionNumber())
	if ver == nil {
		return errors.Errorf("version %s not found", inv.GetHead())
	}
	virtualFilename = filepath.ToSlash(virtualFilename)
	versionWriter.logger.Debug().Msgf("removing '%s' [%s]", virtualFilename, digest)

	if !inv.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := inv.AlreadyExists(virtualFilename, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilename, digest)
	}
	if !dup {
		versionWriter.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilename, digest)
		return nil
	}
	if _, err := ver.DeleteFile(virtualFilename); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilename)
	}
	return nil

}

func (versionWriter *versionWriter) RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error {
	inv := versionWriter.GetInventory()
	ver := inv.GetVersions().GetVersion(inv.GetHead())
	if ver == nil {
		return errors.Errorf("version %s not found", inv.GetHead())
	}
	virtualFilenameSource = filepath.ToSlash(virtualFilenameSource)
	versionWriter.logger.Debug().Msgf("removing '%s' [%s]", virtualFilenameSource, digest)

	if !inv.IsWriteable() {
		return errors.New("object not writeable")
	}

	// if file is not there we do nothing
	dup, err := inv.AlreadyExists(virtualFilenameSource, digest)
	if err != nil {
		return errors.Wrapf(err, "cannot check duplicate for '%s' [%s]", virtualFilenameSource, digest)
	}
	if !dup {
		versionWriter.logger.Debug().Msgf("'%s' [%s] not in archive - ignoring", virtualFilenameSource, digest)
		return nil
	}
	if _, err := ver.RenameFile(virtualFilenameSource, virtualFilenameDest); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", virtualFilenameSource)
	}
	return nil

}

var _ object.VersionWriter = (*versionWriter)(nil)
