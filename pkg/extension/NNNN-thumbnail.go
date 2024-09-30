package extension

import (
	"bufio"
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	"golang.org/x/exp/slices"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
	_ "image/gif"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const ThumbnailName = "NNNN-thumbnail"
const ThumbnailDescription = "preservation management - file thumbnail"

func NewThumbnailFS(fsys fs.FS, thumbnail *thumbnail.Thumbnail, logger zLogger.ZLogger) (*Thumbnail, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &ThumbnailConfig{
		StorageType:     "extension",
		StorageName:     "data",
		Ext:             "png",
		Width:           256,
		Height:          256,
		Compress:        "gzip",
		SingleDirectory: false,
	}
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
		return nil, errors.Wrap(err, "cannot create new thumbnail")
	}
	return ext, nil
}

type ThumbnailConfig struct {
	*ocfl.ExtensionConfig
	Compress        string `json:"compress"`
	ShortFilename   bool   `json:"shortFilename"`
	Ext             string `json:"ext"`
	Width           uint64 `json:"width"`
	Height          uint64 `json:"height"`
	SingleDirectory bool   `json:"singleDirectory"`
	StorageType     string `json:"storageType"`
	StorageName     string `json:"storageName"`
}

type ThumbnailTarget struct {
	Name           string
	Strategy       string        // add or replace
	FilenameRegexp regexp.Regexp // target expression
	Command        string        // command to execute (stdin --> stdout)
}

type ThumbnailResult struct {
	Ext          string   `json:"ext,omitempty"`
	Error        string   `json:"error,omitempty"`
	ID           string   `json:"id,omitempty"`
	Filename     string   `json:"filename,omitempty"`
	StorageType  string   `json:"storageType,omitempty"`
	ThumbDigest  string   `json:"thumbDigest,omitempty"`
	SourceDigest string   `json:"sourceDigest,omitempty"`
	SourceName   []string `json:"sourceName,omitempty"`
}

// map pronom to thumbnail
type ThumbnailMap map[string]*ThumbnailTarget

// map checksum to thumbnail
type ThumbnailFiles map[string]*ThumbnailTarget

func NewThumbnail(config *ThumbnailConfig, mig *thumbnail.Thumbnail, logger zLogger.ZLogger) (*Thumbnail, error) {
	sl := &Thumbnail{
		ThumbnailConfig: config,
		logger:          logger,
		thumbnail:       mig,
		buffer:          map[string]*bytes.Buffer{},
		counter:         map[string]int64{},
		streamInfo:      map[string]map[string]*ThumbnailResult{},
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

type Thumbnail struct {
	*ThumbnailConfig
	logger      zLogger.ZLogger
	fsys        fs.FS
	lastHead    string
	thumbnail   *thumbnail.Thumbnail
	buffer      map[string]*bytes.Buffer
	writer      *brotli.Writer
	sourceFS    fs.FS
	currentHead string
	done        bool
	counter     map[string]int64
	streamInfo  map[string]map[string]*ThumbnailResult
}

func (thumb *Thumbnail) Terminate() error {
	return nil
}

func (thumb *Thumbnail) GetFS() fs.FS {
	return thumb.fsys
}

func (thumb *Thumbnail) GetConfig() any {
	return thumb.ThumbnailConfig
}

func (thumb *Thumbnail) IsRegistered() bool { return false }

func (thumb *Thumbnail) GetName() string { return ThumbnailName }

func (thumb *Thumbnail) SetFS(fsys fs.FS, create bool) { thumb.fsys = fsys }

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

func (thumb *Thumbnail) storeThumbnail(object ocfl.Object, head string, mFile io.ReadCloser) (target string, digest string, err error) {
	var targetName string
	subfolder := thumb.StorageName
	if thumb.StorageType == "area" {
		subfolder = "thumbnails"
	}
	if thumb.SingleDirectory {
		targetName = fmt.Sprintf("%s/%s/%05d.%s", subfolder, head, thumb.counter[head], strings.ToLower(thumb.ThumbnailConfig.Ext))
	} else {
		cs := fmt.Sprintf("%02x", rand.Intn(64))
		targetName = fmt.Sprintf("%s/%s/%s/%s/%05d.%s", subfolder, head, string([]rune(cs)[0]), string([]rune(cs)[1]), thumb.counter[head], strings.ToLower(thumb.ThumbnailConfig.Ext))
	}
	switch strings.ToLower(thumb.StorageType) {
	case "area":
		if digest, err = object.AddReader(mFile, []string{targetName}, thumb.StorageName, true, false); err != nil {
			return "", "", errors.Wrapf(err, "cannot write '%s'", targetName)
		}
		areaPath, err := object.GetAreaPath(thumb.StorageName)
		if err != nil {
			return "", "", errors.Wrapf(err, "cannot get area path for '%s'", thumb.StorageName)
		}
		return fmt.Sprintf("%s/%s", areaPath, targetName), digest, nil
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return "", "", errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, thumb.StorageName, targetName)), "/")

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if digest, err = object.AddReader(mFile, []string{targetname}, "", true, false); err != nil {
			return "", "", errors.Wrapf(err, "cannot write '%s'", targetname)
		}
		return targetname, digest, nil
	case "extension":
		fp, err := writefs.Create(thumb.fsys, targetName)
		if err != nil {
			return "", "", errors.Wrapf(err, "cannot create file '%s/%s'", thumb.fsys, targetName)
		}
		if _, err := io.Copy(fp, mFile); err != nil {
			return "", "", errors.Wrapf(err, "cannot write file '%v/%s'", thumb.fsys, targetName)
		}
		if err := fp.Close(); err != nil {
			return "", "", errors.Wrapf(err, "cannot close file '%v/%s'", thumb.fsys, targetName)
		}
		return targetName, "", nil
	default:
		return "", "", errors.Errorf("unsupported storage type '%s'", thumb.StorageType)
	}
}

