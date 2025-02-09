package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MetaFileName = "NNNN-metafile"
const MetaFileDescription = "adds a file in extension folder"

func GetMetaFileParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"add", "update", "create"},
			Param:         "source",
			//File:          "Source",
			Description: "url or path with metadata file. $ID will be replaced with object ID i.e. c:/temp/$ID.json",
		},
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"extract", "objectextension"},
			Param:         "target",
			//File:          "Target",
			Description: "url or path with metadata target folder",
		},
	}
}

func NewMetaFileFS(fsys fs.FS) (*MetaFile, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetaFileConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}

	if config.MetaName == "config.json" {
		return nil, errors.Errorf("config.json is not allowed for field name in %v/%s", fsys, "config.json")
	}
	if config.MetaSchema == "config.json" {
		return nil, errors.Errorf("config.json is not allowed for field schema in %v/%s", fsys, "config.json")
	}
	var schema []byte
	if config.MetaSchema != "" {
		schema, err = fs.ReadFile(fsys, config.MetaSchema)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read metadata schema %v/%s", fsys, config.MetaSchema)
		}
	} else {
		resp, err := http.Get(config.MetaSchemaUrl)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot load metadata schema %s", config.MetaSchemaUrl)
		}
		schema, err = io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("error loading metadata schema %s - [%v]%s - %s", resp.StatusCode, resp.Status, schema)
		}
		config.MetaSchema = "schema.json"
	}

	return NewMetaFile(config, schema)
}
func NewMetaFile(config *MetaFileConfig, schema []byte) (*MetaFile, error) {
	var err error
	sl := &MetaFile{
		MetaFileConfig: config,
		schema:         schema,
		info:           map[string][]byte{},
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	sl.compiledSchema, err = jsonschema.CompileString(sl.MetaSchemaUrl, string(sl.schema))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot compile schema")
	}
	return sl, nil
}

type MetaFileConfig struct {
	*ocfl.ExtensionConfig
	StorageType   string `json:"storageType"`
	StorageName   string `json:"storageName"`
	MetaName      string `json:"name,omitempty"`
	MetaSchema    string `json:"schema,omitempty"`
	MetaSchemaUrl string `json:"schemaUrl,omitempty"`
}
type MetaFile struct {
	*MetaFileConfig
	schema         []byte
	metadataSource string
	fsys           fs.FS
	compiledSchema *jsonschema.Schema
	stored         bool
	info           map[string][]byte
}

func (sl *MetaFile) Terminate() error {
	return nil
}

func (sl *MetaFile) GetFS() fs.FS {
	return sl.fsys
}

func (sl *MetaFile) GetConfig() any {
	return sl.MetaFileConfig
}

func (sl *MetaFile) IsRegistered() bool {
	return false
}

func (sl *MetaFile) SetParams(params map[string]string) error {
	if params != nil {
		name := fmt.Sprintf("ext-%s-%s", MetaFileName, "source")
		if urlString, ok := params[name]; ok {
			urlString = strings.TrimSpace(urlString)
			if urlString == "" {
				return errors.Errorf("no value for parameter '%s'", name)
			}
			sl.metadataSource = urlString
		}
	}
	return nil
}

func (sl *MetaFile) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *MetaFile) GetName() string { return MetaFileName }

func (sl *MetaFile) WriteConfig() error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	if _, err := writefs.WriteFile(sl.fsys, sl.MetaSchema, sl.schema); err != nil {
		return errors.Wrapf(err, "cannot write schema to %v/%s", sl.fsys, sl.MetaSchema)
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.MetaFileConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func toStringKeys(val interface{}) (interface{}, error) {
	var err error
	switch val := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			k, ok := k.(string)
			if !ok {
				return nil, errors.New("found non-string key")
			}
			m[k], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case []interface{}:
		var l = make([]interface{}, len(val))
		for i, v := range val {
			l[i], err = toStringKeys(v)
			if err != nil {
				return nil, err
			}
		}
		return l, nil
	default:
		return val, nil
	}
}

