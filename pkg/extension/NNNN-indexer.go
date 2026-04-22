package extension

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	ironmaiden "github.com/ocfl-archive/indexer/v3/pkg/indexer"
	indexerutil "github.com/ocfl-archive/indexer/v3/pkg/util"
	"golang.org/x/exp/slices"
)

const IndexerName = "NNNN-indexer"
const IndexerDescription = "technical metadata for all files"

func init() {
	extension.RegisterExtension(IndexerName, func() (extensiontypes.Extension, error) {
		return NewIndexer("", nil, &ironmaiden.IndexerConfig{}, false, nil)
	}, GetIndexerParams)
}

type indexerLine struct {
	Path    string
	Indexer *ironmaiden.ResultV2
}

var actions = []string{"siegfried", "ffprobe", "identify", "tika", "fulltext", "xml", "json", "checksum"}
var compress = []string{"brotli", "gzip", "none"}

func GetIndexerParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{
		{
			ExtensionName: IndexerName,
			Param:         "addr",
			//File:          "Addr",
			Description: "url for indexer format recognition service",
		},
	}, nil
}

func NewIndexer(urlString string, fss map[string]fs.FS, conf *ironmaiden.IndexerConfig, localCache bool, logger ocfllogger.OCFLLogger) (*Indexer, error) {
	var config = &IndexerConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{
			ExtensionName: IndexerName,
		},
		Actions:  []string{},
		Compress: "none",
	}
	sl := &Indexer{
		IndexerConfig: config,
		buffer:        map[string]*bytes.Buffer{},
		active:        true,
		localCache:    localCache,
	}

	if logger != nil {
		indexerActions, availableActions, indexerCloser, err := indexerutil.InitIndexer(conf, logger.Logger())
		//indexerActions, err := ironmaiden.InitActionDispatcher(fss, conf, logger.Logger())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot init indexer")
		}
		sl.indexerActions = indexerActions.ActionDispatcher()
		sl.availableActions = availableActions
		sl.indexerCloser = indexerCloser
	}
	var err error
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		return nil, err
	}
	return sl, nil
}

type IndexerConfig struct {
	*extensiontypes.ExtensionConfig
	StorageType string
	StorageName string
	Actions     []string
	Compress    string
}
type Indexer struct {
	*IndexerConfig
	indexerURL       *url.URL
	buffer           map[string]*bytes.Buffer
	writer           *brotli.Writer
	active           bool
	indexerActions   *ironmaiden.ActionDispatcher
	currentHead      string
	localCache       bool
	fsys             appendfs.FS
	logger           ocfllogger.OCFLLogger
	availableActions []string
	indexerCloser    io.Closer
}

func (sl *Indexer) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", IndexerName)
	return sl
}

func (sl *Indexer) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.IndexerConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal IndexerConfig '%s'", string(data))
	}
	// Normalize and validate actions
	if sl.IndexerConfig.Actions == nil {
		sl.IndexerConfig.Actions = []string{}
	}
	as := make([]string, 0, len(sl.IndexerConfig.Actions))
	for _, a := range sl.IndexerConfig.Actions {
		a = strings.ToLower(a)
		if !slices.Contains(actions, a) {
			return errors.Errorf("invalid action '%s' in config file", a)
		}
		as = append(as, a)
	}
	sl.IndexerConfig.Actions = as
	// Normalize and validate compression
	if sl.IndexerConfig.Compress == "" {
		sl.IndexerConfig.Compress = "none"
	}
	c := strings.ToLower(sl.IndexerConfig.Compress)
	if !slices.Contains(compress, c) {
		return errors.Errorf("invalid compression '%s' in config file", c)
	}
	sl.IndexerConfig.Compress = c
	return nil
}

func (sl *Indexer) Terminate() error {
	return nil
}

func (sl *Indexer) GetConfig() any {
	return sl.IndexerConfig
}

func (sl *Indexer) IsRegistered() bool { return false }

func (sl *Indexer) GetName() string { return IndexerName }

func (sl *Indexer) SetFS(fsys fs.FS, create bool) {
	if sfs, ok := fsys.(appendfs.FS); ok {
		sl.fsys = sfs
	}
}

func (sl *Indexer) GetFS() fs.FS {
	return sl.fsys
}

func (sl *Indexer) SetParams(params map[string]string) error {
	var err error
	name := fmt.Sprintf("ext-%s-%s", IndexerName, "addr")
	urlString, _ := params[name]
	if urlString == "" {
		if sl.indexerURL != nil && sl.indexerURL.String() != "" {
			result, code, err := sl.post("{}")
			if err != nil {
				return errors.Wrapf(err, "cannot post to '%s'", urlString)
			}
			if code != http.StatusBadRequest {
				return errors.Errorf("cannot post to '%s' - %v:'%s'", urlString, code, result)
			}
			_ = result
			return nil
		}
		return nil
		// return errors.Errorf("url '%s' not set", name)
	}
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		return errors.Wrapf(err, "cannot parse '%s' '%s'", name, urlString)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	result, code, err := sl.post("")
	if err != nil {
		return errors.Wrapf(err, "cannot post to '%s'", urlString)
	}
	if code != http.StatusBadRequest {
		return errors.Errorf("cannot post to '%s' - %v:'%s'", urlString, code, result)
	}
	_ = result

	return nil
}

