package extension

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	inventorytypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"golang.org/x/exp/slices"
)

const FilesystemName = "NNNN-filesystem"
const FilesystemDescription = "preserves filesytem metadata"

func init() {
	extension.RegisterExtension(FilesystemName, NewFilesystem, nil)
}

func NewFilesystem() (extensiontypes.Extension, error) {
	config := &FilesystemConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: FilesystemName},
		Folders:         "",
		StorageType:     "",
		StorageName:     "",
		Compress:        "none",
	}
	sl := &Filesystem{
		FilesystemConfig: config,
		buffer:           map[string]*bytes.Buffer{},
	}
	// sl.writer = brotli.NewWriter(sl.buffer)
	return sl, nil
}

type FilesystemMeta struct {
	ATime        time.Time `json:"aTime"`
	MTime        time.Time `json:"mTime"`
	CTime        time.Time `json:"cTime"`
	Attr         string    `json:"attr,omitempty"`
	Symlink      string    `json:"symlink,omitempty"`
	OS           string    `json:"os"`
	SystemStat   any       `json:"sysStat,omitempty"`
	StateVersion string    `json:"stateVersion"`
	Size         uint64    `json:"size"`
}

type FileSystemLine struct {
	Path string          `json:"path"`
	Meta *FilesystemMeta `json:"meta"`
}

type FilesystemConfig struct {
	*extensiontypes.ExtensionConfig
	Folders     string `json:"folders"`
	StorageType string `json:"storageType"`
	StorageName string `json:"storageName"`
	Compress    string `json:"compress"`
}

type Filesystem struct {
	*FilesystemConfig
	lastHead    string
	currentHead string
	buffer      map[string]*bytes.Buffer
	writer      *brotli.Writer
	logger      ocfllogger.OCFLLogger
}

func (fi *Filesystem) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	fi.logger = logger.With("extension", FilesystemName)
	return fi
}

func (fi *Filesystem) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, fi.FilesystemConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal FilesystemConfig '%s'", string(data))
	}
	if fi.Compress == "" {
		fi.Compress = "none"
	}
	return nil
}

func (fi *Filesystem) Terminate() error {
	return nil
}

func (fi *Filesystem) GetConfig() any {
	return fi.FilesystemConfig
}

func (fi *Filesystem) AddFileBefore(object object.VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	return nil
}

func (fi *Filesystem) UpdateFileBefore(object object.VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	return nil
}

func (fi *Filesystem) DeleteFileBefore(versionWriter object.VersionWriter, dest string, area string) error {
	return nil
}

func (fi *Filesystem) AddFileAfter(versionWriter object.VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	if isDir && fi.Folders == "" {
		return nil
	}

	inventory := versionWriter.GetInventory()
	latestVersion := inventory.GetVersions().GetVersion(inventory.GetVersions().LatestVersionNumber())

	var err error
	var emptyChecksum string
	var emptyExists bool
	if isDir && fi.Folders != "" {
		emptyChecksum, err = checksum.Checksum(bytes.NewReader([]byte{}), inventory.GetDigestAlgorithm())
		if err != nil {
			return errors.Wrap(err, "cannot calculate checksum for empty file")
		}
		manifest := inventory.GetManifest()
		emptyFiles, err := manifest.GetFiles(emptyChecksum)
		if err != nil {
			if !errors.Is(err, inventorytypes.DigestNotFound) {
				return errors.Wrap(err, "cannot get files empty manifest")
			}
		}
		emptyExists = len(emptyFiles) != 0
	}

	head := inventory.GetHead().String()
	if _, ok := fi.buffer[head]; !ok {
		fi.buffer[head] = &bytes.Buffer{}
	}
	if fi.currentHead != head {
		fi.writer = brotli.NewWriter(fi.buffer[head])
		fi.currentHead = head
	}

	for _, src := range source {
		fullpath, err := writefs.Fullpath(sourceFS, src)
		if err != nil {
			if errors.Cause(err) == writefs.ErrNotImplemented {
				continue
			}
			return errors.Wrapf(err, "cannot get fullpath for '%v/%s'", sourceFS, src)
		}
		fsMeta := &FilesystemMeta{StateVersion: inventory.GetHead().String()}
		// we work only on local filesystems with this extension
		stat, err := os.Stat(fullpath)
		if err != nil {
			return errors.Wrapf(err, "cannot stat file '%s'", src)
		}
		if err := fsMeta.init(fullpath, stat); err != nil {
			return errors.Wrapf(err, "cannot init fsMeta for '%s'", fullpath)
		}
		newSrc, err := versionWriter.GetExtensionManager().BuildObjectStatePath(src, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build object extract path for '%s'", src)
		}

		if isDir {
			newEmptyFile := filepath.ToSlash(filepath.Join(src, fi.Folders))
			if !emptyExists {
				if _, err := versionWriter.AddReader(io.NopCloser(bytes.NewReader([]byte{})), []string{newEmptyFile}, area, true, false); err != nil {
					return errors.Wrapf(err, "cannot add empty file '%s'", newEmptyFile)
				}
				emptyExists = true
			} else {
				// todo: make it more elegant
				names, err := versionWriter.BuildNames([]string{newEmptyFile}, area)
				if err != nil {
					return errors.Wrapf(err, "cannot build names for '%s'", newEmptyFile)
				}
				if len(names.ExternalPaths) == 0 {
					return errors.Errorf("cannot build external names for '%s'", newEmptyFile)
				}
				newEmptyFile = names.ExternalPaths[0]
				if _, err := latestVersion.CopyFile(names.ExternalPaths[0], emptyChecksum); err != nil {
					return errors.Wrapf(err, "cannot copy empty file to '%s'", newEmptyFile)
				}
			}
			newSrc = newEmptyFile
		}

		fsLine := &FileSystemLine{
			Path: newSrc,
			Meta: fsMeta,
		}
		data, err := json.Marshal(fsLine)
		if err != nil {
			return errors.Errorf("cannot marshal result %v", fsLine)
		}
		if _, err := fi.writer.Write(append(data, []byte("\n")...)); err != nil {
			return errors.Errorf("cannot brotli %s", string(data))
		}
	}
	return nil
}

