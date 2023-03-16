package extension

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	ironmaiden "github.com/je4/indexer/pkg/indexer"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

const IndexerName = "NNNN-indexer"
const IndexerDescription = "technical metadata for all files"

type indexerLine struct {
	Path    string
	Indexer *ironmaiden.ResultV2
}

var actions = []string{"siegfried", "ffprobe", "identify", "tika"}
var compress = []string{"brotli", "gzip", "none"}

func GetIndexerParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: IndexerName,
			Param:         "addr",
			File:          "Addr",
			Description:   "url for indexer format recognition service",
		},
	}
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
	fs             ocfl.OCFLFSRead
	indexerURL     *url.URL
	buffer         *bytes.Buffer
	writer         *brotli.Writer
	active         bool
	indexerActions *ironmaiden.ActionDispatcher
}

func NewIndexerFS(fs ocfl.OCFLFSRead, urlString string, indexerActions *ironmaiden.ActionDispatcher) (*Indexer, error) {
	fp, err := fs.Open("config.json")
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
	ext, err := NewIndexer(config, urlString, indexerActions)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new indexer")
	}
	return ext, nil
}
func NewIndexer(config *IndexerConfig, urlString string, indexerActions *ironmaiden.ActionDispatcher) (*Indexer, error) {
	var err error

	if len(config.Actions) == 0 {
		config.Actions = []string{"siegfried"}
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
		buffer:         new(bytes.Buffer),
		active:         true,
		indexerActions: indexerActions,
	}
	sl.writer = brotli.NewWriter(sl.buffer)
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		return nil, errors.Wrapf(err, "cannot parse url '%s'", urlString)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (sl *Indexer) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.IndexerConfig, "", "  ")
	return string(str)
}

func (sl *Indexer) IsRegistered() bool { return false }

func (sl *Indexer) GetName() string { return IndexerName }

func (sl *Indexer) SetFS(fs ocfl.OCFLFSRead) { sl.fs = fs }

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
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	fsRW, ok := sl.fs.(ocfl.OCFLFS)
	if !ok {
		return errors.Errorf("filesystem is read only - '%s'", sl.fs.String())
	}

	configWriter, err := fsRW.Create("config.json")
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

/*
	func (sl *Indexer) AddFileBefore(object ocfl.Object, sourceFS ocfl.OCFLFSRead, source, dest string) error {
		return nil
	}

	func (sl *Indexer) UpdateFileBefore(object ocfl.Object, sourceFS ocfl.OCFLFSRead, source, dest string) error {
		return nil
	}

	func (sl *Indexer) DeleteFileBefore(object ocfl.Object, dest string) error {
		// nothing to do
		return nil
	}

	func (sl *Indexer) AddFileAfter(object ocfl.Object, sourceFS ocfl.OCFLFSRead, source, internalPath, digest string) error {
		if !sl.active {
			return nil
		}

		var meta *ironmaiden.ResultV2
		if sl.indexerActions != nil {
			f, err := sourceFS.Open(source)
			if err != nil {
				return errors.Wrapf(err, "cannot open '%s/%s'", sourceFS.String(), source)
			}
			defer f.Close()

			meta, err = sl.indexerActions.Stream(f, source)
			if err != nil {
				return errors.Wrapf(err, "cannot process '%s/%s'", sourceFS.String(), source)
			}
		} else {
			filePath := fmt.Sprintf("%s/%s", sourceFS.String(), source)
			param := ironmaiden.ActionParam{
				Url:        filePath,
				Actions:    sl.IndexerConfig.Actions,
				HeaderSize: 0,
				Checksums:  map[string]string{},
			}
			result, code, err := sl.post(param)
			if err != nil {
				return errors.Wrapf(err, "indexer error for '%s'", filePath)
			}
			if code >= 300 {
				return errors.Errorf("indexer error for '%s': %s", filePath, result)
			}
			//		var meta = ironmaiden.ResultV2{}
			if err := json.Unmarshal(result, &meta); err != nil {
				return errors.Errorf("cannot unmarshal indexer result `%s`", string(result))
			}
		}
		var indexerline = indexerLine{
			Path:    internalPath,
			Indexer: meta,
		}
		data, err := json.Marshal(indexerline)
		if err != nil {
			return errors.Errorf("cannot marshal result %v", indexerline)
		}
		if _, err := sl.writer.Write(data); err != nil {
			return errors.Errorf("cannot brotli %s", string(data))
		}
		if _, err := sl.writer.Write([]byte("\n")); err != nil {
			return errors.Errorf("cannot brotli %s", string(data))
		}
		return nil
	}

	func (sl *Indexer) UpdateFileAfter(object ocfl.Object, sourceFS ocfl.OCFLFSRead, source, dest string) error {
		return nil
	}

	func (sl *Indexer) DeleteFileAfter(object ocfl.Object, dest string) error {
		// nothing to do
		return nil
	}
*/

