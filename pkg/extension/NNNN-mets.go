package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/dilcis/mets"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"io"
	"io/fs"
	"net/http"
	"net/url"
)

const METSName = "NNNN-mets"
const METSDescription = "METS/EAD3/PREMIS metadata"

func GetMetsParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: METSName,
			Functions:     []string{"add", "update", "create"},
			Param:         "source",
			//File:          "Source",
			Description: "url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
	}
}

func NewMetsFS(fsys fs.FS) (*Mets, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetsConfig{}
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

	return NewMets(config, schema)
}
func NewMets(config *MetsConfig, schema []byte) (*Mets, error) {
	var err error
	sl := &Mets{
		MetsConfig: config,
		schema:     schema,
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

type MetsConfig struct {
	*ocfl.ExtensionConfig
	StorageType   string `json:"storageType"`
	StorageName   string `json:"storageName"`
	MetaName      string `json:"name,omitempty"`
	MetaSchema    string `json:"schema,omitempty"`
	MetaSchemaUrl string `json:"schemaUrl,omitempty"`
}
type Mets struct {
	*MetsConfig
	schema         []byte
	metadataSource *url.URL
	fsys           fs.FS
	compiledSchema *jsonschema.Schema
	stored         bool
}

func (sl *Mets) GetFS() fs.FS {
	return sl.fsys
}

func (sl *Mets) GetConfig() any {
	return sl.MetsConfig
}

func (sl *Mets) IsRegistered() bool {
	return false
}

func (sl *Mets) SetParams(params map[string]string) error {
	if params != nil {
		name := fmt.Sprintf("ext-%s-%s", METSName, "source")
		if urlString, ok := params[name]; ok {
			u, err := url.Parse(urlString)
			if err != nil {
				return errors.Wrapf(err, "invalid url parameter '%s' - '%s'", name, urlString)
			}
			sl.metadataSource = u
		}
	}
	return nil
}

func (sl *Mets) SetFS(fsys fs.FS) {
	sl.fsys = fsys
}

func (sl *Mets) GetName() string { return METSName }

func (sl *Mets) WriteConfig() error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	if err := writefs.WriteFile(sl.fsys, sl.MetaSchema, sl.schema); err != nil {
		return errors.Wrapf(err, "cannot write schema to %v/%s", sl.fsys, sl.MetaSchema)
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.MetsConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *Mets) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (sl *Mets) UpdateObjectAfter(object ocfl.Object) error {
	metadata, err := object.GetMetadata()
	if err != nil {
		return errors.Wrap(err, "cannot get metadata from object")
	}

	m := &mets.Mets{MetsType: &mets.MetsType{
		XMLName:     xml.Name{},
		IDAttr:      "",
		OBJIDAttr:   "",
		LABELAttr:   "",
		TYPEAttr:    "",
		PROFILEAttr: "",
		MetsHdr:     nil,
		DmdSec:      nil,
		AmdSec:      nil,
		FileSec:     nil,
		StructMap:   nil,
		StructLink:  nil,
		BehaviorSec: nil,
	}}

	_ = m
	_ = metadata
	return nil
}

// check interface satisfaction
var (
	_ ocfl.ExtensionObjectChange = &Mets{}
)
