package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const MetadataName = "NNNN-metadata"
const MetadataDescription = "technical metadata for all files"

func GetMetadataParams() []ocfl.ExtensionExternalParam {
	return []ocfl.ExtensionExternalParam{
		{
			Param:       "metadata-source",
			File:        "MetadataSource",
			Description: "url with metadata files. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
	}
}

type MetadataConfig struct {
	*ocfl.ExtensionConfig
	Versioned bool `json:"versioned,omitempty"`
}
type Metadata struct {
	*MetadataConfig
	metadataSource *url.URL
	fs             ocfl.OCFLFS
}

func NewMetadataFS(fs ocfl.OCFLFS, params map[string]string) (*Metadata, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetadataConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewMetadata(config, params)
}
func NewMetadata(config *MetadataConfig, params map[string]string) (*Metadata, error) {
	sl := &Metadata{MetadataConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	if params != nil {
		if urlString, ok := params["metadata-source"]; ok {
			u, err := url.Parse(urlString)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid url '%s'", urlString)
			}
			sl.metadataSource = u
		}
	}
	if sl.metadataSource == nil {
		return nil, errors.Errorf("no metadata-source for extension '%s'", MetadataName)
	}
	return sl, nil
}

func (sl *Metadata) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *Metadata) GetName() string { return MetadataName }
func (sl *Metadata) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := sl.fs.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *Metadata) UpdateBefore(object ocfl.Object) error {
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

func (sl *Metadata) UpdateAfter(object ocfl.Object) error {
	var err error
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	if sl.metadataSource == nil {
		// only a problem, if first version
		if len(inventory.GetVersionStrings()) < 2 {
			return errors.New("no metadata source configured")
		}
		return nil
	}
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	var metadata []byte
	switch sl.metadataSource.Scheme {
	case "http":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		if metadata, err = downloadFile(fname); err != nil {
			return errors.Wrapf(err, "cannot download from '%s'", fname)
		}
	case "https":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		if metadata, err = downloadFile(fname); err != nil {
			return errors.Wrapf(err, "cannot download from '%s'", fname)
		}
	case "file":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		if metadata, err = os.ReadFile(fname); err != nil {
			return errors.Wrapf(err, "cannot ")
		}
	case "":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		if metadata, err = os.ReadFile(fname); err != nil {
			return errors.Wrapf(err, "cannot ")
		}
	default:
		return errors.Errorf("url scheme '%s' not supported", sl.metadataSource.Scheme)
	}
	target := fmt.Sprintf("%s/%s", inventory.GetHead(), filepath.Base(sl.metadataSource.Path))
	w, err := sl.fs.Create(target)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", target)
	}
	defer w.Close()
	if _, err := io.Copy(w, bytes.NewBuffer(metadata)); err != nil {
		return errors.Wrapf(err, "cannot write data to '%s'", target)
	}

	target2 := filepath.Base(sl.metadataSource.Path)
	w2, err := sl.fs.Create(target2)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", target)
	}
	defer w2.Close()
	if _, err := io.Copy(w2, bytes.NewBuffer(metadata)); err != nil {
		return errors.Wrapf(err, "cannot write data to '%s'", target)
	}

	return nil
}