func (sl *Indexer) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (sl *Indexer) UpdateObjectAfter(object ocfl.Object) error {
	//var err error
	sl.active = false
	if err := sl.writer.Flush(); err != nil {
		return errors.Wrap(err, "cannot flush brotli writer")
	}
	if err := sl.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close brotli writer")
	}
	var reader io.Reader
	var ext string
	switch sl.IndexerConfig.Compress {
	case "brotli":
		ext = ".br"
		reader = sl.buffer
	case "gzip":
		ext = ".gz"
		brotliReader := brotli.NewReader(sl.buffer)
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			gzipWriter := gzip.NewWriter(pw)
			defer gzipWriter.Close()
			if _, err := io.Copy(gzipWriter, brotliReader); err != nil {
				pw.CloseWithError(errors.Wrapf(err, "error on gzip compressor"))
			}
		}()
		reader = pr
	case "none":
		reader = brotli.NewReader(sl.buffer)
	default:
		return errors.Errorf("invalid compression '%s'", sl.IndexerConfig.Compress)
	}

	switch sl.StorageType {
	case "area":
		targetname := fmt.Sprintf("indexer_%s.jsonl%s", object.GetInventory().GetHead(), ext)
		if err := object.AddReader(io.NopCloser(reader), targetname, sl.IndexerConfig.StorageName); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		targetname := fmt.Sprintf("%s/indexer_%s.jsonl%s", sl.IndexerConfig.StorageName, object.GetInventory().GetHead(), ext)
		if err := object.AddReader(io.NopCloser(reader), targetname, "content"); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		fsRW, ok := sl.fs.(ocfl.OCFLFS)
		if !ok {
			return errors.Errorf("filesystem is read only - '%s'", sl.fs.String())
		}

		targetname := strings.TrimLeft(fmt.Sprintf("%s/indexer_%s.jsonl%s", sl.IndexerConfig.StorageName, object.GetInventory().GetHead(), ext), "/")
		fp, err := fsRW.Create(targetname)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s/%s'", sl.fs.String(), targetname)
		}
		defer fp.Close()
		if _, err := io.Copy(fp, reader); err != nil {
			return errors.Wrapf(err, "cannot write '%s/%s'", sl.fs.String(), targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", sl.StorageType)
	}
	return nil
}

func (sl *Indexer) GetMetadata(object ocfl.Object) (map[string]any, error) {
	var result = map[string]any{}
	var ext string
	switch sl.IndexerConfig.Compress {
	case "brotli":
		ext = ".br"
	case "gzip":
		ext = ".gz"
	case "none":
	default:
		return nil, errors.Errorf("invalid compression '%s'", sl.IndexerConfig.Compress)
	}

	inventory := object.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v, _ := range inventory.GetVersions() {
		var targetname string
		switch sl.StorageType {
		case "area":
			path, err := object.GetAreaPath(sl.IndexerConfig.StorageName)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot get area path for '%s'", sl.IndexerConfig.StorageName)
			}
			targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, path, v, ext)
		case "path":
			targetname = fmt.Sprintf("%s/indexer_%s.jsonl%s", sl.IndexerConfig.StorageName, v, ext)
		case "extension":
			targetname = strings.TrimLeft(fmt.Sprintf("%s/indexer_%s.jsonl%s", sl.IndexerConfig.StorageName, v, ext), "/")
		default:
			return nil, errors.Errorf("unsupported storage type '%s'", sl.StorageType)
		}

		f, err := object.GetFS().Open(targetname)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		var reader io.Reader
		switch sl.IndexerConfig.Compress {
		case "brotli":
			reader = brotli.NewReader(f)
		case "gzip":
			reader, err = gzip.NewReader(f)
			if err != nil {
				f.Close()
				return nil, errors.Wrapf(err, "cannot open gzip reader on '%s'", targetname)
			}
		case "none":
			reader = f
		}
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 128*1024)
		r.Split(bufio.ScanLines)
		for r.Scan() {
			line := r.Text()
			var meta = indexerLine{}
			if err := json.Unmarshal([]byte(line), &meta); err != nil {
				_ = f.Close()
				return nil, errors.Wrapf(err, "cannot unmarshal line from '%s' - [%s]", targetname, line)
			}
			result[path2digest[meta.Path]] = meta.Indexer
		}
		if err := r.Err(); err != nil {
			return nil, errors.Wrapf(err, "cannot scan lines from '%s'", targetname)
		}
		f.Close()
	}
	return result, nil
}

func (sl *Indexer) StreamObject(object ocfl.Object, reader io.Reader, source, dest string) error {
	if !sl.active {
		return nil
	}

	result, err := sl.indexerActions.Stream(reader, source)
	if err != nil {
		return errors.Wrapf(err, "cannot index '%s'", source)
	}
	inventory := object.GetInventory()
	head := inventory.GetHead()
	var indexerline = indexerLine{
		Path:    filepath.ToSlash(filepath.Join(head, "content", dest)),
		Indexer: result,
	}
	data, err := json.Marshal(indexerline)
	if err != nil {
		return errors.Errorf("cannot marshal result %v", indexerline)
	}
	if _, err := sl.writer.Write(data); err != nil {
		return errors.Errorf("cannot brotli %s", string(data))
	}
	if _, err := sl.writer.Write([]byte("\n")); err != nil {
		return errors.Errorf("cannot brotli %s", string(data))
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
