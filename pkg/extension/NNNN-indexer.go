package extension

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	ironmaiden "github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const IndexerName = "NNNN-indexer"
const IndexerDescription = "technical metadata for all files"

type indexerLine struct {
	Path    string
	Indexer *ironmaiden.ResultV2
}

var actions = []string{"siegfried", "ffprobe", "identify", "tika", "fulltext", "xml"}
var compress = []string{"brotli", "gzip", "none"}

func GetIndexerParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: IndexerName,
			Param:         "addr",
			//File:          "Addr",
			Description: "url for indexer format recognition service",
		},
	}
}

func NewIndexerFS(fsys fs.FS, urlString string, indexerActions *ironmaiden.ActionDispatcher, localCache bool, logger zLogger.ZLogger) (*Indexer, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &IndexerConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	ext, err := NewIndexer(config, urlString, indexerActions, localCache, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new indexer")
	}
	return ext, nil
}
func NewIndexer(config *IndexerConfig, urlString string, indexerActions *ironmaiden.ActionDispatcher, localCache bool, logger zLogger.ZLogger) (*Indexer, error) {
	var err error

	if config.Actions == nil {
		config.Actions = []string{}
	}
	as := []string{}
	for _, a := range config.Actions {
		a = strings.ToLower(a)
		if !slices.Contains(actions, a) {
			return nil, errors.Errorf("invalid action '%s' in config file", a)
		}
		as = append(as, a)
	}
	config.Actions = as

	if config.Compress == "" {
		config.Compress = "none"
	}
	c := strings.ToLower(config.Compress)
	if !slices.Contains(compress, c) {
		return nil, errors.Errorf("invalid compression '%s' in config file", c)
	}
	config.Compress = c

	sl := &Indexer{
		IndexerConfig:  config,
		buffer:         map[string]*bytes.Buffer{},
		active:         true,
		indexerActions: indexerActions,
		localCache:     localCache,
		logger:         logger,
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		return nil, errors.Wrapf(err, "cannot parse url '%s'", urlString)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type IndexerConfig struct {
	*ocfl.ExtensionConfig
	StorageType string
	StorageName string
	Actions     []string
	Compress    string
}
type Indexer struct {
	*IndexerConfig
	fsys           fs.FS
	indexerURL     *url.URL
	buffer         map[string]*bytes.Buffer
	writer         *brotli.Writer
	active         bool
	indexerActions *ironmaiden.ActionDispatcher
	currentHead    string
	localCache     bool
	logger         zLogger.ZLogger
}

func (sl *Indexer) Terminate() error {
	return nil
}

func (sl *Indexer) GetFS() fs.FS {
	return sl.fsys
}

func (sl *Indexer) GetConfig() any {
	return sl.IndexerConfig
}

func (sl *Indexer) IsRegistered() bool { return false }

func (sl *Indexer) GetName() string { return IndexerName }

func (sl *Indexer) SetFS(fsys fs.FS, create bool) { sl.fsys = fsys }

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

func (sl *Indexer) WriteConfig() error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
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

func (sl *Indexer) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (sl *Indexer) UpdateObjectAfter(object ocfl.Object) error {
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
	head := object.GetInventory().GetHead()
	if head == "" {
		return errors.Errorf("no head for object '%s'", object.GetID())
	}
	buffer, ok := sl.buffer[head]
	if !ok {
		return nil
	}
	if err := ocfl.WriteJsonL(
		object,
		"indexer",
		buffer.Bytes(),
		sl.IndexerConfig.Compress,
		sl.StorageType,
		sl.StorageName,
		sl.fsys,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}
	return nil
}

func (sl *Indexer) GetMetadata(object ocfl.Object) (map[string]any, error) {
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
	for v, _ := range inventory.GetVersions() {
		var data []byte
		if buf, ok := sl.buffer[v]; ok && buf.Len() > 0 {
			//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
			// need a new reader on the buffer
			reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
			data, err = io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", object.GetID(), v)
			}
		} else {
			data, err = ocfl.ReadJsonL(object, "indexer", v, sl.IndexerConfig.Compress, sl.StorageType, sl.StorageName, sl.fsys)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", object.GetID(), v)
			}
		}

		reader := bytes.NewReader(data)
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 16*1024*1024)
		r.Split(bufio.ScanLines)
		for r.Scan() {
			line := r.Text()
			var meta = indexerLine{}
			if err := json.Unmarshal([]byte(line), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", object.GetID(), v, line)
			}
			result[path2digest[meta.Path]] = meta.Indexer
		}
		if err := r.Err(); err != nil {
			return nil, errors.Wrapf(err, "cannot scan lines for '%s' %s", object.GetID(), v)
		}
	}
	return result, nil
}

func (sl *Indexer) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
	if !sl.active {
		return nil
	}
	if sl.indexerActions == nil {
		return errors.New("Please enable indexer in config file")
	}

	inventory := object.GetInventory()
	head := inventory.GetHead()
	if _, ok := sl.buffer[head]; !ok {
		sl.buffer[head] = &bytes.Buffer{}
	}
	if sl.currentHead != head {
		sl.writer = brotli.NewWriter(sl.buffer[head])
		sl.currentHead = head
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
	if result != nil {
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
	_ ocfl.Extension = &Indexer{}
	//	_ ocfl.ExtensionContentChange = &Indexer{}
	_ ocfl.ExtensionObjectChange = &Indexer{}
	_ ocfl.ExtensionMetadata     = &Indexer{}
	_ ocfl.ExtensionStream       = &Indexer{}
)
