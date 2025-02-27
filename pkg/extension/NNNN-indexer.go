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

	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	ironmaiden "github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
)

const (
	IndexerName        = "NNNN-indexer"
	IndexerDescription = "technical metadata for all files"

	// Errors.
	errorExtensionConfig = "ErrorExtensionConfig"
	errorExtensionRunner = "ErrorExtensionRunner"
)

type indexerLine struct {
	Path    string
	Indexer *ironmaiden.ResultV2
}

var actions = []string{
	ironmaiden.NameSiegfried,
	ironmaiden.NameFFProbe,
	ironmaiden.NameIdentify,
	ironmaiden.NameTika,
	ironmaiden.NameFullText,
	ironmaiden.NameXML,
}
var compress = []string{"brotli", "gzip", "none"}

type errorID = archiveerror.ID

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

func NewIndexerFS(
	fsys fs.FS,
	urlString string,
	indexerActions *ironmaiden.ActionDispatcher,
	localCache bool,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*Indexer, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errorFactory.NewError(errorExtensionConfig, "cannot open config.json", err)
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errorFactory.NewError(errorExtensionConfig, "cannot read config.json", err)
	}
	var config = &IndexerConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errorFactory.NewError(
			errorExtensionConfig, fmt.Sprintf("cannot unmarshal DirectCleanConfig '%s'", string(data)),
			err,
		)
	}
	ext, err := NewIndexer(config, urlString, indexerActions, localCache, logger, errorFactory)
	if err != nil {
		return nil, errorFactory.NewError(errorExtensionConfig, "cannot create new indexer", err)
	}
	return ext, nil
}

func NewIndexer(
	config *IndexerConfig,
	urlString string,
	indexerActions *ironmaiden.ActionDispatcher,
	localCache bool,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*Indexer, error) {
	var err error
	if config.Actions == nil {
		config.Actions = []string{}
	}
	as := []string{}
	for _, a := range config.Actions {
		a = strings.ToLower(a)
		if !slices.Contains(actions, a) {
			return nil, errorFactory.NewError(
				errorExtensionConfig, fmt.Sprintf("invalid action '%s' in config file", a),
				err,
			)
		}
		as = append(as, a)
	}
	config.Actions = as
	if config.Compress == "" {
		config.Compress = "none"
	}
	c := strings.ToLower(config.Compress)
	if !slices.Contains(compress, c) {
		return nil, errorFactory.NewError(
			errorExtensionConfig, fmt.Sprintf("invalid compression '%s' in config file", c),
			err,
		)
	}
	config.Compress = c
	sl := &Indexer{
		IndexerConfig:  config,
		buffer:         map[string]*bytes.Buffer{},
		active:         true,
		indexerActions: indexerActions,
		localCache:     localCache,
		logger:         logger,
		errorFactory:   errorFactory,
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		return nil, errorFactory.NewError(
			errorExtensionConfig, fmt.Sprintf("cannot parse url '%s'", urlString),
			err,
		)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errorFactory.NewError(
			errorExtensionConfig,
			fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()),
			err,
		)
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
	errorFactory   *archiveerror.Factory
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
				factoryErr := sl.errorFactory.NewError(
					errorExtensionConfig,
					fmt.Sprintf("cannot post to '%s'", urlString),
					err,
				)
				return factoryErr
			}
			if code != http.StatusBadRequest {
				factoryErr := sl.errorFactory.NewError(
					errorExtensionConfig,
					fmt.Sprintf("cannot post to '%s' - %v:'%s'", urlString, code, result),
					nil,
				)
				return factoryErr
			}
			_ = result
			return nil
		}
		return nil
	}
	if sl.indexerURL, err = url.Parse(urlString); err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig,
			fmt.Sprintf("cannot parse '%s' '%s'", name, urlString),
			err,
		)
		return factoryErr
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	result, code, err := sl.post("")
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig,
			fmt.Sprintf("cannot post to '%s'", urlString),
			err,
		)
		return factoryErr
	}
	if code != http.StatusBadRequest {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig,
			fmt.Sprintf("cannot post to '%s' - %v:'%s'", urlString, code, result),
			nil,
		)
		return factoryErr
	}
	_ = result

	return nil
}

func (sl *Indexer) post(data any) ([]byte, int, error) {

	const statusErr int = 0

	if !(sl.indexerURL != nil && sl.indexerURL.String() != "") {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner, "indexer url not set", nil,
		)
		return nil, statusErr, factoryErr
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner,
			fmt.Sprintf("cannot marshal %v", data),
			err,
		)
		return nil, statusErr, factoryErr
	}
	resp, err := http.Post(sl.indexerURL.String(), "test/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner,
			fmt.Sprintf("cannot post %v to %s", data, sl.indexerURL),
			err,
		)
		return nil, statusErr, factoryErr
	}
	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner,
			fmt.Sprintf("cannot read result of post %v to %s", data, sl.indexerURL),
			err,
		)
		return nil, statusErr, factoryErr
	}
	return result, resp.StatusCode, nil
}

func (sl *Indexer) WriteConfig() error {
	if sl.fsys == nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig, "no filesystem set", nil,
		)
		return factoryErr
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig, "cannot open config.json", err,
		)
		return factoryErr
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.IndexerConfig); err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig, "cannot encode config to file", err,
		)
		return factoryErr
	}
	return nil
}