func (thumb *Thumbnail) DoThumbnail(object ocfl.Object, head string, thumbFunc *thumbnail.Function, ext string, file io.ReadCloser) (string, string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+ext)
	if err != nil {
		return "", "", errors.Wrap(err, "cannot create temp file")
	}
	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		return "", "", errors.Wrap(err, "cannot copy file")
	}
	if err := file.Close(); err != nil {
		return "", "", errors.Wrap(err, "cannot close file")
	}
	thumb.counter[head]++
	tmpFilename := filepath.ToSlash(tmpFile.Name())
	targetTempName := filepath.ToSlash(filepath.Join(filepath.Dir(tmpFilename), fmt.Sprintf("target.%s.%s", filepath.Base(tmpFilename), strings.ToLower(thumb.ThumbnailConfig.Ext))))

	if err := tmpFile.Close(); err != nil {
		return "", "", errors.Wrap(err, "cannot close temp file")
	}
	defer func() {
		if err := os.Remove(tmpFilename); err != nil {
			thumb.logger.Error().Err(err).Msgf("cannot remove temp file '%s'", tmpFilename)
		}
		if err := os.Remove(targetTempName); err != nil {
			thumb.logger.Error().Err(err).Msgf("cannot remove temp file '%s'", targetTempName)
		}
	}()
	if err := thumbFunc.Thumbnail(tmpFilename, targetTempName, thumb.ThumbnailConfig.Width, thumb.ThumbnailConfig.Height, thumb.logger); err != nil {
		//_ = os.Remove(tmpFilename)
		return "", "", errors.Wrapf(err, "cannot create thumbnail file '%v' to object '%s'", targetTempName, object.GetID())
	}

	mFile, err := os.Open(targetTempName)
	if err != nil {
		return "", "", errors.Wrapf(err, "cannot open file '%s'", targetTempName)
	}
	defer func() {
		if err := mFile.Close(); err != nil {
			thumb.logger.Error().Err(err).Msgf("cannot close file '%s'", targetTempName)
		}
	}()

	targetFile, digest, err := thumb.storeThumbnail(object, head, mFile)

	return targetFile, digest, errors.Wrap(err, "cannot store thumbnail")
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

	if _, ok := thumb.counter[head]; !ok {
		thumb.counter[head] = 0
	}

	// first get the metadata from the object
	meta, err := object.GetMetadata()
	if err != nil {
		return errors.Wrapf(err, "cannot get metadata from object %s", object.GetID())
	}
	for cs, m := range meta.Files {
		var found *ThumbnailResult = nil
		for name, info := range thumb.streamInfo[head] {
			if slices.Contains(m.InternalName, name) {
				found = info
				found.SourceDigest = cs
				break
			}
		}
		if found == nil {
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
					thumb.logger.Info().Msgf("create thumbnail for %s", m.InternalName[0])
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
				}
			}
			if file != nil {
				//var ml *ThumbnailResult
				var errStr string
				targetFile, digest, err := thumb.DoThumbnail(object, head, thumbnailFunction, ext, file)
				if err != nil {
					errStr = err.Error()
				}
				found = &ThumbnailResult{
					SourceDigest: cs,
					Filename:     targetFile,
					Ext:          thumb.ThumbnailConfig.Ext,
					Error:        errStr,
					ID:           thumbnailFunction.GetID(),
					ThumbDigest:  digest,
				}
			}
		}
		if found != nil {
			data, err := json.Marshal(found)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal thumbnail line for file '%s' in object '%s'", found.Filename, object.GetID())
			}
			if _, err := thumb.writer.Write(append(data, []byte("\n")...)); err != nil {
				return errors.Wrapf(err, "cannot write thumbnail line for file '%s' in object '%s'", found.Filename, object.GetID())
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
		thumb.StorageType,
		thumb.StorageName,
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
	/*
		path2digest := map[string]string{}
		for checksum, names := range manifest {
			for _, name := range names {
				path2digest[name] = checksum
			}
		}
	*/
	versions := inventory.GetVersions()
	for v := range versions {
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
			data, err = ocfl.ReadJsonL(object, "thumbnail", v, thumb.ThumbnailConfig.Compress, thumb.StorageType, thumb.StorageName, thumb.fsys)
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
			var meta = ThumbnailResult{}
			if err := json.Unmarshal([]byte(line), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", object.GetID(), v, line)
			}
			meta.SourceDigest = strings.ToLower(meta.SourceDigest) // paranoia
			meta.ThumbDigest = strings.ToLower(meta.ThumbDigest)   // paranoia
			meta.StorageType = thumb.StorageType

			// just to make sure, that we have a corresponding file in manifest

			_, ok := manifest[meta.SourceDigest]
			if !ok {
				return nil, errors.Errorf("cannot find checksum for file '%s' in object '%s'", meta.SourceDigest, object.GetID())
			}

			if _, ok := manifest[meta.ThumbDigest]; ok {
				source := ""
				if state, err := inventory.GetStateFiles(inventory.GetHead(), meta.SourceDigest); err == nil && len(state) > 0 {
					source = state[0]
				}
				if meta.SourceName == nil {
					meta.SourceName = []string{}
				}
				meta.SourceName = append(meta.SourceName, source)
				result[meta.ThumbDigest] = meta
			}
			// old versions do not have a filename
			if meta.Filename == "" {
				meta.Filename = fmt.Sprintf("data/%s/%s/%s.%s", string([]rune(meta.SourceDigest)[0]), string([]rune(meta.SourceDigest)[1]), meta.SourceDigest, strings.ToLower(thumb.ThumbnailConfig.Ext))
			}
			switch strings.ToLower(thumb.StorageType) {
			case "area":
				areaPath, err := object.GetAreaPath(thumb.StorageName)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot get area path for '%s'", thumb.StorageName)
				}
				meta.Filename = fmt.Sprintf("%s/%s", areaPath, meta.Filename)
			case "path":
				areaPath, err := object.GetAreaPath("content")
				if err != nil {
					return nil, errors.Wrapf(err, "cannot get area path for '%s'", thumb.StorageName)
				}
				if areaPath == "" {
					meta.Filename = fmt.Sprintf("%s/%s", thumb.StorageName, meta.Filename)
				} else {
					meta.Filename = fmt.Sprintf("%s/%s/%s", areaPath, thumb.StorageName, meta.Filename)
				}
			case "extension":
				//				meta.Filename = fmt.Sprintf("extension/%s", meta.Filename)
			default:
				return nil, errors.Errorf("unsupported storage type '%s'", thumb.StorageType)
			}

			meta.StorageType = thumb.StorageType
			result[meta.SourceDigest] = &meta
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