func (sl *Indexer) post(data any) ([]byte, int, error) {
	if !(sl.indexerURL != nil && sl.indexerURL.String() != "") {
		return nil, 0, errors.New("indexer url not set")
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "cannot marshal %v", data)
	}
	resp, err := http.Post(sl.indexerURL.String(), "test/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, 0, errors.Wrapf(err, "cannot post %v to %s", data, sl.indexerURL)
	}
	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "cannot read result of post %v to %s", data, sl.indexerURL)
	}
	return result, resp.StatusCode, nil
}

func (sl *Indexer) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.IndexerConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *Indexer) UpdateObjectBefore(object object.VersionWriter) error {
	return nil
}

func (sl *Indexer) UpdateObjectAfter(object object.VersionWriter) error {
	if sl.indexerActions == nil {
		return errors.New("Please enable indexer in config file")
	}

	if sl.writer == nil {
		return nil
	}
	//var err error
	//	sl.active = false
	if err := sl.writer.Flush(); err != nil {
		return errors.Wrap(err, "cannot flush brotli writer")
	}
	if err := sl.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close brotli writer")
	}
	sl.writer = nil
	head := object.GetInventory().GetHead()
	if !head.IsValid() {
		return errors.Errorf("no head for object '%s'", object.GetID())
	}
	buffer, ok := sl.buffer[head.String()]
	if !ok {
		return nil
	}
	if err := WriteJsonL(
		sl.GetName(),
		object,
		"indexer",
		buffer.Bytes(),
		sl.IndexerConfig.Compress,
		sl.StorageType,
		sl.StorageName,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}
	return nil
}

func (sl *Indexer) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}

	inventory := obj.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest.Iterate() {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v, ver := range inventory.GetVersions().Iterate() {
		//for v := range inventory.GetVersions().GetVersionNumbers() {
		if ver.InCreation() {
			continue
		}
		var data []byte
		if buf, ok := sl.buffer[v.String()]; ok {
			if buf.Len() > 0 {
				//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
				// need a new reader on the buffer
				reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
				data, err = io.ReadAll(reader)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", obj.GetID(), v)
				}
			} else {
				data = nil
			}
		} else {
			data, err = ReadJsonL(sl.GetName(), sourceFS, obj, v, "indexer", sl.IndexerConfig.Compress, sl.StorageType, sl.StorageName)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", obj.GetID(), v)
			}
		}

		if data != nil {
			reader := bytes.NewReader(data)
			r := bufio.NewScanner(reader)
			r.Buffer(make([]byte, 128*1024), 16*1024*1024)
			r.Split(bufio.ScanLines)
			for r.Scan() {
				line := r.Text()
				var meta = indexerLine{}
				if err := json.Unmarshal([]byte(line), &meta); err != nil {
					return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", obj.GetID(), v, line)
				}
				result[path2digest[meta.Path]] = meta.Indexer
			}
			if err := r.Err(); err != nil {
				return nil, errors.Wrapf(err, "cannot scan lines for '%s' %s", obj.GetID(), v)
			}
		}
	}
	return result, nil
}

func (sl *Indexer) StreamObject(object object.VersionWriter, reader io.Reader, stateFiles []string, dest string) error {
	if !sl.active {
		return nil
	}
	if sl.indexerActions == nil {
		return errors.New("Please enable indexer in config file")
	}

	inventory := object.GetInventory()
	head := inventory.GetHead()
	if _, ok := sl.buffer[head.String()]; !ok {
		sl.buffer[head.String()] = &bytes.Buffer{}
	}
	if head.String() == "" || sl.currentHead != head.String() {
		sl.writer = brotli.NewWriter(sl.buffer[head.String()])
		sl.currentHead = head.String()
	}

	var result *ironmaiden.ResultV2
	var err error
	if sl.localCache {
		if len(stateFiles) == 0 {
			return errors.Wrapf(err, "no statefiles")
		}
		tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+filepath.Ext(stateFiles[0]))
		if err != nil {
			return errors.Wrapf(err, "cannot create temp file")
		}
		fi, err := tmpFile.Stat()
		if err != nil {
			return errors.Wrapf(err, "cannot stat tempfile")
		}
		tmpFilename := filepath.ToSlash(filepath.Join(os.TempDir(), fi.Name()))
		if _, err := io.Copy(tmpFile, reader); err != nil {
			return errors.Wrapf(err, "cannot write to tempfile")
		}
		tmpFile.Close()
		result, err = sl.indexerActions.DoV2(tmpFilename, stateFiles, sl.Actions)
		os.Remove(tmpFilename)
	} else {
		result, err = sl.indexerActions.Stream(reader, stateFiles, sl.Actions)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot index '%s'", stateFiles)
	}
	if result != nil && sl.writer != nil {
		var indexerline = indexerLine{
			Path:    filepath.ToSlash(inventory.BuildManifestName(dest)),
			Indexer: result,
		}
		data, err := json.Marshal(indexerline)
		if err != nil {
			return errors.Errorf("cannot marshal result %v", indexerline)
		}
		if _, err := sl.writer.Write(append(data, []byte("\n")...)); err != nil {
			return errors.Errorf("cannot brotli %s", string(data))
		}
	}
	return nil
}

var (
	_ extensiontypes.Extension = &Indexer{}
	//	_ ocfl.ExtensionContentChange = &Indexer{}
	_ object.ExtensionObjectChange = &Indexer{}
	_ object.ExtensionMetadata     = &Indexer{}
	_ object.ExtensionStream       = &Indexer{}
)
