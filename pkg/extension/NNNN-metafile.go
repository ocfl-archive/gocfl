package extension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/util"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v2"
)

const MetaFileName = "NNNN-metafile"
const MetaFileDescription = "adds a file in extension folder"

func init() {
	extension.RegisterExtension(MetaFileName, NewMetaFile, GetMetaFileParams)
}

func GetMetaFileParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"add", "update", "create"},
			Param:         "source",
			//File:          "Source",
			Description: "url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"extract", "objectextension"},
			Param:         "target",
			//File:          "Target",
			Description: "url with metadata target folder",
		},
	}, nil
}

func NewMetaFile() (extensiontypes.Extension, error) {
	var config = &MetaFileConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{
			ExtensionName: MetaFileName,
		},
	}
	sl := &MetaFile{
		MetaFileConfig: config,
		//schema:         schema,
		info: map[string][]byte{},
	}
	return sl, nil
}

type MetaFileConfig struct {
	*extensiontypes.ExtensionConfig
	StorageType   string `json:"storageType"`
	StorageName   string `json:"storageName"`
	MetaName      string `json:"name,omitempty"`
	MetaSchema    string `json:"schema,omitempty"`
	MetaSchemaUrl string `json:"schemaUrl,omitempty"`
}
type MetaFile struct {
	*MetaFileConfig
	schema         []byte
	metadataSource *url.URL
	compiledSchema *jsonschema.Schema
	stored         bool
	info           map[string][]byte
	logger         ocfllogger.OCFLLogger
}

func (sl *MetaFile) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", MetaFileName)
	return sl
}

func (sl *MetaFile) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, sl.MetaFileConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal MetaFileConfig '%s'", string(data))
	}

	if sl.MetaSchema != "" {
		sl.schema, err = fs.ReadFile(fsys, sl.MetaSchema)
		if err != nil {
			return errors.Wrapf(err, "cannot read metadata schema %v/%s", fsys, sl.MetaSchema)
		}
	} else {
		resp, err := http.Get(sl.MetaSchemaUrl)
		if err != nil {
			return errors.Wrapf(err, "cannot load metadata schema %s", sl.MetaSchemaUrl)
		}
		sl.schema, err = io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("error loading metadata schema %s - [%v]%s - %s", resp.StatusCode, resp.Status, sl.schema)
		}
		sl.MetaSchema = "schema.json"
	}
	sl.compiledSchema, err = jsonschema.CompileString(sl.MetaSchemaUrl, string(sl.schema))
	if err != nil {
		return errors.Wrapf(err, "cannot compile schema")
	}

	return nil
}

func (sl *MetaFile) Terminate() error {
	return nil
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
		urlString, ok := params[name]
		urlString = strings.TrimSpace(urlString)
		if !ok || urlString == "" {
			return nil
		}
		u, err := url.Parse(urlString)
		if err != nil || u.Scheme == "" {
			if urlString[0] == '/' {
				u, err = url.Parse("file://" + urlString)
				if err != nil {
					return errors.Wrapf(err, "cannot parse '%s'", urlString)
				}
			} else {
				d, err := os.Getwd()
				if err != nil {
					return errors.Wrap(err, "cannot get working directory")
				}
				u, err = url.Parse("file://" + filepath.ToSlash(filepath.Join(d, urlString)))
				if err != nil {
					return errors.Wrapf(err, "cannot parse '%s'", urlString)
				}
			}
		}
		sl.metadataSource = u
	}
	return nil
}

func (sl *MetaFile) GetName() string { return MetaFileName }

func (sl *MetaFile) WriteConfig(fsys appendfs.FS) error {
	if _, err := writefs.WriteFile(fsys, sl.MetaSchema, sl.schema); err != nil {
		return errors.Wrapf(err, "cannot write schema to %v/%s", fsys, sl.MetaSchema)
	}
	configWriter, err := writefs.Create(fsys, "config.json")
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

func (sl *MetaFile) UpdateObjectBefore(obj object.VersionWriter) error {
	if sl.metadataSource == nil || sl.metadataSource.Path == "" {
		return nil
	}
	if sl.stored {
		return nil
	}
	sl.stored = true
	var err error
	inventory := obj.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	if sl.metadataSource == nil {
		// only a problem, if first version
		if len(util.SeqToSlice(inventory.GetVersions().GetVersionNumbers())) < 2 {
			return errors.New("no metadata source configured")
		}
		return nil
	}
	var rc io.ReadCloser
	var fname string
	switch strings.ToLower(sl.metadataSource.Scheme) {
	case "http":
		fname = strings.Replace(sl.metadataSource.String(), "$ID", obj.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "https":
		fname = strings.Replace(sl.metadataSource.String(), "$ID", obj.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "file":
		fname = strings.Replace(sl.metadataSource.Path, "$ID", obj.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		if windowsPathWithDrive.Match([]byte(fname)) {
			fname = strings.TrimLeft(fname, "/")
		}
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	case "":
		fname = strings.Replace(sl.metadataSource.Path, "$ID", obj.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	default:
		return errors.Errorf("url scheme '%s' not supported", sl.metadataSource.Scheme)
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
		if _, err := obj.AddReader(io.NopCloser(bytes.NewBuffer(infoData)), []string{targetname}, sl.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := obj.GetExtensionManager().GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, sl.StorageName, sl.MetaName)), "/")

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if _, err := obj.AddReader(io.NopCloser(bytes.NewBuffer(infoData)), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join("extensions", sl.GetName(), sl.StorageName, sl.MetaName)), "/")
		if _, err := writefs.WriteFile(obj.GetFS(), targetname, infoData); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", obj.GetFS(), targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", sl.StorageType)
	}

	// remember the content
	sl.info[inventory.GetHead().String()] = infoData
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

func (sl *MetaFile) UpdateObjectAfter(obj object.VersionWriter) error {
	return nil
}

func (sl *MetaFile) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}
	inv := obj.GetInventory()

	// walk through versions and get the latest info.json
	var versions = []*inventory.VersionNumber{}
	for v := range inv.GetVersions().GetVersionNumbers() {
		versions = append(versions, v)
	}
	slices.SortFunc(versions, func(a, b *inventory.VersionNumber) int {
		if a.Equal(b) {
			return 0
		}
		if a.Less(b) {
			return 1
		}
		return -1
	})

	var metadata []byte
	for _, ver := range versions {
		var ok bool
		if metadata, ok = sl.info[ver.String()]; ok {
			break
		}
		if metadata, err = ReadFile(sl.GetName(), sourceFS, obj, sl.MetaName, ver, sl.StorageType, sl.StorageName); err == nil {
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
	_ extensiontypes.Extension     = &MetaFile{}
	_ object.ExtensionObjectChange = &MetaFile{}
	_ object.ExtensionMetadata     = &MetaFile{}
)