func (sl *Indexer) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (sl *Indexer) UpdateObjectAfter(object ocfl.Object) error {
	if sl.indexerActions == nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig, "please enable indexer in config file", nil,
		)
		return factoryErr
	}
	if sl.writer == nil {
		return nil
	}
	if err := sl.writer.Flush(); err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner, "cannot flush brotli writer", err,
		)
		return factoryErr
	}
	if err := sl.writer.Close(); err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner, "cannot close brotli writer", err,
		)
		return factoryErr
	}
	head := object.GetInventory().GetHead()
	if head == "" {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner,
			fmt.Sprintf("no head for object '%s'", object.GetID()),
			nil,
		)
		return factoryErr
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
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner, "cannot write jsonl", err,
		)
		return factoryErr
	}
	return nil
}

// GetMetadata returns the metadata organized by checksum for every
// object passed to indexing to the caller.
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
				factoryErr := sl.errorFactory.NewError(
					errorExtensionRunner,
					fmt.Sprintf("cannot read buffer for '%s' '%s'", object.GetID(), v),
					err,
				)
				return nil, factoryErr
			}
		} else {
			data, err = ocfl.ReadJsonL(object, "indexer", v, sl.IndexerConfig.Compress, sl.StorageType, sl.StorageName, sl.fsys)
			if err != nil {
				factoryErr := sl.errorFactory.NewError(
					errorExtensionRunner,
					fmt.Sprintf("cannot read jsonl for '%s' version '%s'", object.GetID(), v),
					err,
				)
				return nil, factoryErr
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
				factoryErr := sl.errorFactory.NewError(
					errorExtensionRunner,
					fmt.Sprintf("cannot unmarshal line from for '%s' %s - [%s]", object.GetID(), v, line),
					err,
				)
				return nil, factoryErr
			}
			result[path2digest[meta.Path]] = meta.Indexer
		}
		if err := r.Err(); err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner,
				fmt.Sprintf("cannot scan lines for '%s' %s", object.GetID(), v),
				err,
			)
			return nil, factoryErr
		}
	}
	return result, nil
}

// errorfactory lookup constants.
const (
	errorIndexerSiegfried = "ErrorIndexerSiegfried"
	errorIndexerXML       = "ErrorIndexerXML"
	errorIndexerChecksum  = "ErrorIndexerChecksum"
	errorIndexerTika      = "ErrorIndexerTika"
	errorIndexerFFProbe   = "ErrorIndexerFFProbe"
	errorIndexerIdentify  = "ErrorIndexerIdentify"
	errorIndexerFullText  = "ErrorIndexerFullText"
)

// extractIndexErrors extracts from the metadata errors that have
// occured that might impact archiving or preservation.
func (sl *Indexer) extractIndexErrors(indexErrors map[string]string) {
	for k, v := range indexErrors {
		switch k {
		// this switch treats all errors the same at the tooling level,
		// in future, these can be parsed so that they can return
		// specificly weighted errors to the caller.
		case ironmaiden.NameSiegfried:
			err := sl.errorFactory.NewError(errorIndexerSiegfried, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameXML:
			err := sl.errorFactory.NewError(errorIndexerXML, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameChecksum:
			err := sl.errorFactory.NewError(errorIndexerChecksum, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameTika:
			err := sl.errorFactory.NewError(errorIndexerTika, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameFFProbe:
			err := sl.errorFactory.NewError(errorIndexerFFProbe, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameIdentify:
			err := sl.errorFactory.NewError(errorIndexerIdentify, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		case ironmaiden.NameFullText:
			err := sl.errorFactory.NewError(errorIndexerFullText, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg("")
		default:
			err := sl.errorFactory.NewError(archiveerror.IDUnknownError, v, nil)
			sl.logger.Error().Any(fmt.Sprintf("%s", err.Type), err).Msg(
				"unknown indexer error in dispatcher",
			)
		}
	}
}

// StreamObject streams the data to the Indexer module to be analysed
// by the different Indexer components.
func (sl *Indexer) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
	if !sl.active {
		return nil
	}
	if sl.indexerActions == nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionConfig,
			"please enable indexer in config file",
			nil,
		)
		return factoryErr
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
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner, "no statefiles", nil,
			)
			return factoryErr
		}
		tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+filepath.Ext(stateFiles[0]))
		if err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner, "cannot create temp file", err,
			)
			return factoryErr
		}
		fi, err := tmpFile.Stat()
		if err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner, "cannot stat tempfile", err,
			)
			return factoryErr
		}
		tmpFilename := filepath.ToSlash(filepath.Join(os.TempDir(), fi.Name()))
		if _, err := io.Copy(tmpFile, reader); err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner, "cannot write to tempfile", err,
			)
			return factoryErr
		}
		tmpFile.Close()
		result, err = sl.indexerActions.DoV2(tmpFilename, stateFiles, sl.Actions)
		os.Remove(tmpFilename)
	} else {
		result, err = sl.indexerActions.Stream(reader, stateFiles, sl.Actions)
	}
	if err != nil {
		factoryErr := sl.errorFactory.NewError(
			errorExtensionRunner,
			fmt.Sprintf("cannot index '%s'", stateFiles),
			err,
		)
		return factoryErr
	}
	if result != nil {
		// errors hidden in the metadata extract need to be discovered
		// and returned to aid in analysis.
		sl.extractIndexErrors(result.Errors)
		// build our JSON object and output it to JSONL.
		var indexerline = indexerLine{
			Path:    filepath.ToSlash(inventory.BuildManifestName(dest)),
			Indexer: result,
		}
		data, err := json.Marshal(indexerline)
		if err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner,
				fmt.Sprintf("cannot marshal result %v", indexerline),
				nil,
			)
			return factoryErr
		}
		if _, err := sl.writer.Write(append(data, []byte("\n")...)); err != nil {
			factoryErr := sl.errorFactory.NewError(
				errorExtensionRunner,
				fmt.Sprintf("cannot brotli %s", string(data)),
				nil,
			)
			return factoryErr
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