func (sl *MetaFile) UpdateObjectBefore(object ocfl.Object) error {
	if sl.stored {
		return nil
	}
	sl.stored = true
	var err error
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	if sl.metadataSource == "" {
		// only a problem, if first version
		if len(inventory.GetVersionStrings()) < 2 {
			return errors.New("no metadata source configured")
		}
		return nil
	}
	var rc io.ReadCloser
	var fname = strings.Replace(sl.metadataSource, "$ID", object.GetID(), -1)
	u, err := url.Parse(fname)
	if err == nil && u.Scheme != "" && u.Host != "" {
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	} else {
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	}
	defer rc.Close()

	var infoData []byte
	var info any

	switch strings.ToLower(filepath.Ext(fname)) {
	case ".json":
		jr := json.NewDecoder(rc)
		if err := jr.Decode(&info); err != nil {
			return errors.Wrap(err, "cannot decode info file")
		}
		if err := sl.compiledSchema.Validate(info); err != nil {
			return errors.Wrap(err, "cannot validate info file")
		}
	case ".yaml":
		jr := yaml.NewDecoder(rc)
		if err := jr.Decode(&info); err != nil {
			return errors.Wrap(err, "cannot decode info file")
		}
		info, err = toStringKeys(info)
		if err != nil {
			return errors.Wrap(err, "cannot convert map[any]any to map[string]any")
		}
		if err := sl.compiledSchema.Validate(info); err != nil {
			return errors.Wrap(err, "cannot validate info file")
		}
	case ".toml":
		jr := toml.NewDecoder(rc)
		if _, err := jr.Decode(&info); err != nil {
			return errors.Wrap(err, "cannot decode info file")
		}
		info, err = toStringKeys(info)
		if err != nil {
			return errors.Wrap(err, "cannot convert map[any]any to map[string]any")
		}
		if err := sl.compiledSchema.Validate(info); err != nil {
			return errors.Wrap(err, "cannot validate info file")
		}
	default:
		return errors.Errorf("unknown file extension in '%s' only .json, .toml and .yaml supported", fname)
	}

	infoData, err = json.MarshalIndent(info, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal info json")
	}

	switch strings.ToLower(sl.StorageType) {
	case "area":
		targetname := strings.TrimLeft(sl.MetaName, "/")
		if _, err := object.AddReader(io.NopCloser(bytes.NewBuffer(infoData)), []string{targetname}, sl.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, sl.StorageName, sl.MetaName)), "/")

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if _, err := object.AddReader(io.NopCloser(bytes.NewBuffer(infoData)), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(sl.StorageName, sl.MetaName)), "/")
		if _, err := writefs.WriteFile(sl.fsys, targetname, infoData); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", sl.fsys, targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", sl.StorageType)
	}

	// remember the content
	sl.info[inventory.GetHead()] = infoData
	return nil
}

func downloadFile(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get '%s'", u)
	}
	defer resp.Body.Close()
	metadata, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body from '%s'", u)
	}
	return metadata, nil
}

var windowsPathWithDrive = regexp.MustCompile("^/[a-zA-Z]:")

func (sl *MetaFile) UpdateObjectAfter(object ocfl.Object) error {
	return nil
}

func (sl *MetaFile) GetMetadata(object ocfl.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}
	inventory := object.GetInventory()
	versions := inventory.GetVersionStrings()
	slices.Reverse(versions)
	var metadata []byte
	for _, ver := range versions {
		var ok bool
		if metadata, ok = sl.info[ver]; ok {
			break
		}
		if metadata, err = ocfl.ReadFile(object, sl.MetaName, ver, sl.StorageType, sl.StorageName, sl.fsys); err == nil {
			break
		}
	}
	if metadata == nil {
		return nil, errors.Wrapf(err, "cannot read %s", sl.MetaName)
	}
	var metaStruct = map[string]any{}
	if err := json.Unmarshal(metadata, &metaStruct); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal '%s'", sl.MetaName)
	}
	result[""] = metaStruct
	return result, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &MetaFile{}
	_ ocfl.ExtensionObjectChange = &MetaFile{}
	_ ocfl.ExtensionMetadata     = &MetaFile{}
)
