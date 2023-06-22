package extension

import (
	"bufio"
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	"github.com/je4/indexer/v2/pkg/indexer"
	"github.com/op/go-logging"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const ThumbnailName = "NNNN-thumbnail"
const ThumbnailDescription = "preservation management - file thumbnail"

func NewThumbnailFS(fsys fs.FS, thumbnail *thumbnail.Thumbnail, logger *logging.Logger) (*Thumbnail, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &ThumbnailConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	if config.Ext == "" {
		config.Ext = "png"
	} else {
		config.Ext = strings.ToLower(config.Ext)
	}
	if config.Width == 0 {
		config.Width = 256
	}
	if config.Height == 0 {
		config.Height = 256
	}
	ext, err := NewThumbnail(config, thumbnail, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new indexer")
	}
	return ext, nil
}
func NewThumbnail(config *ThumbnailConfig, mig *thumbnail.Thumbnail, logger *logging.Logger) (*Thumbnail, error) {
	sl := &Thumbnail{
		ThumbnailConfig: config,
		logger:          logger,
		thumbnail:       mig,
		buffer:          map[string]*bytes.Buffer{},
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	if mig != nil {
		sl.sourceFS = mig.SourceFS
	}
	return sl, nil
}

type ThumbnailConfig struct {
	*ocfl.ExtensionConfig
	Compress string
	Ext      string
	Width    uint64
	Height   uint64
}

type ThumbnailTarget struct {
	Name           string
	Strategy       string        // add or replace
	FilenameRegexp regexp.Regexp // target expression
	Command        string        // command to execute (stdin --> stdout)
}

type ThumbnailResult struct {
	Ext   string `json:"ext"`
	Error string `json:"error,omitempty"`
	ID    string `json:"id"`
}

type thumbnailLine struct {
	ThumbnailResult
	Checksum string `json:"checksum"`
}

// map pronom to thumbnail
type ThumbnailMap map[string]*ThumbnailTarget

// map checksum to thumbnail
type ThumbnailFiles map[string]*ThumbnailTarget

type Thumbnail struct {
	*ThumbnailConfig
	logger    *logging.Logger
	fsys      fs.FS
	lastHead  string
	thumbnail *thumbnail.Thumbnail
	//buffer         *bytes.Buffer
	buffer      map[string]*bytes.Buffer
	writer      *brotli.Writer
	sourceFS    fs.FS
	currentHead string
	done        bool
}

func (thumb *Thumbnail) GetFS() fs.FS {
	return thumb.fsys
}

func (thumb *Thumbnail) GetConfig() any {
	return thumb.ThumbnailConfig
}

func (thumb *Thumbnail) IsRegistered() bool { return false }

func (thumb *Thumbnail) GetName() string { return ThumbnailName }

func (thumb *Thumbnail) SetFS(fsys fs.FS) { thumb.fsys = fsys }

func (thumb *Thumbnail) SetParams(map[string]string) error {
	return nil
}

func (thumb *Thumbnail) WriteConfig() error {
	if thumb.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(thumb.fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(thumb.ThumbnailConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (thumb *Thumbnail) DoThumbnail(object ocfl.Object, mig *thumbnail.Function, ext string, targetName string, file io.ReadCloser) error {
	tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+ext)
	if err != nil {
		return errors.Wrap(err, "cannot create temp file")
	}
	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		return errors.Wrap(err, "cannot copy file")
	}
	if err := file.Close(); err != nil {
		return errors.Wrap(err, "cannot close file")
	}
	tmpFilename := filepath.ToSlash(tmpFile.Name())
	targetTempName := filepath.ToSlash(filepath.Join(filepath.Dir(tmpFilename), "target."+filepath.Base(tmpFilename)+filepath.Ext(targetName)))

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "cannot close temp file")
	}
	if err := mig.Thumbnail(tmpFilename, targetTempName, thumb.ThumbnailConfig.Width, thumb.ThumbnailConfig.Height, thumb.logger); err != nil {
		//_ = os.Remove(tmpFilename)
		return errors.Wrapf(err, "cannot create thumbnail file '%v' to object '%s'", targetName, object.GetID())
	}
	/*
		if err := os.Remove(tmpFilename); err != nil {
			return errors.Wrapf(err, "cannot remove temp file '%s'", tmpFilename)
		}
	*/

	mFile, err := os.Open(targetTempName)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", targetTempName)
	}
	tf, err := writefs.Create(thumb.fsys, targetName)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s/%s", thumb.fsys, targetName)
	}
	if _, err := io.Copy(tf, mFile); err != nil {
		return errors.Wrapf(err, "cannot write %s to %s/%s", targetTempName, thumb.fsys, targetName)
	}
	if err := tf.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file %s/%s", thumb.fsys, targetName)
	}
	if err := mFile.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file '%s'", targetTempName)
	}
	/*
		if err := os.Remove(targetTempName); err != nil {
			return errors.Wrapf(err, "cannot remove temp file '%s'", targetTempName)
		}
	*/

	return nil
}

func (thumb *Thumbnail) UpdateObjectBefore(ocfl.Object) error {
	return nil
}