func (fi *Filesystem) UpdateFileAfter(versionWriter object.VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error {
	return errors.WithStack(
		fi.AddFileAfter(versionWriter, sourceFS, []string{source}, "", "", area, isDir),
	)

}

func (fi *Filesystem) DeleteFileAfter(object object.VersionWriter, dest string, area string) error {
	return nil
}

func (fi *Filesystem) NeedNewVersion(object object.VersionWriter) (bool, error) {
	return false, nil
}

func (fi *Filesystem) DoNewVersion(object object.VersionWriter) error {
	return nil
}

func (fi *Filesystem) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	var err error
	var result = map[string]map[string][]*FileSystemLine{}

	inventory := obj.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest.Iterate() {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v := range inventory.GetVersions().GetVersionNumbers() {
		var data []byte
		if buf, ok := fi.buffer[v.String()]; ok && buf.Len() > 0 {
			//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
			// need a new reader on the buffer
			reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
			data, err = io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", obj.GetID(), v)
			}
		} else {
			data, err = ReadJsonL(fi.GetName(), sourceFS, obj, v, "filesystem", fi.FilesystemConfig.Compress, fi.StorageType, fi.StorageName)
			if err != nil {
				continue
				// return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", object.GetID(), v)
			}
		}

		reader := bytes.NewReader(data)
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 16*1024*1024)
		r.Split(bufio.ScanLines)
		lines := []*FileSystemLine{}
		for r.Scan() {
			lineStr := r.Text()
			var meta = &FileSystemLine{}
			if err := json.Unmarshal([]byte(lineStr), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", obj.GetID(), v, lineStr)
			}
			lines = append(lines, meta)
		}
		if err := inventory.IterateFiles(v, func(internals, externals []string, digest string) error {
			for _, line := range lines {
				if slices.Contains(externals, line.Path) {
					if _, ok := result[digest]; !ok {
						result[digest] = map[string][]*FileSystemLine{}
					}
					if _, ok := result[digest][line.Meta.StateVersion]; !ok {
						result[digest][line.Meta.StateVersion] = []*FileSystemLine{}
					}
					result[digest][line.Meta.StateVersion] = append(result[digest][line.Meta.StateVersion], line)
				}
			}
			return nil
		}); err != nil {
			return nil, errors.Wrapf(err, "cannot iterate state files for '%s' version '%s'", obj.GetID(), v)
		}
	}
	var retResult = map[string]any{}
	for digest, lines := range result {
		retResult[digest] = lines
	}
	return retResult, nil
}

func (fi *Filesystem) UpdateObjectBefore(object object.VersionWriter) error {
	return nil
}

func (fi *Filesystem) UpdateObjectAfter(object object.VersionWriter) error {
	if fi.writer == nil {
		return nil
	}
	if err := fi.writer.Flush(); err != nil {
		return errors.Wrap(err, "cannot flush brotli writer")
	}
	if err := fi.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close brotli writer")
	}
	head := object.GetInventory().GetHead()
	if !head.IsValid() {
		return errors.Errorf("no head for object '%s'", object.GetID())
	}
	buffer, ok := fi.buffer[head.String()]
	if !ok {
		return nil
	}
	if err := WriteJsonL(
		fi.GetName(),
		object,
		"filesystem",
		buffer.Bytes(),
		fi.FilesystemConfig.Compress,
		fi.StorageType,
		fi.StorageName,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}
	return nil
}

func (fi *Filesystem) SetParams(params map[string]string) error {
	return nil
}

func (fi *Filesystem) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(fi.FilesystemConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil

}

func (fi *Filesystem) IsRegistered() bool {
	return false
}

func (fi *Filesystem) GetName() string {
	return FilesystemName
}

var (
	_ extensiontypes.Extension      = &Filesystem{}
	_ object.ExtensionObjectChange  = &Filesystem{}
	_ object.ExtensionContentChange = &Filesystem{}
	_ object.ExtensionMetadata      = &Filesystem{}
	_ object.ExtensionNewVersion    = &Filesystem{}
)