func (thumb *Thumbnail) UpdateObjectAfter(object ocfl.Object) error {
	inventory := object.GetInventory()
	head := inventory.GetHead()
	thumb.buffer[head] = &bytes.Buffer{}
	thumb.writer = brotli.NewWriter(thumb.buffer[head])
	if inventory == nil {
		return errors.Errorf("inventory is nil")
	}

	// first get the metadata from the object
	meta, err := object.GetMetadata()
	if err != nil {
		return errors.Wrapf(err, "cannot get metadata from object %s", object.GetID())
	}
	for cs, m := range meta.Files {
		_, ok := m.Extension[ThumbnailName]
		if ok {
			continue
		}
		indexerMetaAny, ok := m.Extension[IndexerName]
		if !ok {
			continue
		}
		indexerMeta, ok := indexerMetaAny.(*indexer.ResultV2)
		if !ok {
			continue
		}
		thumbnailFunction, err := thumb.thumbnail.GetFunctionByPronom(indexerMeta.Pronom)
		if err != nil {
			thumbnailFunction, err = thumb.thumbnail.GetFunctionByMimetype(indexerMeta.Mimetype)
			if err != nil {
				continue
			}
		}

		targetFile := fmt.Sprintf("data/%s/%s/%s.%s", string([]rune(cs)[0]), string([]rune(cs)[1]), cs, strings.ToLower(thumb.ThumbnailConfig.Ext))

		var file io.ReadCloser
		var ext string
		fsys := object.GetFS()
		if fsys != nil {
			file, err = fsys.Open(m.InternalName[0])
			if err != nil {
				file = nil
			}
			ext = filepath.Ext(m.InternalName[0])
		}
		if file == nil {
			if thumb.sourceFS != nil {
				thumb.logger.Infof("create thumbnail for %s", m.InternalName[0])
				stateFiles, err := inventory.GetStateFiles("", cs)
				if err != nil {
					return errors.Wrapf(err, "cannot get state files for checksum '%s' in object '%s'", cs, object.GetID())
				}
				if len(stateFiles) == 0 {
					return errors.Errorf("zero state file for checksum '%s' in object '%s'", cs, object.GetID())
				}
				external, err := object.GetExtensionManager().BuildObjectExtractPath(object, stateFiles[len(stateFiles)-1], "")
				if err != nil {
					return errors.Wrapf(err, "cannot build external path for file '%s' in object '%s'", stateFiles[len(stateFiles)-1], object.GetID())
				}
				file, err = thumb.sourceFS.Open(external)
				if err != nil {
					continue
					// return errors.Wrapf(err, "cannot open file '%v/%s' in source filesystem", thumb.sourceFS, external)
				}
				ext = filepath.Ext(external)
				var ml *thumbnailLine
				var errStr string
				if err := thumb.DoThumbnail(object, thumbnailFunction, ext, targetFile, file); err != nil {
					errStr = err.Error()
				}
				ml = &thumbnailLine{
					Checksum: cs,
					ThumbnailResult: ThumbnailResult{
						Ext:   thumb.ThumbnailConfig.Ext,
						Error: errStr,
						ID:    thumbnailFunction.GetID(),
					},
				}

				data, err := json.Marshal(ml)
				if err != nil {
					return errors.Wrapf(err, "cannot marshal thumbnail line for file '%s' in object '%s'", targetFile, object.GetID())
				}
				if _, err := thumb.writer.Write(append(data, []byte("\n")...)); err != nil {
					return errors.Wrapf(err, "cannot write thumbnail line for file '%s' in object '%s'", targetFile, object.GetID())
				}
			}
		}
	}
	thumb.writer.Flush()
	thumb.writer.Close()
	thumb.writer = nil
	buffer, ok := thumb.buffer[head]
	if !ok {
		return nil
	}
	if err := ocfl.WriteJsonL(
		object,
		"thumbnail",
		buffer.Bytes(),
		thumb.ThumbnailConfig.Compress,
		"extension",
		"",
		thumb.fsys,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}

	return nil
}

func (thumb *Thumbnail) GetMetadata(object ocfl.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}

	inventory := object.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v := range inventory.GetVersions() {
		var data []byte
		if buf, ok := thumb.buffer[v]; ok && buf.Len() > 0 {
			//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
			// need a new reader on the buffer
			reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
			data, err = io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", object.GetID(), v)
			}
		} else {
			data, err = ocfl.ReadJsonL(object, "thumbnail", v, thumb.ThumbnailConfig.Compress, "extension", "", thumb.fsys)
			if err != nil {
				continue
				// return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", object.GetID(), v)
			}
		}

		reader := bytes.NewReader(data)
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 16*1024*1024)
		r.Split(bufio.ScanLines)
		for r.Scan() {
			line := r.Text()
			var meta = thumbnailLine{}
			if err := json.Unmarshal([]byte(line), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", object.GetID(), v, line)
			}
			var digest string
			for cs, _ := range manifest {
				if cs == meta.Checksum {
					digest = cs
					break
				}
			}
			if digest == "" {
				return nil, errors.Errorf("cannot find checksum for file '%s' in object '%s'", meta.Checksum, object.GetID())
			}
			result[digest] = &meta.ThumbnailResult
		}
		if err := r.Err(); err != nil {
			return nil, errors.Wrapf(err, "cannot scan lines for '%s' %s", object.GetID(), v)
		}
	}
	return result, nil
}

var (
	_ ocfl.Extension             = &Thumbnail{}
	_ ocfl.ExtensionObjectChange = &Thumbnail{}
	_ ocfl.ExtensionMetadata     = &Thumbnail{}
)
